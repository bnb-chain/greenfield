package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func createSlash(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Slash {
	items := make([]types.Slash, n)
	for i := range items {
		items[i].ObjectId = sdkmath.NewUint(uint64(i))
		items[i].Height = uint64(i)
		items[i].SpId = uint32(i + 1)
		keeper.SaveSlash(ctx, items[i])
	}
	return items
}

func TestRecentSlashRemove(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	items := createSlash(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveSlashUntil(ctx, item.Height)
		found := keeper.ExistsSlash(ctx, item.SpId, item.ObjectId)
		require.False(t, found)
	}
}
