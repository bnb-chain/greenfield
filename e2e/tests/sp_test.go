package tests

import (
	"context"
	sdkmath "cosmossdk.io/math"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
)

type SpTestSuite struct {
	core.BaseSuite
}

func (s *SpTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *SpTestSuite) SetupTest() {}

func (s *SpTestSuite) TestSpStoragePrice() {
	ctx := context.Background()
	s.CheckSecondarySpPrice()
	spAddr := s.StorageProvider.OperatorKey.GetAddr().String()
	spStoragePrice, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAddr,
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log(spStoragePrice)
	// update storage price
	newReadPrice := sdkmath.NewInt(randInt64(100, 200))
	newStorePrice := sdkmath.NewInt(randInt64(10000, 20000))
	msgUpdateSpStoragePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     spAddr,
		ExpireTime:    time.Now().Unix() + 86400,
		ReadPrice:     newReadPrice,
		StorePrice:    newStorePrice,
		FreeReadQuota: spStoragePrice.SpStoragePrice.FreeReadQuota,
	}
	_ = s.SendTxBlock(msgUpdateSpStoragePrice, s.StorageProvider.OperatorKey)
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

func (s *SpTestSuite) CheckSecondarySpPrice() {
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
	spNum := sdkmath.NewIntFromUint64(sps.Pagination.Total)
	total := sdkmath.NewInt(0)
	for _, sp := range sps.Sps {
		spStoragePrice, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
			SpAddr:    sp.OperatorAddress,
			Timestamp: 0,
		})
		s.Require().NoError(err)
		s.T().Logf("sp: %s, storage price: %s", sp.OperatorAddress, spStoragePrice)
		total = total.Add(spStoragePrice.SpStoragePrice.StorePrice)
	}
	expectedSecondarySpStorePrice := total.Quo(spNum).MulRaw(sptypes.SecondarySpStorePriceRatio).QuoRaw(sptypes.RatioUnit)
	s.Require().Equal(expectedSecondarySpStorePrice, queryGetSecondarySpStorePriceByTimeResp.SecondarySpStorePrice.StorePrice)
}

func TestSpTestSuite(t *testing.T) {
	suite.Run(t, new(SpTestSuite))
}

// generate random int64 between min and max
func randInt64(min, max int64) int64 {
	return min + rand.Int63n(max-min)
}
