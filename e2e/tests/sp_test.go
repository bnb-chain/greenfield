package tests

import (
	"context"
	"github.com/bnb-chain/greenfield/testutil/keeper"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

type StorageProviderTestSuite struct {
	core.BaseSuite
}

func (s *StorageProviderTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StorageProviderTestSuite) SetupTest() {
}

func (s *StorageProviderTestSuite) NewSpAcc() *core.SPKeyManagers {
	userAccs := s.GenAndChargeAccounts(4, 1000000)
	operatorAcc := userAccs[0]
	fundingAcc := userAccs[1]
	approvalAcc := userAccs[2]
	sealAcc := userAccs[3]

	return &core.SPKeyManagers{OperatorKey: operatorAcc, SealKey: fundingAcc, FundingKey: approvalAcc, ApprovalKey: sealAcc}
}

func (s *StorageProviderTestSuite) NewSpAccAndGrant() *core.SPKeyManagers {
	// 1. create new newStorageProvider
	newSP := keeper.NewSpAcc()

	// 2. grant deposit authorization of sp to gov module account
	coins := sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB))
	authorization, err := sptypes.NewDepositAuthorization(newSP.OperatorKey.GetAddr(), &coins)
	s.Require().NoError(err)

	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	s.T().Logf("acc %s", govAddr)
	now := time.Now().Add(24 * time.Hour)
	grantMsg, err := authz.NewMsgGrant(
		newSP.FundingKey.GetAddr(), govAddr, authorization, &now)
	s.Require().NoError(err)
	s.SendTxBlock(grantMsg, newSP.FundingKey)

	return newSP
}

func (s *StorageProviderTestSuite) TestCreateStorageProvider() {
	ctx := context.Background()
	validator := s.Validator.GetAddr()

	// 1. submit CreateStorageProviderParams
	deposit := sdk.Coin{
		Denom:  s.Config.Denom,
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}
	description := sptypes.Description{
		Moniker:  "sp_test",
		Identity: "",
	}

	// create new newStorageProvider and grant
	newSP := s.NewSpAccAndGrant()

	// 2. grant deposit authorization of sp to gov module account
	coins := sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB))
	authorization, err := sptypes.NewDepositAuthorization(newSP.OperatorKey.GetAddr(), &coins)
	s.Require().NoError(err)

	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	s.T().Logf("acc %s", govAddr)
	now := time.Now().Add(24 * time.Hour)
	grantMsg, err := authz.NewMsgGrant(
		newSP.FundingKey.GetAddr(), govAddr, authorization, &now)
	s.Require().NoError(err)
	s.SendTxBlock(grantMsg, newSP.FundingKey)

	// 3. submit CreateStorageProvider proposal
	msgCreateSP, _ := sptypes.NewMsgCreateStorageProvider(govAddr,
		newSP.OperatorKey.GetAddr(), newSP.FundingKey.GetAddr(),
		newSP.SealKey.GetAddr(),
		newSP.ApprovalKey.GetAddr(), description,
		"sp0.greenfield.io", deposit)
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgCreateSP},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
		validator.String(),
		"test",
	)
	s.Require().NoError(err)

	txRes := s.SendTxBlock(msgProposal, s.Validator)
	s.Require().Equal(txRes.Code, uint32(0))

	// 4. query proposal and get proposal ID
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

	// 5. submit MsgVote and wait the proposal exec
	msgVote := govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	txRes = s.SendTxBlock(msgVote, s.Validator)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := govtypesv1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(ctx, &queryVoteParamsReq)
	s.Require().NoError(err)

	// 6. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.VotingParams.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.VotingParams.VotingPeriod)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	if proposalRes.Proposal.Status != govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
		s.Require().True(false)
	}

	// 7. query storage provider
	querySPReq := sptypes.QueryStorageProviderRequest{
		SpAddress: newSP.OperatorKey.GetAddr().String(),
	}
	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider.OperatorAddress, newSP.OperatorKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.FundingAddress, newSP.FundingKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.SealAddress, newSP.SealKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.ApprovalAddress, newSP.ApprovalKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.Endpoint, "sp0.greenfield.io")
}

func (s *StorageProviderTestSuite) TestEditStorageProvider() {
	ctx := context.Background()

	// 1. query previous storage provider
	querySPReq := sptypes.QueryStorageProviderRequest{
		SpAddress: s.StorageProvider.OperatorKey.GetAddr().String(),
	}

	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider.Endpoint, "sp0.greenfield.io")

	// 2. edit storage provider
	endpoint := "127.0.0.1:9033"
	description := sptypes.Description{
		Moniker:  "sp_test_edit",
		Identity: "",
	}

	msgEditSP := sptypes.NewMsgEditStorageProvider(
		s.StorageProvider.OperatorKey.GetAddr(), endpoint, description)
	txRes := s.SendTxBlock(msgEditSP, s.StorageProvider.OperatorKey)
	s.Require().Equal(txRes.Code, uint32(0))

	// 3. query modifyed storage provider
	querySPReq = sptypes.QueryStorageProviderRequest{
		SpAddress: s.StorageProvider.OperatorKey.GetAddr().String(),
	}

	querySPResp, err = s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider.Endpoint, endpoint)
	s.Require().Equal(querySPResp.StorageProvider.Description.Moniker, "sp_test_edit")
}

func (s *StorageProviderTestSuite) TestMsgCreateStorageProvider() {

}

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}
