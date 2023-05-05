package keeper_test

import (
	"fmt"
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

func GetParamsWithTimestamp(k *keeper.Keeper, ctx sdk.Context, ts int64) (val types.Params) {
	params, err := k.GetParamsWithTs(ctx, ts)
	if err != nil {
		fmt.Printf("GetParamsWithTimestamp err %s\n", err)
	}
	return params
}

func TestMultiVersiontParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()
	beginTs := time.Now().Unix()

	t.Logf("beginTs time %d\n", beginTs)

	blockTimeT1 := ctx.BlockTime().Unix()
	params.MaxSegmentSize = 1
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	blockTimeT2 := ctx.BlockTime().Unix()
	params.MaxSegmentSize = 2
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	blockTimeT3 := ctx.BlockTime().Unix()
	params.MaxSegmentSize = 3
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
	// default params
	require.EqualValues(t, GetParamsWithTimestamp(k, ctx, blockTimeT1+1).MaxSegmentSize, 1)
	require.EqualValues(t, GetParamsWithTimestamp(k, ctx, blockTimeT2+1).MaxSegmentSize, 2)
	require.EqualValues(t, GetParamsWithTimestamp(k, ctx, blockTimeT3+1).MaxSegmentSize, 3)

}
