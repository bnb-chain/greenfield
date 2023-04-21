package keeper_test

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/permission/keeper"
	"github.com/bnb-chain/greenfield/x/permission/types"
)

func makeKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		&types.MockAccountKeeper{},
		authtypes.NewModuleAddress(types.ModuleName).String(),
	)

	return k, testCtx.Ctx
}

func TestGetParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
