package keeper_test

import (
	"strconv"
	"testing"

	"github.com/bnb-chain/bfs/x/payment/keeper"
	"github.com/bnb-chain/bfs/x/payment/types"
	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/testutil/nullify"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNAutoSettleRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.AutoSettleRecord {
	items := make([]types.AutoSettleRecord, n)
	for i := range items {
		items[i].Timestamp = int32(i)
        items[i].Addr = strconv.Itoa(i)
        
		keeper.SetAutoSettleRecord(ctx, items[i])
	}
	return items
}

func TestAutoSettleRecordGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNAutoSettleRecord(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetAutoSettleRecord(ctx,
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
func TestAutoSettleRecordRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNAutoSettleRecord(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveAutoSettleRecord(ctx,
		    item.Timestamp,
            item.Addr,
            
		)
		_, found := keeper.GetAutoSettleRecord(ctx,
		    item.Timestamp,
            item.Addr,
            
		)
		require.False(t, found)
	}
}

func TestAutoSettleRecordGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNAutoSettleRecord(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllAutoSettleRecord(ctx)),
	)
}
