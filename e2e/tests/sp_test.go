package tests

import (
	"context"
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

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}

// todo(Chris Li): Fix this test
func (s *StorageProviderTestSuite) CreateStorageProvider() {
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

	// 2. grant deposit authorization of sp to gov module account
	sp := s.StorageProviders[0]
	coins := sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB))
	authorization, err := sptypes.NewDepositAuthorization(sp.OperatorKey.GetAddr(), &coins)
	s.Require().NoError(err)

	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	now := time.Now().Add(24 * time.Hour)
	grantMsg, err := authz.NewMsgGrant(
		sp.OperatorKey.GetAddr(), govAddr, authorization, &now)
	s.Require().NoError(err)
	s.SendTxBlock(grantMsg, sp.OperatorKey)

	// 3. submit CreateStorageProvider proposal
	msgCreateSP, _ := sptypes.NewMsgCreateStorageProvider(govAddr,
		sp.OperatorKey.GetAddr(), sp.FundingKey.GetAddr(),
		sp.SealKey.GetAddr(),
		sp.ApprovalKey.GetAddr(), description,
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
	if proposalRes.Proposal.Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
		s.Require().True(false)
	}

	// 7. query storage provider
	querySPReq := sptypes.QueryStorageProviderRequest{
		SpAddress: sp.OperatorKey.GetAddr().String(),
	}
	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider.OperatorAddress, sp.OperatorKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.FundingAddress, sp.FundingKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.SealAddress, sp.SealKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.ApprovalAddress, sp.ApprovalKey.GetAddr().String())
	s.Require().Equal(querySPResp.StorageProvider.Endpoint, "sp0.greenfield.io")
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
		ExpireTime:    time.Now().Unix() + 86400,
		ReadPrice:     newReadPrice,
		StorePrice:    newStorePrice,
		FreeReadQuota: spStoragePrice.SpStoragePrice.FreeReadQuota,
	}
	_ = s.SendTxBlock(msgUpdateSpStoragePrice, sp.OperatorKey)
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
}

func (s *StorageProviderTestSuite) CheckSecondarySpPrice() {
	ctx := context.Background()
	queryGetSecondarySpStorePriceByTimeResp, err := s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log(queryGetSecondarySpStorePriceByTimeResp)
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
		s.T().Logf("sp: %s, storage price: %s", sp.OperatorAddress, spStoragePrice)
		total = total.Add(spStoragePrice.SpStoragePrice.StorePrice)
	}
	expectedSecondarySpStorePrice := sptypes.SecondarySpStorePriceRatio.Mul(total).QuoInt64(spNum)
	s.Require().Equal(expectedSecondarySpStorePrice, queryGetSecondarySpStorePriceByTimeResp.SecondarySpStorePrice.StorePrice)
}
