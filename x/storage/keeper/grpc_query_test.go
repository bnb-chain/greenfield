package keeper_test

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func makeKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		&types.MockAccountKeeper{},
		&types.MockSpKeeper{},
		&types.MockPaymentKeeper{},
		&types.MockPermissionKeeper{},
		&types.MockCrossChainKeeper{},
		authtypes.NewModuleAddress(types.ModuleName).String(),
	)

	return k, testCtx.Ctx
}

func TestParamsQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
