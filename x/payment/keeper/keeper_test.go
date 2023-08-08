package keeper_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/x/payment"
	"github.com/bnb-chain/greenfield/x/payment/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

type DepKeepers struct {
	BankKeeper    *types.MockBankKeeper
	AccountKeeper *types.MockAccountKeeper
}

func makePaymentKeeper(t *testing.T) (*keeper.Keeper, sdk.Context, DepKeepers) {
	encCfg := moduletestutil.MakeTestEncodingConfig(payment.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	ctrl := gomock.NewController(t)
	bankKeeper := types.NewMockBankKeeper(ctrl)
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		bankKeeper,
		accountKeeper,
		authtypes.NewModuleAddress(types.ModuleName).String(),
	)
	err := k.SetParams(testCtx.Ctx, types.DefaultParams())
	if err != nil {
		panic(err)
	}

	depKeepers := DepKeepers{
		BankKeeper:    bankKeeper,
		AccountKeeper: accountKeeper,
	}

	return k, testCtx.Ctx, depKeepers
}
