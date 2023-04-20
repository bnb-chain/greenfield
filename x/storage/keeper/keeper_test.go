package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/testutil/sample"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	keeper           *keeper.Keeper
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
