package storage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	k "github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
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

	deleted := uint64(0)
	reachMax := false

	blockHeight := uint64(ctx.BlockHeight())
	// delete objects
	discontinueObjects := keeper.GetDiscontinueObjects(ctx, blockHeight)
	if len(discontinueObjects) == 0 {
		return
	}
	for operator, objectIds := range discontinueObjects {
		operatorAcc := sdk.AccAddress(operator)
		if reachMax {
			// move the deletion to the next height, and insert to the beginning
			keeper.InsertDiscontinueObjects(ctx, blockHeight+1, operatorAcc, objectIds.Id, false)
			continue
		}

		moveToNextHeight := make([]types.Uint, 0)
		for _, objectId := range objectIds.Id {
			if reachMax {
				moveToNextHeight = append(moveToNextHeight, objectId)
				continue
			}

			err := keeper.ForceDeleteObject(ctx, operatorAcc, objectId)
			if err != nil {
				ctx.Logger().Error("failed to delete object", "height", blockHeight, "id", objectId, "err", err.Error())
			} else {
				deleted++
				if deleted >= deletionMax {
					reachMax = true
				}
			}
		}
		if len(moveToNextHeight) > 0 {
			// move the deletion to the next height, and insert to the beginning
			keeper.InsertDiscontinueObjects(ctx, blockHeight+1, operatorAcc, moveToNextHeight, false)
		}
	}
	keeper.ClearDiscontinueObjects(ctx, blockHeight)

	if reachMax {
		return
	}

	// delete buckets
	discontinueBuckets := keeper.GetDiscontinueBuckets(ctx, blockHeight)
	if len(discontinueBuckets) == 0 {
		return
	}
	for operator, bucketIds := range discontinueBuckets {
		operatorAcc := sdk.AccAddress(operator)
		if reachMax {
			// move the deletion to the next height, and insert to the beginning
			keeper.InsertDiscontinueBuckets(ctx, blockHeight+1, operatorAcc, bucketIds.Id, false)
			continue
		}

		moveToNextHeight := make([]types.Uint, 0)
		for _, bucketId := range bucketIds.Id {
			if reachMax {
				moveToNextHeight = append(moveToNextHeight, bucketId)
				continue
			}

			bucketDeleted, objectDeleted, err := keeper.ForceDeleteBucket(ctx, operatorAcc, bucketId, deletionMax-deleted)
			if err != nil {
				ctx.Logger().Error("failed to delete bucket", "height", blockHeight, "id", bucketId, "err", err.Error())
			}
			deleted = deleted + objectDeleted
			if deleted >= deletionMax {
				reachMax = true
			}

			if !bucketDeleted {
				moveToNextHeight = append(moveToNextHeight, bucketId)
			}
		}
		if len(moveToNextHeight) > 0 {
			// move the deletion to the next height, and insert to the beginning
			keeper.InsertDiscontinueBuckets(ctx, blockHeight+1, operatorAcc, moveToNextHeight, false)
		}
	}
	keeper.ClearDiscontinueBuckets(ctx, blockHeight)
}
