package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
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
	userAccs := s.GenAndChargeAccounts(5, 1000000)
	operatorAcc := userAccs[0]
	fundingAcc := userAccs[1]
	approvalAcc := userAccs[2]
	sealAcc := userAccs[3]
	gcAcc := userAccs[4]

	return &core.SPKeyManagers{OperatorKey: operatorAcc, SealKey: fundingAcc,
		FundingKey: approvalAcc, ApprovalKey: sealAcc, GcKey: gcAcc}
}

func (s *StorageProviderTestSuite) NewSpAccAndGrant() *core.SPKeyManagers {
	// 1. create new newStorageProvider
	newSP := s.NewSpAcc()

	// 2. grant deposit authorization of sp to gov module account
	coins := sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB))
	authorization := sptypes.NewDepositAuthorization(newSP.OperatorKey.GetAddr(), &coins)

	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	now := time.Now().Add(24 * time.Hour)
	grantMsg, err := authz.NewMsgGrant(
		newSP.FundingKey.GetAddr(), govAddr, authorization, &now)
	s.Require().NoError(err)
	s.SendTxBlock(newSP.FundingKey, grantMsg)

	return newSP
}

func (s *StorageProviderTestSuite) TestCreateStorageProvider() {
	ctx := context.Background()
	validator := s.Validator.GetAddr()

	// 1. create new newStorageProvider and grant
	newSP := s.NewSpAccAndGrant()

	// 2. submit CreateStorageProvider proposal
	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	deposit := sdk.Coin{
		Denom:  s.Config.Denom,
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}
	description := sptypes.Description{
		Moniker:  "sp_test",
		Identity: "",
	}

	endpoint := "http://127.0.0.1:9034"
	newReadPrice := sdk.NewDec(core.RandInt64(100, 200))
	newStorePrice := sdk.NewDec(core.RandInt64(10000, 20000))
	msgCreateSP, _ := sptypes.NewMsgCreateStorageProvider(govAddr,
		newSP.OperatorKey.GetAddr(), newSP.FundingKey.GetAddr(),
		newSP.SealKey.GetAddr(),
		newSP.ApprovalKey.GetAddr(),
		newSP.GcKey.GetAddr(), description,
		endpoint, deposit, newReadPrice, 10000, newStorePrice)
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgCreateSP},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
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
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(proposalRes.Proposal.Status, govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED)

	// 6. query storage provider
	querySPReq := sptypes.QueryStorageProviderRequest{
		SpAddress: newSP.OperatorKey.GetAddr().String(),
	}
	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider.OperatorAddress, newSP.OperatorKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.FundingAddress, newSP.FundingKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.SealAddress, newSP.SealKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.ApprovalAddress, newSP.ApprovalKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.Endpoint, endpoint)
}

func (s *StorageProviderTestSuite) TestEditStorageProvider() {
	ctx := context.Background()
	sp := s.StorageProviders[0]

	// 1. query previous storage provider
	querySPReq := sptypes.QueryStorageProviderRequest{
		SpAddress: sp.OperatorKey.GetAddr().String(),
	}

	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	prevSP := querySPResp.StorageProvider

	// 2. edit storage provider
	newSP := &sptypes.StorageProvider{
		OperatorAddress: prevSP.OperatorAddress,
		FundingAddress:  prevSP.FundingAddress,
		SealAddress:     prevSP.SealAddress,
		ApprovalAddress: prevSP.ApprovalAddress,
		GcAddress:       prevSP.GcAddress,
		Description: sptypes.Description{
			Moniker:  "sp_test_edit",
			Identity: "",
		},
		Endpoint:     "http://127.0.0.1:9034",
		TotalDeposit: prevSP.TotalDeposit,
	}

	msgEditSP := sptypes.NewMsgEditStorageProvider(
		sp.OperatorKey.GetAddr(), newSP.Endpoint, &newSP.Description,
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr())
	txRes := s.SendTxBlock(sp.OperatorKey, msgEditSP)
	s.Require().Equal(txRes.Code, uint32(0))

	// 3. query modifyed storage provider
	querySPReq = sptypes.QueryStorageProviderRequest{
		SpAddress: sp.OperatorKey.GetAddr().String(),
	}

	querySPResp, err = s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, newSP)

	// 4. revert storage provider info
	msgEditSP = sptypes.NewMsgEditStorageProvider(
		sp.OperatorKey.GetAddr(), prevSP.Endpoint, &prevSP.Description,
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr())
	txRes = s.SendTxBlock(sp.OperatorKey, msgEditSP)
	s.Require().Equal(txRes.Code, uint32(0))

	// 5. query revert storage provider again
	querySPReq = sptypes.QueryStorageProviderRequest{
		SpAddress: sp.OperatorKey.GetAddr().String(),
	}

	querySPResp, err = s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, prevSP)
}

