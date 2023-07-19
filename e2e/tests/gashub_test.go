package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
)

type GashubTestSuite struct {
	core.BaseSuite
}

func (s *GashubTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *GashubTestSuite) SetupTest() {}

func (s *GashubTestSuite) TestUpdateParams() {
	ctx := context.Background()
	validator := s.Validator.GetAddr()

	// 1. submit MsgUpdateMsgGasParams
	typeUrl := sdk.MsgTypeURL(&banktypes.MsgSend{})
	msgSendGasParams := gashubtypes.NewMsgGasParamsWithFixedGas(typeUrl, 1e6)
	msgUpdateGasParams := gashubtypes.NewMsgSetMsgGasParams(authtypes.NewModuleAddress(gov.ModuleName).String(), []*gashubtypes.MsgGasParams{msgSendGasParams}, nil)
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdateGasParams},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
		validator.String(),
		"test",
		"update gas params",
		"pdate gas params",
	)
	s.Require().NoError(err)

	txRes := s.SendTxBlock(s.Validator, msgProposal)
	s.Require().Equal(txRes.Code, uint32(0))

	// 2. query proposal
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

	// 3. submit MsgVote and wait the proposal exec
	msgVote := govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	txRes = s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	for {
		proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
		s.Require().NoError(err)
		if proposalRes.Proposal.Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		}
		s.T().Logf("waiting for proposal to be passed, now: %s", time.Now())
		time.Sleep(time.Second)
	}

	// 4. query new gas params
	queryRequest := &gashubtypes.QueryMsgGasParamsRequest{}
	queryRes, err := s.Client.GashubQueryClient.MsgGasParams(ctx, queryRequest)
	s.Require().NoError(err)

	for _, params := range queryRes.GetMsgGasParams() {
		if params.MsgTypeUrl == typeUrl {
			s.Require().True(params.GetFixedType().Equal(msgSendGasParams.GetFixedType()))
		}
	}
}

func (s *VirtualGroupTestSuite) TestUpdateGasHubParams() {
	// 1. create proposal
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	queryParamsResp, err := s.Client.GashubQueryClient.Params(context.Background(), &gashubtypes.QueryParamsRequest{})
	s.Require().NoError(err)
	msgUpdateParams := &gashubtypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    queryParamsResp.Params,
	}

	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgUpdateParams}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "update GasHub params", "Test update GasHub params")
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
	txOpt := &types.TxOption{
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
				return
			case v1.StatusFailed:
				s.T().Errorf("proposal failed, reason %s", queryProposalResp.Proposal.FailedReason)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func TestGashubTestSuite(t *testing.T) {
	suite.Run(t, new(GashubTestSuite))
}
