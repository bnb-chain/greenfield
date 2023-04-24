package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/bnb-chain/greenfield/testutil/sample"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	//depKeepers    keepertest.StorageDepKeepers
	paymentKeeper    *types.MockPaymentKeeper
	spKeeper         *types.MockSpKeeper
	bankKeeper       *types.MockBankKeeper
	accountKeeper    *types.MockAccountKeeper
	permissionKeeper *types.MockPermissionKeeper
	crossChainKeeper *types.MockCrossChainKeeper
	ctx              sdk.Context
	PrimarySpAddr    sdk.AccAddress
	UserAddr         sdk.AccAddress
	Denom            string
}

func (s *IntegrationTestSuite) SetupTest() {
	s.Denom = "BNB"
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)

	ctrl := gomock.NewController(s.T())

	paymentKeeper := types.NewMockPaymentKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	bankKeeper := types.NewMockBankKeeper(ctrl)
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	permissionKeeper := types.NewMockPermissionKeeper(ctrl)
	crossChainKeeper := types.NewMockCrossChainKeeper(ctrl)

	s.keeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		spKeeper,
		paymentKeeper,
		permissionKeeper,
		crossChainKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	ctx := s.ctx.WithBlockTime(time.Now())
	// init data
	s.PrimarySpAddr = sample.RandAccAddress()
	s.UserAddr = sample.RandAccAddress()
	s.spKeeper.EXPECT().SetSpStoragePrice(ctx, sptypes.SpStoragePrice{
		SpAddress:     s.PrimarySpAddr.String(),
		UpdateTimeSec: 1,
		ReadPrice:     sdk.NewDec(2),
		StorePrice:    sdk.NewDec(5),
		FreeReadQuota: 10000,
	})
	spKeeper.EXPECT().SetSecondarySpStorePrice(ctx, sptypes.SecondarySpStorePrice{
		UpdateTimeSec: 1,
		StorePrice:    sdk.NewDec(4),
	})
	coins := sdk.Coins{sdk.Coin{Denom: s.Denom, Amount: sdkmath.NewInt(1e18)}}
	balances := coins
	bankKeeper.EXPECT().GetBalance(ctx, s.UserAddr, "BNB").Return(balances).AnyTimes()
}

// this may should put into payment module test
//func (s *IntegrationTestSuite) TestCreateCreateBucket_Payment() {
//	ctx := s.ctx.WithBlockTime(time.Now())
//	// mock create bucket
//	ChargedReadQuota := uint64(1000)
//	bucket := types.BucketInfo{
//		ChargedReadQuota: ChargedReadQuota,
//		PaymentAddress:   s.UserAddr.String(),
//		PrimarySpAddress: s.PrimarySpAddr.String(),
//	}
//	t1 := int64(200)
//	ctx = ctx.WithBlockTime(time.Unix(t1, 0))
//	err := s.keeper.ChargeInitialReadFee(ctx, &bucket)
//	s.Require().NoError(err)
//	userStreamRecordCreateBucket, found := s.paymentKeeper.EXPECT().GetStreamRecord(ctx, s.UserAddr)
//	s.Require().True(found)
//	s.T().Logf("userStreamRecordCreateBucket: %+v", userStreamRecordCreateBucket)
//	spStreamRecordCreateBucket, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.PrimarySpAddr)
//	s.Require().True(found)
//	s.T().Logf("spStreamRecordCreateBucket: %+v", spStreamRecordCreateBucket)
//
//	// mock add a object
//	t2 := t1 + 5000
//	ctx = ctx.WithBlockTime(time.Unix(t2, 0))
//	bucket.BillingInfo.PriceTime = t2
//	object := types.ObjectInfo{
//		PayloadSize: 100,
//		CreateAt:    t2,
//	}
//	err = s.keeper.LockStoreFee(ctx, &bucket, &object)
//	s.Require().NoError(err)
//	s.T().Logf("create object")
//	userStreamRecordCreateObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.UserAddr)
//	s.Require().True(found)
//	s.T().Logf("userStreamRecordCreateObject: %+v", userStreamRecordCreateObject)
//	spStreamRecordCreateObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.PrimarySpAddr)
//	s.Require().True(found)
//	s.T().Logf("spStreamRecordCreateObject: %+v", spStreamRecordCreateObject)
//
//	// mock seal object
//	var secondarySpAddresses []string
//	for i := 0; i < 6; i++ {
//		secondarySpAddresses = append(secondarySpAddresses, sample.RandAccAddress().String())
//	}
//	object.SecondarySpAddresses = secondarySpAddresses
//	err = s.keeper.UnlockAndChargeStoreFee(ctx, &bucket, &object)
//	s.Require().NoError(err)
//	s.T().Logf("seal object")
//	userStreamRecordSealObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.UserAddr)
//	s.Require().True(found)
//	s.T().Logf("userStreamRecordSealObject: %+v", userStreamRecordSealObject)
//	spStreamRecordSealObject, found := s.depKeepers.PaymentKeeper.GetStreamRecord(ctx, s.PrimarySpAddr)
//	s.Require().True(found)
//	s.T().Logf("spStreamRecordSealObject: %+v", spStreamRecordSealObject)
//
//	// check
//	primaryStorePriceRes, err := s.spKeeper.EXPECT().GetSpStoragePriceByTime(ctx, s.PrimarySpAddr, t2)
//	s.Require().NoError(err)
//	s.T().Logf("primaryStorePriceRes: %+v", primaryStorePriceRes)
//	primarySpRateDiff := spStreamRecordSealObject.NetflowRate.Sub(spStreamRecordCreateBucket.NetflowRate)
//	expectedRate := primaryStorePriceRes.StorePrice.MulInt(sdk.NewIntFromUint64(bucket.BillingInfo.TotalChargeSize)).TruncateInt()
//	readRate := primaryStorePriceRes.ReadPrice.MulInt(sdk.NewIntFromUint64(ChargedReadQuota)).TruncateInt()
//	s.T().Logf("primarySpRateDiff: %s, expectedRate: %s, readRate: %s", primarySpRateDiff, expectedRate, readRate)
//	s.Require().Equal(expectedRate.String(), primarySpRateDiff.String())
//}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
