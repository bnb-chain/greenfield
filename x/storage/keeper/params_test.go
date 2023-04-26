package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}

func TestMultiVersiontParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()
	beginTs := time.Now().Unix()

	t.Logf("beginTs time %d\n", beginTs)

	params.CreateTimestamp = beginTs + 10
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	params.CreateTimestamp = beginTs + 20
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	params.CreateTimestamp = beginTs + 30
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
	// default params
	require.EqualValues(t, k.GetParamsWithTimestamp(ctx, beginTs).CreateTimestamp, 0)
	require.EqualValues(t, k.GetParamsWithTimestamp(ctx, beginTs+10).CreateTimestamp, beginTs+10)
	require.EqualValues(t, k.GetParamsWithTimestamp(ctx, beginTs+20).CreateTimestamp, beginTs+20)

}
