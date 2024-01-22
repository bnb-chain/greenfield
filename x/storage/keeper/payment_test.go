package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/challenge"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type TestSuite struct {
	suite.Suite

	cdc           codec.Codec
	storageKeeper *keeper.Keeper

	accountKeeper      *types.MockAccountKeeper
	spKeeper           *types.MockSpKeeper
	permissionKeeper   *types.MockPermissionKeeper
	crossChainKeeper   *types.MockCrossChainKeeper
	paymentKeeper      *types.MockPaymentKeeper
	virtualGroupKeeper *types.MockVirtualGroupKeeper

	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(challenge.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	upgradeChecker := func(sdk.Context, string) bool { return true }
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	header := testCtx.Ctx.BlockHeader()
	header.Time = time.Now()
	testCtx = testutil.TestContext{
		Ctx: sdk.NewContext(testCtx.CMS, header, false, upgradeChecker, testCtx.Ctx.Logger()),
		DB:  testCtx.DB,
		CMS: testCtx.CMS,
	}
	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	accountKeeper := types.NewMockAccountKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	permissionKeeper := types.NewMockPermissionKeeper(ctrl)
	crossChainKeeper := types.NewMockCrossChainKeeper(ctrl)
	paymentKeeper := types.NewMockPaymentKeeper(ctrl)
	virtualGroupKeeper := types.NewMockVirtualGroupKeeper(ctrl)

	s.storageKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		accountKeeper,
		spKeeper,
		paymentKeeper,
		permissionKeeper,
		crossChainKeeper,
		virtualGroupKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.cdc = encCfg.Codec
	s.accountKeeper = accountKeeper
	s.spKeeper = spKeeper
	s.permissionKeeper = permissionKeeper
	s.crossChainKeeper = crossChainKeeper
	s.paymentKeeper = paymentKeeper
	s.virtualGroupKeeper = virtualGroupKeeper

	err := s.storageKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.storageKeeper)

	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.storageKeeper)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestGetObjectLockFee() {
	primarySp := &sptypes.StorageProvider{Status: sptypes.STATUS_IN_SERVICE, Id: 100, OperatorAddress: sample.RandAccAddress().String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(primarySp.Id)).
		Return(primarySp, true).AnyTimes()

	price := sptypes.GlobalSpStorePrice{
		ReadPrice:           sdk.NewDec(100),
		PrimaryStorePrice:   sdk.NewDec(1000),
		SecondaryStorePrice: sdk.NewDec(500),
	}
	s.spKeeper.EXPECT().GetGlobalSpStorePriceByTime(gomock.Any(), gomock.Any()).
		Return(price, nil).AnyTimes()
	params := paymenttypes.DefaultParams()
	s.paymentKeeper.EXPECT().GetVersionedParamsWithTs(gomock.Any(), gomock.Any()).
		Return(params.VersionedParams, nil).AnyTimes()

	// verify lock fee calculation
	payloadSize := int64(10 * 1024 * 1024)
	amount, err := s.storageKeeper.GetObjectLockFee(s.ctx, time.Now().Unix(), uint64(payloadSize))
	s.Require().NoError(err)
	secondarySPNum := int64(s.storageKeeper.GetExpectSecondarySPNumForECObject(s.ctx, time.Now().Unix()))
	spRate := price.PrimaryStorePrice.Add(price.SecondaryStorePrice.MulInt64(secondarySPNum)).MulInt64(payloadSize)
	validatorTaxRate := params.VersionedParams.ValidatorTaxRate.MulInt(spRate.TruncateInt())
	expectedAmount := spRate.Add(validatorTaxRate).MulInt64(int64(params.VersionedParams.ReserveTime)).TruncateInt()
	s.Require().True(amount.Equal(expectedAmount))
}

