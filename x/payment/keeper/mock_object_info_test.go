package keeper_test

import (
	"strconv"
	"testing"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/payment/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNMockObjectInfo(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.MockObjectInfo {
	items := make([]types.MockObjectInfo, n)
	for i := range items {
		items[i].BucketName = strconv.Itoa(i)
		items[i].ObjectName = strconv.Itoa(i)

		keeper.SetMockObjectInfo(ctx, items[i])
	}
	return items
}

func TestMockObjectInfoGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNMockObjectInfo(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetMockObjectInfo(ctx,
			item.BucketName,
			item.ObjectName,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestMockObjectInfoRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNMockObjectInfo(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveMockObjectInfo(ctx,
			item.BucketName,
			item.ObjectName,
		)
		_, found := keeper.GetMockObjectInfo(ctx,
			item.BucketName,
			item.ObjectName,
		)
		require.False(t, found)
	}
}

func TestMockObjectInfoGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNMockObjectInfo(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllMockObjectInfo(ctx)),
	)
}
