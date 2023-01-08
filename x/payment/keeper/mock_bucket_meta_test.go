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

func createNMockBucketMeta(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.MockBucketMeta {
	items := make([]types.MockBucketMeta, n)
	for i := range items {
		items[i].BucketName = strconv.Itoa(i)
        
		keeper.SetMockBucketMeta(ctx, items[i])
	}
	return items
}

func TestMockBucketMetaGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNMockBucketMeta(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetMockBucketMeta(ctx,
		    item.BucketName,
            
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestMockBucketMetaRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNMockBucketMeta(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveMockBucketMeta(ctx,
		    item.BucketName,
            
		)
		_, found := keeper.GetMockBucketMeta(ctx,
		    item.BucketName,
            
		)
		require.False(t, found)
	}
}

func TestMockBucketMetaGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNMockBucketMeta(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllMockBucketMeta(ctx)),
	)
}
