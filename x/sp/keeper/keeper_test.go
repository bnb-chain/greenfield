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
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc      codec.Codec
	spKeeper *keeper.Keeper

	bankKeeper    *types.MockBankKeeper
	accountKeeper *types.MockAccountKeeper
	authzKeeper   *types.MockAuthzKeeper

	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	bankKeeper := types.NewMockBankKeeper(ctrl)
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	authzKeeper := types.NewMockAuthzKeeper(ctrl)

	s.spKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		authzKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.cdc = encCfg.Codec

	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.authzKeeper = authzKeeper

	err := s.spKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.spKeeper)

	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.spKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
