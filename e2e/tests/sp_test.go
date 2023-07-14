package tests

import (
	"context"
	"encoding/hex"
	"sort"
	"testing"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/testutil/sample"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
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
	// Create a New SP
	sp := s.BaseSuite.CreateNewStorageProvider()

	// query sp by id
	querySPResp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{
		Id: sp.Info.Id,
	})
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, querySPResp.StorageProvider)

	// sp exit
	msgSPExit := virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: sp.OperatorKey.GetAddr().String(),
	}
	s.SendTxBlock(sp.OperatorKey, &msgSPExit)

	// 9 query sp status
	querySPResp2, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(querySPResp2.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 10 complete sp exit
	msgCompleteSPExit := virtualgroupmoduletypes.MsgCompleteStorageProviderExit{
		StorageProvider: sp.OperatorKey.GetAddr().String(),
	}

	s.SendTxBlock(sp.OperatorKey, &msgCompleteSPExit)

	// 10 query sp
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
	s.Require().Error(err)
}

func (s *StorageProviderTestSuite) TestEditStorageProvider() {
	ctx := context.Background()
	sp := s.BaseSuite.PickStorageProvider()
	blsProof, _ := sp.BlsKey.Sign(tmhash.Sum(sp.BlsKey.PubKey().Bytes()))

	// 1. query previous storage provider
	querySPByOperatorAddressReq := sptypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: sp.OperatorKey.GetAddr().String(),
	}

	querySPByOperatorAddressResp, err := s.Client.StorageProviderByOperatorAddress(ctx, &querySPByOperatorAddressReq)
	s.Require().NoError(err)
	prevSP := querySPByOperatorAddressResp.StorageProvider

	// 2. edit storage provider
	newBlsPubKeyBz, newBlsProofBz := sample.RandBlsPubKeyAndBlsProofBz()

	newSP := &sptypes.StorageProvider{
		OperatorAddress: prevSP.OperatorAddress,
		FundingAddress:  prevSP.FundingAddress,
		SealAddress:     prevSP.SealAddress,
		ApprovalAddress: prevSP.ApprovalAddress,
		GcAddress:       prevSP.GcAddress,
		BlsKey:          newBlsPubKeyBz,
		Description: sptypes.Description{
			Moniker:  "sp_test_edit",
			Identity: "",
		},
		Endpoint:     "http://127.0.0.1:9034",
		TotalDeposit: prevSP.TotalDeposit,
	}
	msgEditSP := sptypes.NewMsgEditStorageProvider(
		sp.OperatorKey.GetAddr(), newSP.Endpoint, &newSP.Description,
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr(),
		hex.EncodeToString(newBlsPubKeyBz),
		hex.EncodeToString(newBlsProofBz),
	)

	txRes := s.SendTxBlock(sp.OperatorKey, msgEditSP)
	s.Require().Equal(txRes.Code, uint32(0))

	// 3. query modified storage provider
	querySPReq := sptypes.QueryStorageProviderRequest{
		Id: sp.Info.Id,
	}

	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	newSP.Id = querySPResp.StorageProvider.Id
	s.Require().Equal(querySPResp.StorageProvider, newSP)

	// 4. revert storage provider info
	msgEditSP = sptypes.NewMsgEditStorageProvider(
		sp.OperatorKey.GetAddr(), prevSP.Endpoint, &prevSP.Description,
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr(),
		hex.EncodeToString(prevSP.BlsKey), hex.EncodeToString(blsProof))

	txRes = s.SendTxBlock(sp.OperatorKey, msgEditSP)
	s.Require().Equal(txRes.Code, uint32(0))

	// 5. query revert storage provider again
	querySPReq = sptypes.QueryStorageProviderRequest{
		Id: sp.Info.Id,
	}

	querySPResp, err = s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, prevSP)
}

func (s *StorageProviderTestSuite) TestDeposit() {
	sp := s.BaseSuite.PickStorageProvider()

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
	sp := s.BaseSuite.PickStorageProvider()
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
	prices := make([]sdk.Dec, 0)
	for _, sp := range sps.Sps {
		spStoragePrice, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
			SpAddr:    sp.OperatorAddress,
			Timestamp: 0,
		})
		s.Require().NoError(err)
		s.T().Logf("sp: %s, storage price: %s", sp.OperatorAddress, core.YamlString(spStoragePrice.SpStoragePrice))
		prices = append(prices, spStoragePrice.SpStoragePrice.StorePrice)
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i].LT(prices[j]) })
	var median sdk.Dec
	if spNum%2 == 0 {
		median = prices[spNum/2-1].Add(prices[spNum/2]).QuoInt64(2)
	} else {
		median = prices[spNum/2]
	}

	params, err := s.Client.SpQueryClient.Params(ctx, &sptypes.QueryParamsRequest{})
	s.Require().NoError(err)
	expectedSecondarySpStorePrice := params.Params.SecondarySpStorePriceRatio.Mul(median)
	s.Require().Equal(expectedSecondarySpStorePrice, queryGetSecondarySpStorePriceByTimeResp.SecondarySpStorePrice.StorePrice)
}

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}
