package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/testutil/nullify"
	"github.com/bnb-chain/bfs/x/payment/keeper"
	"github.com/bnb-chain/bfs/x/payment/types"
)

func createTestBnbPrice(keeper *keeper.Keeper, ctx sdk.Context) types.BnbPrice {
	item := types.BnbPrice{}
	keeper.SetBnbPrice(ctx, item)
	return item
}

func TestBnbPriceGet(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	item := createTestBnbPrice(k, ctx)
	rst, found := k.GetBnbPrice(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestBnbPriceRemove(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	createTestBnbPrice(k, ctx)
	k.RemoveBnbPrice(ctx)
	_, found := k.GetBnbPrice(ctx)
	require.False(t, found)
}

func TestGetBNBPrice(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	k.SubmitBNBPrice(ctx, 1000, 1000)
	k.SubmitBNBPrice(ctx, 1234, 1234)
	k.SubmitBNBPrice(ctx, 2345, 2345)
}