func (s *StorageProviderTestSuite) TestDeposit() {
	sp := s.StorageProviders[0]

	deposit := sdk.Coin{
		Denom:  s.Config.Denom,
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}

	msgDeposit := sptypes.NewMsgDeposit(
		sp.FundingKey.GetAddr(), sp.OperatorKey.GetAddr(), deposit)
	txRes := s.SendTxBlock(sp.FundingKey, msgDeposit)
	s.Require().Equal(txRes.Code, uint32(0))
}

func (s *StorageProviderTestSuite) TestSpStoragePrice() {
	ctx := context.Background()
	s.CheckSecondarySpPrice()
	sp := s.StorageProviders[0]
	spAddr := sp.OperatorKey.GetAddr().String()
	spStoragePrice, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAddr,
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log(spStoragePrice)
	// update storage price
	newReadPrice := sdk.NewDec(core.RandInt64(100, 200))
	newStorePrice := sdk.NewDec(core.RandInt64(10000, 20000))
	msgUpdateSpStoragePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     spAddr,
		ReadPrice:     newReadPrice,
		StorePrice:    newStorePrice,
		FreeReadQuota: spStoragePrice.SpStoragePrice.FreeReadQuota,
	}
	_ = s.SendTxBlock(sp.OperatorKey, msgUpdateSpStoragePrice)
	// query and assert
	spStoragePrice2, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAddr,
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log(spStoragePrice2)
	// check price changed as expected
	s.Require().Equal(newReadPrice, spStoragePrice2.SpStoragePrice.ReadPrice)
	s.Require().Equal(newStorePrice, spStoragePrice2.SpStoragePrice.StorePrice)
	s.CheckSecondarySpPrice()
	// query sp storage price by time before it exists, expect error
	_, err = s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: 1,
	})
	s.Require().Error(err)
	_, err = s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAddr,
		Timestamp: 1,
	})
	s.Require().Error(err)
}

func (s *StorageProviderTestSuite) CheckSecondarySpPrice() {
	ctx := context.Background()
	queryGetSecondarySpStorePriceByTimeResp, err := s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Logf("Secondary SP store price: %s", core.YamlString(queryGetSecondarySpStorePriceByTimeResp.SecondarySpStorePrice))
	// query all sps
	sps, err := s.Client.StorageProviders(ctx, &sptypes.QueryStorageProvidersRequest{})
	s.Require().NoError(err)
	s.T().Logf("sps: %s", sps)
	spNum := int64(sps.Pagination.Total)
	total := sdk.ZeroDec()
	for _, sp := range sps.Sps {
		spStoragePrice, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
			SpAddr:    sp.OperatorAddress,
			Timestamp: 0,
		})
		s.Require().NoError(err)
		s.T().Logf("sp: %s, storage price: %s", sp.OperatorAddress, core.YamlString(spStoragePrice.SpStoragePrice))
		total = total.Add(spStoragePrice.SpStoragePrice.StorePrice)
	}
	params, err := s.Client.SpQueryClient.Params(ctx, &sptypes.QueryParamsRequest{})
	s.Require().NoError(err)
	expectedSecondarySpStorePrice := params.Params.SecondarySpStorePriceRatio.Mul(total).QuoInt64(spNum)
	s.Require().Equal(expectedSecondarySpStorePrice, queryGetSecondarySpStorePriceByTimeResp.SecondarySpStorePrice.StorePrice)
}

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}
