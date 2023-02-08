package keeper_test

import (
	"testing"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func createSlash(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Slash {
	items := make([]types.Slash, n)
	for i := range items {
		items[i].Id = keeper.AppendRecentSlash(ctx, items[i])
	}
	return items
}

func TestRecentSlashGet(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	items := createSlash(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetRecentSlash(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestRecentSlashRemove(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	items := createSlash(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveRecentSlash(ctx, item.Id)
		_, found := keeper.GetRecentSlash(ctx, item.Id)
		require.False(t, found)
	}
}

func TestRecentSlashGetAll(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	items := createSlash(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllRecentSlash(ctx)),
	)
}

func TestRecentSlashCount(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	items := createSlash(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetRecentSlashCount(ctx))
}
