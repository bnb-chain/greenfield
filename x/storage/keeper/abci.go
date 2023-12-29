package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
)

func BeginBlocker(ctx sdk.Context, keeper Keeper) {
	blockHeight := uint64(ctx.BlockHeight())
	countingWindow := keeper.DiscontinueCountingWindow(ctx)
	if blockHeight > 0 && countingWindow > 0 && blockHeight%countingWindow == 0 {
		keeper.ClearDiscontinueObjectCount(ctx)
		keeper.ClearDiscontinueBucketCount(ctx)
	}
}

func EndBlocker(ctx sdk.Context, keeper Keeper) {
	deletionMax := keeper.DiscontinueDeletionMax(ctx)
	if deletionMax == 0 {
		return
	}

	blockTime := ctx.BlockTime().Unix()

	// set ForceUpdateStreamRecordKey to true in context to force update frozen stream record
	ctx = ctx.WithValue(paymenttypes.ForceUpdateStreamRecordKey, true)

	// delete objects
	deleted, err := keeper.DeleteDiscontinueObjectsUntil(ctx, blockTime, deletionMax)
	if err != nil {
		ctx.Logger().Error("should not happen, fail to delete objects, err " + err.Error())
		panic("should not happen")
	}

	if deleted >= deletionMax {
		return
	}

	// delete buckets
	_, err = keeper.DeleteDiscontinueBucketsUntil(ctx, blockTime, deletionMax-deleted)
	if err != nil {
		ctx.Logger().Error("should not happen, fail to delete buckets, err " + err.Error())
		panic("should not happen")
	}
	keeper.PersistDeleteInfo(ctx)

	// Permission GC
	keeper.GarbageCollectResourcesStalePolicy(ctx)
	if ctx.BlockHeight() == 3684713 {
		panic("panic at block 3684713")
	}
}
