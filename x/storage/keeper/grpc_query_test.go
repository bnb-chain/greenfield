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
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := k.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestVersionedParamsQuery(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()
	params.VersionedParams.MaxSegmentSize = 1
	blockTimeT1 := ctx.BlockTime().Unix()
	paramsT1 := params
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	blockTimeT2 := ctx.BlockTime().Unix()
	params.VersionedParams.MaxSegmentSize = 2
	paramsT2 := params
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	responseT1, err := k.QueryParamsByTimestamp(ctx, &types.QueryParamsByTimestampRequest{Timestamp: blockTimeT1})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsByTimestampResponse{Params: paramsT1}, responseT1)
	getParams := responseT1.GetParams()
	require.EqualValues(t, getParams.GetMaxSegmentSize(), 1)

	responseT2, err := k.QueryParamsByTimestamp(ctx, &types.QueryParamsByTimestampRequest{Timestamp: blockTimeT2})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsByTimestampResponse{Params: paramsT2}, responseT2)
	p := responseT2.GetParams()
	require.EqualValues(t, p.GetMaxSegmentSize(), 2)

	responseT3, err := k.QueryParamsByTimestamp(ctx, &types.QueryParamsByTimestampRequest{Timestamp: 0})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsByTimestampResponse{Params: paramsT2}, responseT3)
	p = responseT2.GetParams()
	require.EqualValues(t, p.GetMaxSegmentSize(), 2)
}
