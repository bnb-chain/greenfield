package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/x/challenge"
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

type TestSuite struct {
	suite.Suite

	cdc             codec.Codec
	challengeKeeper *keeper.Keeper

	bankKeeper    *types.MockBankKeeper
	storageKeeper *types.MockStorageKeeper
	spKeeper      *types.MockSpKeeper
	stakingKeeper *types.MockStakingKeeper
	paymentKeeper *types.MockPaymentKeeper

	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(challenge.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	bankKeeper := types.NewMockBankKeeper(ctrl)
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	stakingKeeper := types.NewMockStakingKeeper(ctrl)
	paymentKeeper := types.NewMockPaymentKeeper(ctrl)

	s.challengeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		bankKeeper,
		storageKeeper,
		spKeeper,
		stakingKeeper,
		paymentKeeper,
		authtypes.NewModuleAddress(types.ModuleName).String(),
	)

	s.cdc = encCfg.Codec
	s.bankKeeper = bankKeeper
	s.storageKeeper = storageKeeper
	s.spKeeper = spKeeper
	s.stakingKeeper = stakingKeeper
	s.paymentKeeper = paymentKeeper

	err := s.challengeKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.challengeKeeper)

	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.challengeKeeper)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
