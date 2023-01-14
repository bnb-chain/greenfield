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

func createNStreamRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.StreamRecord {
	items := make([]types.StreamRecord, n)
	for i := range items {
		items[i].Account = strconv.Itoa(i)
		items[i].NetflowRate = sdkmath.ZeroInt()
		items[i].StaticBalance = sdkmath.ZeroInt()
		items[i].BufferBalance = sdkmath.ZeroInt()

		keeper.SetStreamRecord(ctx, items[i])
	}
	return items
}

func TestStreamRecordGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNStreamRecord(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetStreamRecord(ctx,
			item.Account,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestStreamRecordRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNStreamRecord(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveStreamRecord(ctx,
			item.Account,
		)
		_, found := keeper.GetStreamRecord(ctx,
			item.Account,
		)
		require.False(t, found)
	}
}

func TestStreamRecordGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNStreamRecord(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllStreamRecord(ctx)),
	)
}
