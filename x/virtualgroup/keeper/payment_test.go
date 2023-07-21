package keeper_test

import (
	"errors"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/challenge"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/keeper"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type TestSuite struct {
	suite.Suite

	cdc                codec.Codec
	virtualgroupKeeper *keeper.Keeper

	bankKeeper    *types.MockBankKeeper
	accountKeeper *types.MockAccountKeeper
	spKeeper      *types.MockSpKeeper
	paymentKeeper *types.MockPaymentKeeper

	ctx sdk.Context
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(challenge.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))

	ctrl := gomock.NewController(s.T())

	bankKeeper := types.NewMockBankKeeper(ctrl)
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	paymentKeeper := types.NewMockPaymentKeeper(ctrl)

	s.ctx = testCtx.Ctx
	s.virtualgroupKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		spKeeper,
		accountKeeper,
		bankKeeper,
		paymentKeeper,
	)

	s.cdc = encCfg.Codec
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.spKeeper = spKeeper
	s.paymentKeeper = paymentKeeper

	err := s.virtualgroupKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestSettleAndDistributeGVGFamily() {
	sp := &sptypes.StorageProvider{Id: 1, FundingAddress: sample.RandAccAddress().String()}
	family := &types.GlobalVirtualGroupFamily{Id: 1, VirtualPaymentAddress: sample.RandAccAddress().String()}

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), nil)
	err := s.virtualgroupKeeper.SettleAndDistributeGVGFamily(s.ctx, sp, family)
	require.NoError(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), errors.New("error"))
	err = s.virtualgroupKeeper.SettleAndDistributeGVGFamily(s.ctx, sp, family)
	require.Error(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1024), nil)
	s.paymentKeeper.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
	err = s.virtualgroupKeeper.SettleAndDistributeGVGFamily(s.ctx, sp, family)
	require.NoError(s.T(), err)
}

func (s *TestSuite) TestSettleAndDistributeGVG() {
	gvg := &types.GlobalVirtualGroup{Id: 1,
		VirtualPaymentAddress: sample.RandAccAddress().String(),
		SecondarySpIds:        []uint32{3, 6, 9}}

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), nil)
	err := s.virtualgroupKeeper.SettleAndDistributeGVG(s.ctx, gvg)
	require.NoError(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), errors.New("error"))
	err = s.virtualgroupKeeper.SettleAndDistributeGVG(s.ctx, gvg)
	require.Error(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1024), nil).AnyTimes()
	s.paymentKeeper.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	sp := &sptypes.StorageProvider{Id: 1, FundingAddress: sample.RandAccAddress().String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Any()).
		Return(sp, true).AnyTimes()
	err = s.virtualgroupKeeper.SettleAndDistributeGVG(s.ctx, gvg)
	require.NoError(s.T(), err)
}
