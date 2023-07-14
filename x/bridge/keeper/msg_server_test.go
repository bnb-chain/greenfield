package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/x/bridge/keeper"
	"github.com/bnb-chain/greenfield/x/bridge/types"
)

type TestSuite struct {
	suite.Suite

	cdc          codec.Codec
	bridgeKeeper *keeper.Keeper

	bankKeeper       *types.MockBankKeeper
	crossChainKeeper *types.MockCrossChainKeeper
	stakingKeeper    *types.MockStakingKeeper

	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	crossChainKeeper := types.NewMockCrossChainKeeper(ctrl)
	bankKeeper := types.NewMockBankKeeper(ctrl)
	stakingKeeper := types.NewMockStakingKeeper(ctrl)

	s.bridgeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		bankKeeper,
		stakingKeeper,
		crossChainKeeper,
		authtypes.NewModuleAddress(types.ModuleName).String(),
	)

	s.cdc = encCfg.Codec

	s.stakingKeeper = stakingKeeper
	s.bankKeeper = bankKeeper
	s.crossChainKeeper = crossChainKeeper

	err := s.bridgeKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.bridgeKeeper)

	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.bridgeKeeper)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestCrossTransferOut() {
	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")

	addr2, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")

	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	s.crossChainKeeper.EXPECT().GetDestBscChainID().Return(sdk.ChainID(714)).AnyTimes()
	s.crossChainKeeper.EXPECT().CreateRawIBCPackageWithFee(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), nil).AnyTimes()

	msgTransferOut := types.NewMsgTransferOut(addr1.String(), addr2.String(), &sdk.Coin{
		Denom:  "BNB",
		Amount: sdk.NewInt(1),
	})

	_, err = s.msgServer.TransferOut(s.ctx, msgTransferOut)
	s.Require().Nil(err, "error should be nil")
}

func (s *TestSuite) TestCrossTransferOutWrong() {
	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")

	addr2, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")

	msgTransferOut := types.NewMsgTransferOut(addr1.String(), addr2.String(), &sdk.Coin{
		Denom:  "wrongdenom",
		Amount: sdk.NewInt(1),
	})

	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()

	_, err = s.msgServer.TransferOut(s.ctx, msgTransferOut)
	s.Require().NotNil(err, "error should not be nil")
	s.Require().Contains(err.Error(), "denom is not supported")
}
