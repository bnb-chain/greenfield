package keeper_test

import (
	"testing"
	"time"

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

// TODO: add more tests
