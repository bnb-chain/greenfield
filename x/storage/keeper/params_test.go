package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}

func GetParamsWithTimestamp(k *keeper.Keeper, ctx sdk.Context, ts int64) (time int64) {
	params, _ := k.GetParamsWithTimestamp(ctx, ts)
}

func TestMultiVersiontParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()
	beginTs := time.Now().Unix()

	t.Logf("beginTs time %d\n", beginTs)

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
	// default params
	require.EqualValues(t, GetParamsWithTimestamp(k, ctx, beginTs), 0)
	require.EqualValues(t, GetParamsWithTimestamp(k, ctx, beginTs+20), beginTs+10)
	require.EqualValues(t, GetParamsWithTimestamp(k, ctx, beginTs+30), beginTs+20)

}
