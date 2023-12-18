package challenge

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	k "github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func BeginBlocker(ctx sdk.Context, keeper k.Keeper) {
	blockHeight := uint64(ctx.BlockHeight())
	// delete expired challenges at this height
	keeper.RemoveChallengeUntil(ctx, blockHeight)

	params := keeper.GetParams(ctx)
	// delete too old slashes at this height
	coolingOffPeriod := params.SlashCoolingOffPeriod
	if blockHeight > coolingOffPeriod {
		height := blockHeight - coolingOffPeriod
		keeper.RemoveSlashUntil(ctx, height)
	}

	// delete storage provider slash amount records
	if blockHeight > 0 && blockHeight%params.SpSlashCountingWindow == 0 {
		keeper.ClearSpSlashAmount(ctx)
	}
}

func EndBlocker(ctx sdk.Context, keeper k.Keeper) {
	count := keeper.GetChallengeCountCurrentBlock(ctx)

	params := keeper.GetParams(ctx)
	needed := params.ChallengeCountPerBlock
	if count >= needed {
		return
	}

	objectCount := keeper.StorageKeeper.GetObjectInfoCount(ctx)
	if objectCount.IsZero() {
		return
	}

	expiredHeight := params.ChallengeKeepAlivePeriod + uint64(ctx.BlockHeight())

	events := make([]proto.Message, 0)                      // for events
	objectMap := make(map[string]struct{})                  // for de-duplication
	iteration, maxIteration := uint64(0), 10*(needed-count) // to prevent endless loop
	for count < needed && iteration < maxIteration {
		iteration++
		seed := k.SeedFromRandaoMix(ctx.BlockHeader().RandaoMix, iteration)

		// random object info
		objectId := k.RandomObjectId(seed, objectCount)
		objectInfo, found := keeper.StorageKeeper.GetObjectInfoById(ctx, objectId)
		if !found || objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_SEALED {
			continue
		}

		// skip empty object
		if objectInfo.PayloadSize == 0 {
			continue
		}

		// random redundancy index (sp address)
		var spOperatorId uint32

		bucket, found := keeper.StorageKeeper.GetBucketInfo(ctx, objectInfo.BucketName)
		if !found {
			continue
		}
		gvg, found := keeper.StorageKeeper.GetObjectGVG(ctx, bucket.Id, objectInfo.LocalVirtualGroupId)
		if !found {
			continue
		}
		redundancyIndex := k.RandomRedundancyIndex(seed, uint64(len(gvg.SecondarySpIds)+1))
		if redundancyIndex == types.RedundancyIndexPrimary { // primary sp
			spOperatorId = gvg.PrimarySpId
		} else {
			spOperatorId = gvg.SecondarySpIds[redundancyIndex]
		}

		sp, found := keeper.SpKeeper.GetStorageProvider(ctx, spOperatorId)
		if !found {
			continue
		}
		if sp.Status != sptypes.STATUS_IN_SERVICE && sp.Status != sptypes.STATUS_GRACEFUL_EXITING && sp.Status != sptypes.STATUS_FORCED_EXITING {
			continue
		}

		mapKey := fmt.Sprintf("%d-%s", spOperatorId, objectInfo.Id.String())
		if _, ok := objectMap[mapKey]; ok { // already generated for this pair
			continue
		}

		// check recent slash
		if keeper.ExistsSlash(ctx, sp.Id, objectInfo.Id) {
			continue
		}

		// random segment/piece index
		segmentSize, err := keeper.StorageKeeper.MaxSegmentSize(ctx, objectInfo.CreateAt)
		if err != nil {
			ctx.Logger().Error("fail to get segment size", "timestamp", objectInfo.CreateAt,
				"err", err.Error())
			continue
		}
		segments := k.CalculateSegments(objectInfo.PayloadSize, segmentSize)
		segmentIndex := k.RandomSegmentIndex(seed, segments)

		objectMap[mapKey] = struct{}{}

		challengeId := keeper.GetChallengeId(ctx) + 1
		keeper.SaveChallenge(ctx, types.Challenge{
			Id:            challengeId,
			ExpiredHeight: expiredHeight,
		})
		events = append(events, &types.EventStartChallenge{
			ChallengeId:       challengeId,
			ObjectId:          objectInfo.Id,
			SegmentIndex:      segmentIndex,
			SpId:              sp.Id,
			SpOperatorAddress: sp.OperatorAddress,
			RedundancyIndex:   redundancyIndex,
			ChallengerAddress: "",
			ExpiredHeight:     expiredHeight,
		})

		count++
	}
	err := ctx.EventManager().EmitTypedEvents(events...)
	if err != nil {
		ctx.Logger().Error("failed to emit challenge events", "err", err.Error())
	}
}
