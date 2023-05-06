package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/sample"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type IntegrationTestSuiteWithoutMock struct {
	suite.Suite

	keeper               *keeper.Keeper
	depKeepers           keepertest.StorageDepKeepers
	ctx                  sdk.Context
	PrimarySpAddr        sdk.AccAddress
	PrimarySpFundingAddr sdk.AccAddress
	PrimarySp            sptypes.StorageProvider
	SecondarySps         []sptypes.StorageProvider
	UserAddr             sdk.AccAddress
	Denom                string
}

func (s *IntegrationTestSuiteWithoutMock) SetupTest() {
	s.Denom = "BNB"
	s.keeper, s.depKeepers, s.ctx = keepertest.StorageKeeper(s.T())
	ctx := s.ctx.WithBlockTime(time.Now())
	// init data
	s.PrimarySpAddr = sample.RandAccAddress()
	s.PrimarySpFundingAddr = sample.RandAccAddress()
	s.UserAddr = sample.RandAccAddress()
	sp := sptypes.StorageProvider{
		OperatorAddress: s.PrimarySpAddr.String(),
		FundingAddress:  s.PrimarySpFundingAddr.String(),
	}
	s.depKeepers.SpKeeper.SetStorageProvider(ctx, &sp)
	for i := 0; i < 6; i++ {
		secondarySpAddr := sample.RandAccAddress()
		secondarySpFundingAddr := sample.RandAccAddress()
		secondarySp := sptypes.StorageProvider{
			OperatorAddress: secondarySpAddr.String(),
			FundingAddress:  secondarySpFundingAddr.String(),
		}
		s.SecondarySps = append(s.SecondarySps, secondarySp)
		s.depKeepers.SpKeeper.SetStorageProvider(ctx, &secondarySp)
	}
	s.depKeepers.SpKeeper.SetSpStoragePrice(ctx, sptypes.SpStoragePrice{
		SpAddress:     s.PrimarySpAddr.String(),
		UpdateTimeSec: 1,
		ReadPrice:     sdk.NewDec(2),
		StorePrice:    sdk.NewDec(5),
		FreeReadQuota: 10000,
	})
	s.depKeepers.SpKeeper.SetSecondarySpStorePrice(ctx, sptypes.SecondarySpStorePrice{
		UpdateTimeSec: 1,
		StorePrice:    sdk.NewDec(4),
	})
	coins := sdk.Coins{sdk.Coin{Denom: s.Denom, Amount: sdkmath.NewInt(1e18)}}
	bankKeeper := s.depKeepers.BankKeeper
	balances := bankKeeper.GetAllBalances(ctx, s.depKeepers.AccountKeeper.GetModuleAddress(authtypes.Minter))
	s.T().Logf("Minter module balances: %s", balances)
	err := bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.Minter, s.UserAddr, coins)
	s.Require().NoError(err)
	balance := bankKeeper.GetBalance(ctx, s.UserAddr, "BNB")
	s.T().Logf("s.UserAddr: %s, balance: %s", s.UserAddr, balance)
}

func (s *IntegrationTestSuiteWithoutMock) TestCreateCreateBucket_Payment() {
	ctx := s.ctx.WithBlockTime(time.Now())
	// mock create bucket
	ChargedReadQuota := uint64(1000)
	bucket := types.BucketInfo{
		ChargedReadQuota: ChargedReadQuota,
		PaymentAddress:   s.UserAddr.String(),
		PrimarySpAddress: s.PrimarySpAddr.String(),
	}
	t1 := int64(200)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(t1) * time.Second))
	err := s.keeper.ChargeInitialReadFee(ctx, &bucket)
	s.Require().NoError(err)
	userStreamRecordCreateBucket, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.UserAddr)
	s.Require().True(found)
	s.T().Logf("userStreamRecordCreateBucket: %+v", userStreamRecordCreateBucket)
	spStreamRecordCreateBucket, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.PrimarySpFundingAddr)
	s.Require().True(found)
	s.T().Logf("spStreamRecordCreateBucket: %+v", spStreamRecordCreateBucket)

	// mock add a object
	t2 := t1 + 5000
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(t2) * time.Second))
	bucket.BillingInfo.PriceTime = t2
	object := types.ObjectInfo{
		PayloadSize: 100,
		CreateAt:    ctx.BlockTime().Unix(),
	}
	err = s.keeper.LockStoreFee(ctx, &bucket, &object)
	s.Require().NoError(err)
	s.T().Logf("create object")
	userStreamRecordCreateObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.UserAddr)
	s.Require().True(found)
	s.T().Logf("userStreamRecordCreateObject: %+v", userStreamRecordCreateObject)
	spStreamRecordCreateObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.PrimarySpFundingAddr)
	s.Require().True(found)
	s.T().Logf("spStreamRecordCreateObject: %+v", spStreamRecordCreateObject)

	// mock seal object
	secondarySpAddresses := lo.Map(s.SecondarySps, func(sp sptypes.StorageProvider, index int) string {
		return sp.OperatorAddress
	})
	object.SecondarySpAddresses = secondarySpAddresses
	err = s.keeper.UnlockAndChargeStoreFee(ctx, &bucket, &object)
	s.Require().NoError(err)
	s.T().Logf("seal object")
	userStreamRecordSealObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.UserAddr)
	s.Require().True(found)
	s.T().Logf("userStreamRecordSealObject: %+v", userStreamRecordSealObject)
	spStreamRecordSealObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.PrimarySpFundingAddr)
	s.Require().True(found)
	s.T().Logf("spStreamRecordSealObject: %+v", spStreamRecordSealObject)

	// check
	primaryStorePriceRes, err := s.depKeepers.SpKeeper.GetSpStoragePriceByTime(ctx, s.PrimarySpAddr, t2)
	s.Require().NoError(err)
	s.T().Logf("primaryStorePriceRes: %+v", primaryStorePriceRes)
	primarySpRateDiff := spStreamRecordSealObject.NetflowRate.Sub(spStreamRecordCreateBucket.NetflowRate)
	expectedRate := primaryStorePriceRes.StorePrice.MulInt(sdk.NewIntFromUint64(bucket.BillingInfo.TotalChargeSize)).TruncateInt()
	readRate := primaryStorePriceRes.ReadPrice.MulInt(sdk.NewIntFromUint64(ChargedReadQuota)).TruncateInt()
	s.T().Logf("primarySpRateDiff: %s, expectedRate: %s, readRate: %s", primarySpRateDiff, expectedRate, readRate)
	s.Require().Equal(expectedRate.String(), primarySpRateDiff.String())
}

func TestKeeperTestSuiteWithoutMock(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuiteWithoutMock))
}
