package keeper_test

import (
	"testing"
	"time"

	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"

	"github.com/bnb-chain/greenfield/testutil/sample"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"

	"github.com/bnb-chain/greenfield/x/challenge"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
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
)

type TestSuite struct {
	suite.Suite

	cdc           codec.Codec
	storageKeeper *keeper.Keeper

	accountKeeper      *types.MockAccountKeeper
	spKeeper           *types.MockSpKeeper
	permissionKeeper   *types.MockPermissionKeeper
	crosschainKeeper   *types.MockCrossChainKeeper
	paymentKeeper      *types.MockPaymentKeeper
	virtualGroupKeeper *types.MockVirtualGroupKeeper

	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(challenge.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	header := testCtx.Ctx.BlockHeader()
	header.Time = time.Now()
	testCtx = testutil.TestContext{
		Ctx: sdk.NewContext(testCtx.CMS, header, false, nil, testCtx.Ctx.Logger()),
		DB:  testCtx.DB,
		CMS: testCtx.CMS,
	}
	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	accountKeeper := types.NewMockAccountKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	permissionKeeper := types.NewMockPermissionKeeper(ctrl)
	crosschainKeeper := types.NewMockCrossChainKeeper(ctrl)
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
		crosschainKeeper,
		virtualGroupKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.cdc = encCfg.Codec
	s.accountKeeper = accountKeeper
	s.spKeeper = spKeeper
	s.permissionKeeper = permissionKeeper
	s.crosschainKeeper = crosschainKeeper
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

	price := paymenttypes.StoragePrice{
		ReadPrice:           sdk.NewDec(100),
		PrimaryStorePrice:   sdk.NewDec(1000),
		SecondaryStorePrice: sdk.NewDec(500),
	}
	s.paymentKeeper.EXPECT().GetStoragePrice(gomock.Any(), gomock.Any()).
		Return(price, nil).AnyTimes()
	params := paymenttypes.DefaultParams()
	s.paymentKeeper.EXPECT().GetParams(gomock.Any()).
		Return(params).AnyTimes()

	// verify lock fee calculation
	payloadSize := int64(10 * 1024 * 1024)
	amount, err := s.storageKeeper.GetObjectLockFee(s.ctx, sample.RandAccAddress().String(), time.Now().Unix(), uint64(payloadSize))
	s.Require().NoError(err)
	expectedAmount := price.PrimaryStorePrice.Add(price.SecondaryStorePrice.MulInt64(types.SecondarySPNum)).
		MulInt64(payloadSize).MulInt64(int64(params.ReserveTime)).TruncateInt()
	s.Require().True(amount.Equal(expectedAmount))
}

func (s *TestSuite) TestGetBucketBill() {
	gvgFamily := &virtualgroupmoduletypes.GlobalVirtualGroupFamily{
		Id:                    1,
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(gvgFamily, true).AnyTimes()

	primarySp := &sptypes.StorageProvider{
		Status:          sptypes.STATUS_IN_SERVICE,
		Id:              100,
		OperatorAddress: sample.RandAccAddress().String(),
		FundingAddress:  sample.RandAccAddress().String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(primarySp.Id)).
		Return(primarySp, true).AnyTimes()

	price := paymenttypes.StoragePrice{
		ReadPrice:           sdk.NewDec(100),
		PrimaryStorePrice:   sdk.NewDec(1000),
		SecondaryStorePrice: sdk.NewDec(500),
	}
	s.paymentKeeper.EXPECT().GetStoragePrice(gomock.Any(), gomock.Any()).
		Return(price, nil).AnyTimes()
	params := paymenttypes.DefaultParams()
	s.paymentKeeper.EXPECT().GetParams(gomock.Any()).
		Return(params).AnyTimes()

	// empty bucket, zero read quota
	bucketInfo := &types.BucketInfo{
		Owner:                      "",
		BucketName:                 "bucketname",
		Id:                         sdk.NewUint(1),
		PaymentAddress:             sample.RandAccAddress().String(),
		PrimarySpId:                primarySp.Id,
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
		ChargedReadQuota:           0,
		BillingInfo: types.BillingInfo{
			TotalChargeSize: 0,
		},
	}
	flows, err := s.storageKeeper.GetBucketBill(s.ctx, bucketInfo)
	s.Require().NoError(err)
	s.Require().True(len(flows.Flows) == 0)

	// empty bucket
	bucketInfo = &types.BucketInfo{
		Owner:                      "",
		BucketName:                 "bucketname",
		Id:                         sdk.NewUint(1),
		PaymentAddress:             sample.RandAccAddress().String(),
		PrimarySpId:                primarySp.Id,
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
		ChargedReadQuota:           100,
		BillingInfo: types.BillingInfo{
			TotalChargeSize: 0,
		},
	}
	flows, err = s.storageKeeper.GetBucketBill(s.ctx, bucketInfo)
	s.Require().NoError(err)
	readRate := price.ReadPrice.MulInt64(int64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.Require().Equal(flows.Flows[0].ToAddress, gvgFamily.VirtualPaymentAddress)
	s.Require().Equal(flows.Flows[0].Rate, readRate)
	taxPoolRate := s.paymentKeeper.GetParams(s.ctx).ValidatorTaxRate.MulInt(readRate).TruncateInt()
	s.Require().Equal(flows.Flows[1].ToAddress, paymenttypes.ValidatorTaxPoolAddress.String())
	s.Require().Equal(flows.Flows[1].Rate, taxPoolRate)

	// none empty bucket
	bucketInfo = &types.BucketInfo{
		Owner:                      "",
		BucketName:                 "bucketname",
		Id:                         sdk.NewUint(1),
		PaymentAddress:             sample.RandAccAddress().String(),
		PrimarySpId:                primarySp.Id,
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
		ChargedReadQuota:           100,
		BillingInfo: types.BillingInfo{
			TotalChargeSize: 300,
			LvgObjectsSize: []types.LVGObjectsSize{
				{
					LvgId:           1,
					TotalChargeSize: 100,
				}, {
					LvgId:           2,
					TotalChargeSize: 200,
				},
			},
		},
	}
	lvg1 := &virtualgroupmoduletypes.LocalVirtualGroup{
		Id:                   1,
		BucketId:             bucketInfo.Id,
		GlobalVirtualGroupId: 1,
	}
	lvg2 := &virtualgroupmoduletypes.LocalVirtualGroup{
		Id:                   2,
		BucketId:             bucketInfo.Id,
		GlobalVirtualGroupId: 2,
	}
	s.virtualGroupKeeper.EXPECT().GetLVG(gomock.Any(), gomock.Any(), lvg1.Id).
		Return(lvg1, true).AnyTimes()
	s.virtualGroupKeeper.EXPECT().GetLVG(gomock.Any(), gomock.Any(), lvg2.Id).
		Return(lvg2, true).AnyTimes()

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

	flows, err = s.storageKeeper.GetBucketBill(s.ctx, bucketInfo)
	s.Require().NoError(err)

	gvg1StoreSize := bucketInfo.BillingInfo.LvgObjectsSize[0].TotalChargeSize * uint64(len(gvg1.SecondarySpIds))
	gvg1StoreRate := price.SecondaryStorePrice.MulInt64(int64(gvg1StoreSize)).TruncateInt()
	s.Require().Equal(flows.Flows[0].ToAddress, gvg1.VirtualPaymentAddress)
	s.Require().Equal(flows.Flows[0].Rate, gvg1StoreRate)

	gvg2StoreSize := bucketInfo.BillingInfo.LvgObjectsSize[1].TotalChargeSize * uint64(len(gvg2.SecondarySpIds))
	gvg2StoreRate := price.SecondaryStorePrice.MulInt64(int64(gvg2StoreSize)).TruncateInt()
	s.Require().Equal(flows.Flows[1].ToAddress, gvg2.VirtualPaymentAddress)
	s.Require().Equal(flows.Flows[1].Rate, gvg2StoreRate)

	readRate = price.ReadPrice.MulInt64(int64(bucketInfo.ChargedReadQuota)).TruncateInt()
	primaryStoreRate := price.PrimaryStorePrice.MulInt64(int64(bucketInfo.BillingInfo.TotalChargeSize)).TruncateInt()
	s.Require().Equal(flows.Flows[2].ToAddress, gvgFamily.VirtualPaymentAddress)
	s.Require().Equal(flows.Flows[2].Rate, readRate.Add(primaryStoreRate))

	totalRate := readRate.Add(primaryStoreRate).Add(gvg1StoreRate).Add(gvg2StoreRate)
	taxPoolRate = s.paymentKeeper.GetParams(s.ctx).ValidatorTaxRate.MulInt(totalRate).TruncateInt()
	s.Require().Equal(flows.Flows[3].ToAddress, paymenttypes.ValidatorTaxPoolAddress.String())
	s.Require().Equal(flows.Flows[3].Rate, taxPoolRate)
}
