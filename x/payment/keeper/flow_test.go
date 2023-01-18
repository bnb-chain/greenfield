package keeper_test

import (
	sdkmath "cosmossdk.io/math"
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

func createNFlow(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Flow {
	items := make([]types.Flow, n)
	for i := range items {
		items[i].From = strconv.Itoa(i)
		items[i].To = strconv.Itoa(i)
		items[i].Rate = sdkmath.NewInt(int64(i))

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

func TestUpdateFlow(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	flow := types.Flow{
		From: "from",
		To:   "to",
		Rate: sdkmath.NewInt(100),
	}
	_, found := keeper.GetFlow(ctx, flow.From, flow.To)
	require.False(t, found)
	err := keeper.ApplyFlow(ctx, flow)
	require.NoError(t, err)
	rst, found := keeper.GetFlow(ctx, flow.From, flow.To)
	require.True(t, found)
	t.Logf("flow: %+v", flow)
	require.Equal(t, flow, rst)
	// test update
	flow2 := types.Flow{
		From: "from",
		To:   "to",
		Rate: sdkmath.NewInt(200),
	}
	err = keeper.ApplyFlow(ctx, flow2)
	require.NoError(t, err)
	rst, found = keeper.GetFlow(ctx, flow.From, flow.To)
	require.True(t, found)
	t.Logf("after update flow2: %+v", rst)
	require.Equal(t, flow2.Rate.Add(flow.Rate), rst.Rate)
	// test update negative
	flow3 := types.Flow{
		From: "from",
		To:   "to",
		Rate: sdkmath.NewInt(-400),
	}
	err = keeper.ApplyFlow(ctx, flow3)
	t.Logf("after update flow3: %+v", err)
	require.Error(t, err)
}
