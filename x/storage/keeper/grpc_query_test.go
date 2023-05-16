package keeper_test

import (
	"testing"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func makeKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	tStorekey := storetypes.NewTransientStoreKey(types.TStoreKey)

	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		tStorekey,
		&types.MockAccountKeeper{},
		&types.MockSpKeeper{},
		&types.MockPaymentKeeper{},
		&types.MockPermissionKeeper{},
		&types.MockCrossChainKeeper{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	return k, testCtx.Ctx
}

func TestParamsQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestVersionedParamsQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	params := types.DefaultParams()
	params.VersionedParams.MaxSegmentSize = 1
	blockTimeT1 := ctx.BlockTime().Unix()
	versionedParamsT1 := params.VersionedParams
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	blockTimeT2 := ctx.BlockTime().Unix()
	params.VersionedParams.MaxSegmentSize = 2
	versionedParamsT2 := params.VersionedParams
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	responseT1, err := keeper.VersionedParams(ctx, &types.QueryVersionedParamsRequest{CurrentTimestamp: blockTimeT1})
	require.NoError(t, err)
	require.Equal(t, &types.QueryVersionedParamsResponse{VersionedParams: versionedParamsT1}, responseT1)
	require.EqualValues(t, responseT1.GetVersionedParams().MaxSegmentSize, 1)

	responseT2, err := keeper.VersionedParams(ctx, &types.QueryVersionedParamsRequest{CurrentTimestamp: blockTimeT2})
	require.NoError(t, err)
	require.Equal(t, &types.QueryVersionedParamsResponse{VersionedParams: versionedParamsT2}, responseT2)
	require.EqualValues(t, responseT2.GetVersionedParams().MaxSegmentSize, 2)
}
