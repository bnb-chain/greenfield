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

func createNPaymentAccountCount(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.PaymentAccountCount {
	items := make([]types.PaymentAccountCount, n)
	for i := range items {
		items[i].Owner = strconv.Itoa(i)

		keeper.SetPaymentAccountCount(ctx, items[i])
	}
	return items
}

func TestPaymentAccountCountGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNPaymentAccountCount(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetPaymentAccountCount(ctx,
			item.Owner,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestPaymentAccountCountRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNPaymentAccountCount(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemovePaymentAccountCount(ctx,
			item.Owner,
		)
		_, found := keeper.GetPaymentAccountCount(ctx,
			item.Owner,
		)
		require.False(t, found)
	}
}

func TestPaymentAccountCountGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNPaymentAccountCount(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllPaymentAccountCount(ctx)),
	)
}
