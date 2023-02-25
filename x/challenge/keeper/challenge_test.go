package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createChallenge(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Challenge {
	items := make([]types.Challenge, n)
	for i := range items {
		items[i].Id = uint64(i)

		keeper.SaveChallenge(ctx, items[i])
	}
	return items
}

func TestChallengeGet(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	items := createChallenge(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetChallenge(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestChallengeRemove(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	items := createChallenge(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveChallengeUntil(ctx, item.Id)
		_, found := keeper.GetChallenge(ctx, item.Id)
		require.False(t, found)
	}
}
