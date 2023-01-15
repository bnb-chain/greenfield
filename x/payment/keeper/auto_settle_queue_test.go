package keeper_test

import (
	"strconv"
	"testing"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/testutil/nullify"
	"github.com/bnb-chain/bfs/x/payment/keeper"
	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNAutoSettleQueue(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.AutoSettleQueue {
	items := make([]types.AutoSettleQueue, n)
	for i := range items {
		items[i].Timestamp = int64(int32(i))
		items[i].Addr = strconv.Itoa(i)

		keeper.SetAutoSettleQueue(ctx, items[i])
	}
	return items
}

func TestAutoSettleQueueGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNAutoSettleQueue(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetAutoSettleQueue(ctx,
			item.Timestamp,
			item.Addr,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestAutoSettleQueueRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNAutoSettleQueue(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveAutoSettleQueue(ctx,
			item.Timestamp,
			item.Addr,
		)
		_, found := keeper.GetAutoSettleQueue(ctx,
			item.Timestamp,
			item.Addr,
		)
		require.False(t, found)
	}
}

func TestAutoSettleQueueGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNAutoSettleQueue(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllAutoSettleQueue(ctx)),
	)
}
