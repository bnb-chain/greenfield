package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func createSlash(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Slash {
	items := make([]types.Slash, n)
	for i := range items {
		items[i].ObjectId = uint64(i)
		items[i].Height = uint64(i)
		items[i].SpOperatorAddress = fmt.Sprintf("addr-%d", i)
		keeper.SaveSlash(ctx, items[i])
	}
	return items
}

func TestRecentSlashRemove(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	items := createSlash(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveSlashUntil(ctx, item.Height)
		found := keeper.ExistsSlash(ctx, item.SpOperatorAddress, item.ObjectId)
		require.False(t, found)
	}
}
