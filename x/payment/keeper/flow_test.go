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

func createNFlow(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Flow {
	items := make([]types.Flow, n)
	for i := range items {
		items[i].From = strconv.Itoa(i)
        items[i].To = strconv.Itoa(i)
        
		keeper.SetFlow(ctx, items[i])
	}
	return items
}

func TestFlowGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNFlow(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetFlow(ctx,
		    item.From,
            item.To,
            
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestFlowRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNFlow(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveFlow(ctx,
		    item.From,
            item.To,
            
		)
		_, found := keeper.GetFlow(ctx,
		    item.From,
            item.To,
            
		)
		require.False(t, found)
	}
}

func TestFlowGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNFlow(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllFlow(ctx)),
	)
}
