package tests

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	gnfdtypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	_ "github.com/ghodss/yaml"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
)

type BridgeTestSuite struct {
	core.BaseSuite
}

func (s *BridgeTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *BridgeTestSuite) SetupTest() {}

func (s *BridgeTestSuite) TestTransferOut() {
	users := s.GenAndChargeAccounts(2, 1000000)

	from, to := users[0], users[1]
	ctx := context.Background()

	// transfer out token
	transferAmount := sdkmath.NewInt(10000)
	msgTransferOut := &bridgetypes.MsgTransferOut{
		From:   from.GetAddr().String(),
		To:     to.GetAddr().String(),
		Amount: &types.Coin{Denom: types2.Denom, Amount: transferAmount},
	}

	params, err := s.Client.BridgeQueryClient.Params(ctx, &bridgetypes.QueryParamsRequest{})
	s.Require().NoError(err)

	totalTransferOutRelayerFee := params.Params.BscTransferOutRelayerFee.Add(params.Params.BscTransferOutAckRelayerFee)

	moduleAccount := types.MustAccAddressFromHex("0xB73C0Aac4C1E606C6E495d848196355e6CB30381")
	// query balance before
	moduleBalanceBefore, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: moduleAccount.String(),
		Denom:   s.Config.Denom,
	})

	s.Require().NoError(err)

	s.T().Logf("balance before: %s %s", from.GetAddr().String(), moduleBalanceBefore.Balance.String())

	txRes := s.SendTxBlock(from, msgTransferOut)
	s.T().Log(txRes.RawLog)

	moduleBalanceAfter, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: moduleAccount.String(),
		Denom:   s.Config.Denom,
	})
	s.T().Logf("balance after: %s %s", from.GetAddr().String(), moduleBalanceAfter.Balance.String())
	s.Require().NoError(err)

	s.Require().Equal(moduleBalanceBefore.Balance.Amount.Add(transferAmount).Add(totalTransferOutRelayerFee).String(), moduleBalanceAfter.Balance.Amount.String())
}

func (s *BridgeTestSuite) TestGovChannel() {
	var err error
	validator := s.Validator.GetAddr()

	ctx := context.Background()

	msgUpdatePermissions := &crosschaintypes.MsgUpdateChannelPermissions{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		ChannelPermissions: []*crosschaintypes.ChannelPermission{
			{
				DestChainId: 714,
				ChannelId:   uint32(bridgetypes.TransferOutChannelID),
				Permission:  uint32(sdk.ChannelForbidden),
			},
		},
	}

	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdatePermissions},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types2.NewIntFromInt64WithDecimal(100, types2.DecimalBNB))},
		validator.String(),
		"test", "test", "test",
	)
	s.Require().NoError(err)

	txRes := s.SendTxBlock(s.Validator, msgProposal)
	s.Require().Equal(txRes.Code, uint32(0))

	// 3. query proposal and get proposal ID
	var proposalId uint64
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if attr.Key == "proposal_id" {
					proposalId, err = strconv.ParseUint(attr.Value, 10, 0)
					s.Require().NoError(err)
					break
				}
			}
			break
		}
	}
	s.Require().True(proposalId != 0)

	queryProposal := &govtypesv1.QueryProposalRequest{ProposalId: proposalId}
	_, err = s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)

	// 4. submit MsgVote and wait the proposal exec
	msgVote := govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	txRes = s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := govtypesv1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(ctx, &queryVoteParamsReq)
	s.Require().NoError(err)

	// 5. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(1 * time.Second)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(proposalRes.Proposal.Status, govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED)

	users := s.GenAndChargeAccounts(2, 1000000)

	from, to := users[0], users[1]

	// transfer out token
	transferAmount := sdkmath.NewInt(10000)
	msgTransferOut := &bridgetypes.MsgTransferOut{
		From:   from.GetAddr().String(),
		To:     to.GetAddr().String(),
		Amount: &types.Coin{Denom: types2.Denom, Amount: transferAmount},
	}

	s.Require().NoError(err)

	s.SendTxBlockWithExpectErrorString(msgTransferOut, from, "not allowed to write syn package")

	msgUpdatePermissions = &crosschaintypes.MsgUpdateChannelPermissions{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		ChannelPermissions: []*crosschaintypes.ChannelPermission{
			{
				DestChainId: 714,
				ChannelId:   uint32(bridgetypes.TransferOutChannelID),
				Permission:  uint32(sdk.ChannelAllow),
			},
		},
	}

	msgProposal, err = govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdatePermissions},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types2.NewIntFromInt64WithDecimal(100, types2.DecimalBNB))},
		validator.String(),
		"test", "test", "test",
	)
	s.Require().NoError(err)

	txRes = s.SendTxBlock(s.Validator, msgProposal)
	s.Require().Equal(txRes.Code, uint32(0))

	// 3. query proposal and get proposal ID
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if attr.Key == "proposal_id" {
					proposalId, err = strconv.ParseUint(attr.Value, 10, 0)
					s.Require().NoError(err)
					break
				}
			}
			break
		}
	}
	s.Require().True(proposalId != 0)

	queryProposal = &govtypesv1.QueryProposalRequest{ProposalId: proposalId}
	_, err = s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)

	// 4. submit MsgVote and wait the proposal exec
	msgVote = govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	txRes = s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq = govtypesv1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err = s.Client.GovQueryClientV1.Params(ctx, &queryVoteParamsReq)
	s.Require().NoError(err)

	// 5. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(1 * time.Second)
	proposalRes, err = s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(proposalRes.Proposal.Status, govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
}

