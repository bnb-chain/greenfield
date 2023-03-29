package storage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	k "github.com/bnb-chain/greenfield/x/storage/keeper"
)

func BeginBlocker(ctx sdk.Context, keeper k.Keeper) {
	blockHeight := uint64(ctx.BlockHeight())
	countingWindow := keeper.DiscontinueCountingWindow(ctx)
	if blockHeight > 0 && countingWindow > 0 && blockHeight%countingWindow == 0 {
		keeper.ClearDiscontinueObjectCount(ctx)
		keeper.ClearDiscontinueBucketCount(ctx)
	}
}

func EndBlocker(ctx sdk.Context, keeper k.Keeper) {
	deletionMax := keeper.DiscontinueDeletionMax(ctx)
	if deletionMax == 0 {
		return
	}

	blockTime := ctx.BlockTime().Unix()
	// delete objects
	deleted := keeper.DeleteDiscontinueObjectsUntil(ctx, blockTime, deletionMax)

	if deleted >= deletionMax {
		return
	}

	// delete buckets
	_ = keeper.DeleteDiscontinueBucketsUntil(ctx, blockTime, deletionMax-deleted)
}
