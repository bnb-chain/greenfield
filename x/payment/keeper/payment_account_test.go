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

func createNPaymentAccount(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.PaymentAccount {
	items := make([]types.PaymentAccount, n)
	for i := range items {
		items[i].Addr = strconv.Itoa(i)

		keeper.SetPaymentAccount(ctx, items[i])
	}
	return items
}

func TestPaymentAccountGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNPaymentAccount(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetPaymentAccount(ctx,
			item.Addr,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestPaymentAccountRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNPaymentAccount(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemovePaymentAccount(ctx,
			item.Addr,
		)
		_, found := keeper.GetPaymentAccount(ctx,
			item.Addr,
		)
		require.False(t, found)
	}
}

func TestPaymentAccountGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNPaymentAccount(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllPaymentAccount(ctx)),
	)
}
