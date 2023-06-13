package keeper_test

import (
	"testing"

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

	"github.com/bnb-chain/greenfield/x/challenge"
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

type TestSuite struct {
	suite.Suite

	cdc             codec.Codec
	challengeKeeper *keeper.Keeper

	bankKeeper         *types.MockBankKeeper
	storageKeeper      *types.MockStorageKeeper
	spKeeper           *types.MockSpKeeper
	stakingKeeper      *types.MockStakingKeeper
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

	// set mock randao mix
	randaoMix := sdk.Keccak256([]byte{1})
	randaoMix = append(randaoMix, sdk.Keccak256([]byte{2})...)
	header := testCtx.Ctx.BlockHeader()
	header.RandaoMix = randaoMix
	testCtx = testutil.TestContext{
		Ctx: sdk.NewContext(testCtx.CMS, header, false, nil, testCtx.Ctx.Logger()),
		DB:  testCtx.DB,
		CMS: testCtx.CMS,
	}

	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	bankKeeper := types.NewMockBankKeeper(ctrl)
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	stakingKeeper := types.NewMockStakingKeeper(ctrl)
	paymentKeeper := types.NewMockPaymentKeeper(ctrl)
	virtualGroupKeeper := types.NewMockVirtualGroupKeeper(ctrl)

	s.challengeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		bankKeeper,
		storageKeeper,
		spKeeper,
		stakingKeeper,
		paymentKeeper,
		virtualGroupKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.cdc = encCfg.Codec
	s.bankKeeper = bankKeeper
	s.storageKeeper = storageKeeper
	s.spKeeper = spKeeper
	s.stakingKeeper = stakingKeeper
	s.paymentKeeper = paymentKeeper
	s.virtualGroupKeeper = virtualGroupKeeper

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