func (s *TestSuite) TestGetBucketReadBill() {
	gvgFamily := &virtualgroupmoduletypes.GlobalVirtualGroupFamily{
		Id:                    1,
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), gomock.Any()).
		Return(gvgFamily, true).AnyTimes()

	primarySp := &sptypes.StorageProvider{
		Status:          sptypes.STATUS_IN_SERVICE,
		Id:              100,
		OperatorAddress: sample.RandAccAddress().String(),
		FundingAddress:  sample.RandAccAddress().String(),
	}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(primarySp.Id)).
		Return(primarySp, true).AnyTimes()

	price := sptypes.GlobalSpStorePrice{
		ReadPrice:           sdk.NewDec(100),
		PrimaryStorePrice:   sdk.NewDec(1000),
		SecondaryStorePrice: sdk.NewDec(500),
	}
	s.spKeeper.EXPECT().GetGlobalSpStorePriceByTime(gomock.Any(), gomock.Any()).
		Return(price, nil).AnyTimes()
	params := paymenttypes.DefaultParams()
	s.paymentKeeper.EXPECT().GetVersionedParamsWithTs(gomock.Any(), gomock.Any()).
		Return(params.VersionedParams, nil).AnyTimes()

	// empty bucket, zero read quota
	bucketInfo := &types.BucketInfo{
		Owner:                      "",
		BucketName:                 "bucketname",
		Id:                         sdk.NewUint(1),
		PaymentAddress:             sample.RandAccAddress().String(),
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
		ChargedReadQuota:           0,
	}
	internalBucketInfo := &types.InternalBucketInfo{}
	flows, err := s.storageKeeper.GetBucketReadStoreBill(s.ctx, bucketInfo, internalBucketInfo)
	s.Require().NoError(err)
	s.Require().True(len(flows.Flows) == 0)

	// empty bucket
	bucketInfo = &types.BucketInfo{
		Owner:                      "",
		BucketName:                 "bucketname",
		Id:                         sdk.NewUint(1),
		PaymentAddress:             sample.RandAccAddress().String(),
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
		ChargedReadQuota:           100,
	}
	internalBucketInfo = &types.InternalBucketInfo{}
	flows, err = s.storageKeeper.GetBucketReadStoreBill(s.ctx, bucketInfo, internalBucketInfo)
	s.Require().NoError(err)
	readRate := price.ReadPrice.MulInt64(int64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.Require().Equal(flows.Flows[0].ToAddress, gvgFamily.VirtualPaymentAddress)
	s.Require().Equal(flows.Flows[0].Rate, readRate)
	taxPoolRate := params.VersionedParams.ValidatorTaxRate.MulInt(readRate).TruncateInt()
	s.Require().Equal(flows.Flows[1].ToAddress, paymenttypes.ValidatorTaxPoolAddress.String())
	s.Require().Equal(flows.Flows[1].Rate, taxPoolRate)
}

