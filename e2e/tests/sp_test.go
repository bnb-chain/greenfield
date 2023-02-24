package tests

import (
	"context"
	"fmt"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"strconv"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"

	authz "github.com/cosmos/cosmos-sdk/x/authz"
)

type StorageProviderTestSuite struct {
	core.BaseSuite
}

func (s *StorageProviderTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StorageProviderTestSuite) SetupTest() {
}

func (s *StorageProviderTestSuite) TestCreateStorageProvider() {
	//user := s.GenAndChargeAccounts(1, 1000000)[0]

	ctx := context.Background()
	validator := s.Validator.GetAddr()

	// 1. submit CreateStorageProviderParams
	deposit := sdk.Coin{
		Denom:  "bnb",
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}
	description := sptypes.Description{
		Moniker:  "sp0",
		Identity: "",
	}

	// grant
	coins := sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB))
	authorization, err := sptypes.NewDepositAuthorization(s.StorageProvider.OperatorKey.GetAddr(), &coins)
	s.Require().NoError(err)

	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	now := time.Now().Add(24 * time.Hour)
	grantMsg, err := authz.NewMsgGrant(
		s.StorageProvider.OperatorKey.GetAddr(), govAddr, authorization, &now)
	s.SendTxBlock(grantMsg, s.StorageProvider.OperatorKey)

	msgCreateSP, _ := sptypes.NewMsgCreateStorageProvider(govAddr,
		s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.FundingKey.GetAddr(),
		s.StorageProvider.SealKey.GetAddr(),
		s.StorageProvider.ApprovalKey.GetAddr(), description,
		"sp0.greenfield.io", deposit)
	//s.StorageProvider.OperatorKey.GetAddr().String(),
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgCreateSP},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
		validator.String(),
		"test",
	)
	s.Require().NoError(err)

	txRes := s.SendTxBlock(msgProposal, s.Validator)
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
	txRes = s.SendTxBlock(msgVote, s.Validator)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := govtypesv1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(ctx, &queryVoteParamsReq)

	s.T().Logf("voting period %s", *queryVoteParamsResp.VotingParams.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.VotingParams.VotingPeriod)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	if proposalRes.Proposal.Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
		s.Require().True(false)
	}

	// 4. query new gas params
	queryRequest := &sptypes.QueryParamsRequest{}
	queryRes, err := s.Client.SpQueryClient.Params(ctx, queryRequest)
	s.Require().NoError(err)

	fmt.Println(queryRes)
}

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}
