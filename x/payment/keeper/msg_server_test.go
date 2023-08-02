package keeper_test

import (
	"testing"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/cosmos/cosmos-sdk/baseapp"
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
	"github.com/bnb-chain/greenfield/x/payment/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

type TestSuite struct {
	suite.Suite

	cdc           codec.Codec
	paymentKeeper *keeper.Keeper

	bankKeeper    *types.MockBankKeeper
	accountKeeper *types.MockAccountKeeper
	spKeeper      *types.MockSpKeeper

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
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)

	s.paymentKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		bankKeeper,
		accountKeeper,
		spKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.cdc = encCfg.Codec
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.spKeeper = spKeeper

	err := s.paymentKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.paymentKeeper)

	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.paymentKeeper)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestUpdateParams() {
	params := types.DefaultParams()
	params.MaxAutoResumeFlowCount = 5

	tests := []struct {
		name string
		msg  types.MsgUpdateParams
		err  bool
	}{
		{
			name: "invalid authority",
			msg: types.MsgUpdateParams{
				Authority: sample.RandAccAddressHex(),
			},
			err: true,
		}, {
			name: "success",
			msg: types.MsgUpdateParams{
				Authority: s.paymentKeeper.GetAuthority(),
				Params:    params,
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			_, err := s.msgServer.UpdateParams(s.ctx, &tt.msg)
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}

	// verify storage
	s.Require().Equal(params, s.paymentKeeper.GetParams(s.ctx))
}