func (s *TestSuite) TestGetBucketReadStoreBill() {
	gvgFamily := &virtualgroupmoduletypes.GlobalVirtualGroupFamily{
		Id:                    1,
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), gomock.Any()).
		Return(gvgFamily, true).AnyTimes()

	primarySp := &sptypes.StorageProvider{
		Status:          sptypes.STATUS_IN_SERVICE,
		Id:              100,
		OperatorAddress: sample.RandAccAddress().String(),
		FundingAddress:  sample.RandAccAddress().String(),
	}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(primarySp.Id)).
		Return(primarySp, true).AnyTimes()

	price := sptypes.GlobalSpStorePrice{
		ReadPrice:           sdk.NewDec(100),
		PrimaryStorePrice:   sdk.NewDec(1000),
		SecondaryStorePrice: sdk.NewDec(500),
	}
	s.spKeeper.EXPECT().GetGlobalSpStorePriceByTime(gomock.Any(), gomock.Any()).
		Return(price, nil).AnyTimes()
	params := paymenttypes.DefaultParams()
	s.paymentKeeper.EXPECT().GetVersionedParamsWithTs(gomock.Any(), gomock.Any()).
		Return(params.VersionedParams, nil).AnyTimes()

	// none empty bucket
	bucketInfo := &types.BucketInfo{
		Owner:                      "",
		BucketName:                 "bucketname",
		Id:                         sdk.NewUint(1),
		PaymentAddress:             sample.RandAccAddress().String(),
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
		ChargedReadQuota:           100,
	}

	lvg1 := &types.LocalVirtualGroup{
		Id:                   1,
		TotalChargeSize:      100,
		GlobalVirtualGroupId: 1,
	}
	lvg2 := &types.LocalVirtualGroup{
		Id:                   2,
		TotalChargeSize:      200,
		GlobalVirtualGroupId: 2,
	}
	internalBucketInfo := &types.InternalBucketInfo{
		TotalChargeSize: 300,
		LocalVirtualGroups: []*types.LocalVirtualGroup{
			lvg1, lvg2,
		},
	}

	gvg1 := &virtualgroupmoduletypes.GlobalVirtualGroup{
		Id:                    1,
		PrimarySpId:           primarySp.Id,
		SecondarySpIds:        []uint32{101, 102, 103, 104, 105, 106},
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	gvg2 := &virtualgroupmoduletypes.GlobalVirtualGroup{
		Id:                    2,
		PrimarySpId:           primarySp.Id,
		SecondarySpIds:        []uint32{201, 202, 203, 204, 205, 206},
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	s.virtualGroupKeeper.EXPECT().GetGVG(gomock.Any(), gvg1.Id).
		Return(gvg1, true).AnyTimes()
	s.virtualGroupKeeper.EXPECT().GetGVG(gomock.Any(), gvg2.Id).
		Return(gvg2, true).AnyTimes()

	flows, err := s.storageKeeper.GetBucketReadStoreBill(s.ctx, bucketInfo, internalBucketInfo)
	s.Require().NoError(err)

	// read rate to gvg family
	s.Require().Equal(flows.Flows[0].ToAddress, gvgFamily.VirtualPaymentAddress)
	readRate := price.ReadPrice.MulInt64(int64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.Require().Equal(flows.Flows[0].Rate, readRate)

	// read rate to validator tax pool
	s.Require().Equal(flows.Flows[1].ToAddress, paymenttypes.ValidatorTaxPoolAddress.String())
	taxPoolRate := params.VersionedParams.ValidatorTaxRate.MulInt(readRate).TruncateInt()
	s.Require().Equal(flows.Flows[1].Rate, taxPoolRate)

	// first gvg
	// store rate to gvg family
	s.Require().Equal(flows.Flows[2].ToAddress, gvgFamily.VirtualPaymentAddress)
	primaryStoreRate := price.PrimaryStorePrice.MulInt64(int64(lvg1.TotalChargeSize)).TruncateInt()
	s.Require().Equal(flows.Flows[2].Rate, primaryStoreRate)

	// store rate to gvg
	gvg1StoreSize := lvg1.TotalChargeSize * uint64(len(gvg1.SecondarySpIds))
	gvg1StoreRate := price.SecondaryStorePrice.MulInt64(int64(gvg1StoreSize)).TruncateInt()
	s.Require().Equal(flows.Flows[3].ToAddress, gvg1.VirtualPaymentAddress)
	s.Require().Equal(flows.Flows[3].Rate, gvg1StoreRate)

	// store rate to validator tax pool
	s.Require().Equal(flows.Flows[4].ToAddress, paymenttypes.ValidatorTaxPoolAddress.String())
	taxPoolRate = params.VersionedParams.ValidatorTaxRate.MulInt(primaryStoreRate.Add(gvg1StoreRate)).TruncateInt()
	s.Require().Equal(flows.Flows[4].Rate, taxPoolRate)

	// secondary gvg
	// store rate to gvg family
	s.Require().Equal(flows.Flows[5].ToAddress, gvgFamily.VirtualPaymentAddress)
	primaryStoreRate = price.PrimaryStorePrice.MulInt64(int64(lvg2.TotalChargeSize)).TruncateInt()
	s.Require().Equal(flows.Flows[5].Rate, primaryStoreRate)

	// store rate to gvg
	gvg2StoreSize := lvg2.TotalChargeSize * uint64(len(gvg2.SecondarySpIds))
	gvg2StoreRate := price.SecondaryStorePrice.MulInt64(int64(gvg2StoreSize)).TruncateInt()
	s.Require().Equal(flows.Flows[6].ToAddress, gvg2.VirtualPaymentAddress)
	s.Require().Equal(flows.Flows[6].Rate, gvg2StoreRate)

	// store rate to validator tax pool
	s.Require().Equal(flows.Flows[7].ToAddress, paymenttypes.ValidatorTaxPoolAddress.String())
	taxPoolRate = params.VersionedParams.ValidatorTaxRate.MulInt(primaryStoreRate.Add(gvg2StoreRate)).TruncateInt()
	s.Require().Equal(flows.Flows[7].Rate, taxPoolRate)
}
