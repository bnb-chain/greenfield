package storage

import (
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
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
	// set ForceUpdateFrozenStreamRecordKey to true in context to force update frozen stream record
	ctx = ctx.WithValue(paymenttypes.ForceUpdateFrozenStreamRecordKey, true)
	// delete objects
	deleted, err := keeper.DeleteDiscontinueObjectsUntil(ctx, blockTime, deletionMax)
	if err != nil {
		panic("fail to delete objects, err " + err.Error())
	}

	if deleted >= deletionMax {
		return
	}

	// delete buckets
	_, err = keeper.DeleteDiscontinueBucketsUntil(ctx, blockTime, deletionMax-deleted)
	if err != nil {
		panic("fail to delete buckets, err " + err.Error())
	}
	keeper.PersistDeleteInfo(ctx)

	// Permission GC
	keeper.GarbageCollectResourcesStalePolicy(ctx)
}