func (s *BridgeTestSuite) TestUpdateBridgeParams() {
	// 1. create proposal
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	queryParamsResp, err := s.Client.BridgeQueryClient.Params(context.Background(), &bridgetypes.QueryParamsRequest{})
	s.Require().NoError(err)

	updatedParams := queryParamsResp.Params
	updatedParams.BscTransferOutRelayerFee = sdkmath.NewInt(250000000000000)
	msgUpdateParams := &bridgetypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    updatedParams,
	}

	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgUpdateParams}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "update Bridge params", "Test update Bridge params")
	s.Require().NoError(err)
	txBroadCastResp, err := s.SendTxBlockWithoutCheck(proposal, s.Validator)
	s.Require().NoError(err)
	s.T().Log("create proposal tx hash: ", txBroadCastResp.TxResponse.TxHash)

	// get proposal id
	proposalID := 0
	txResp, err := s.WaitForTx(txBroadCastResp.TxResponse.TxHash)
	s.Require().NoError(err)
	if txResp.Code == 0 && txResp.Height > 0 {
		for _, event := range txResp.Events {
			if event.Type == "submit_proposal" {
				proposalID, err = strconv.Atoi(event.GetAttributes()[0].Value)
				s.Require().NoError(err)
			}
		}
	}

	// 2. vote
	if proposalID == 0 {
		s.T().Errorf("proposalID is 0")
		return
	}
	s.T().Log("proposalID: ", proposalID)
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &gnfdtypes.TxOption{
		Mode:      &mode,
		Memo:      "",
		FeeAmount: sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
	}
	voteBroadCastResp, err := s.SendTxBlockWithoutCheckWithTxOpt(v1.NewMsgVote(s.Validator.GetAddr(), uint64(proposalID), v1.OptionYes, ""),
		s.Validator, txOpt)
	s.Require().NoError(err)
	voteResp, err := s.WaitForTx(voteBroadCastResp.TxResponse.TxHash)
	s.Require().NoError(err)
	s.T().Log("vote tx hash: ", voteResp.TxHash)
	if voteResp.Code > 0 {
		s.T().Errorf("voteTxResp.Code > 0")
		return
	}

	// 3. query proposal until it is end voting period
CheckProposalStatus:
	for {
		queryProposalResp, err := s.Client.Proposal(context.Background(), &v1.QueryProposalRequest{ProposalId: uint64(proposalID)})
		s.Require().NoError(err)
		if queryProposalResp.Proposal.Status != v1.StatusVotingPeriod {
			switch queryProposalResp.Proposal.Status {
			case v1.StatusDepositPeriod:
				s.T().Errorf("proposal deposit period")
				return
			case v1.StatusRejected:
				s.T().Errorf("proposal rejected")
				return
			case v1.StatusPassed:
				s.T().Logf("proposal passed")
				break CheckProposalStatus
			case v1.StatusFailed:
				s.T().Errorf("proposal failed, reason %s", queryProposalResp.Proposal.FailedReason)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}

	// 4. check params updated
	err = s.WaitForNextBlock()
	s.Require().NoError(err)

	updatedQueryParamsResp, err := s.Client.BridgeQueryClient.Params(context.Background(), &bridgetypes.QueryParamsRequest{})
	s.Require().NoError(err)
	if reflect.DeepEqual(updatedQueryParamsResp.Params, updatedParams) {
		s.T().Logf("update params success")
	} else {
		s.T().Errorf("update params failed")
	}
}

func TestBridgeTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeTestSuite))
}
