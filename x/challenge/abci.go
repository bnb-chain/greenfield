package challenge

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	k "github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func BeginBlocker(ctx sdk.Context, keeper k.Keeper) {
	blockHeight := uint64(ctx.BlockHeight())
	// delete expired challenges at this height
	keeper.RemoveChallengeUntil(ctx, blockHeight)

	// delete too old slashes at this height
	coolingOff := keeper.SlashCoolingOffPeriod(ctx)
	if blockHeight <= coolingOff {
		return
	}

	height := blockHeight - coolingOff
	keeper.RemoveSlashUntil(ctx, height)
}

func EndBlocker(ctx sdk.Context, keeper k.Keeper) {
	count := keeper.GetChallengeCountCurrentBlock(ctx)
	needed := keeper.ChallengeCountPerBlock(ctx)
	if count >= needed {
		return
	}

	objectCount := keeper.StorageKeeper.GetObjectInfoCount(ctx)
	if objectCount.IsZero() {
		return
	}

	segmentSize := keeper.StorageKeeper.MaxSegmentSize(ctx)
	expiredHeight := keeper.ChallengeKeepAlivePeriod(ctx) + uint64(ctx.BlockHeight())

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

		// random redundancy index (sp address)
		var spOperatorAddress string
		secondarySpAddresses := objectInfo.SecondarySpAddresses

		redundancyIndex := k.RandomRedundancyIndex(seed, uint64(len(secondarySpAddresses)+1))
		if redundancyIndex == types.RedundancyIndexPrimary { // primary sp
			bucket, found := keeper.StorageKeeper.GetBucketInfo(ctx, objectInfo.BucketName)
			if !found {
				continue
			}
			spOperatorAddress = bucket.PrimarySpAddress
		} else {
			spOperatorAddress = objectInfo.SecondarySpAddresses[redundancyIndex]
		}

		spOperatorAddr, err := sdk.AccAddressFromHexUnsafe(spOperatorAddress)
		if err != nil {
			continue
		}
		sp, found := keeper.SpKeeper.GetStorageProvider(ctx, spOperatorAddr)
		if !found || sp.Status != sptypes.STATUS_IN_SERVICE {
			continue
		}

		mapKey := fmt.Sprintf("%s-%s", spOperatorAddress, objectInfo.Id.String())
		if _, ok := objectMap[mapKey]; ok { // already generated for this pair
			continue
		}

		// check recent slash
		if keeper.ExistsSlash(ctx, spOperatorAddr, objectInfo.Id) {
			continue
		}

		// skip empty object
		if objectInfo.PayloadSize == 0 {
			continue
		}

		// random segment/piece index
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
			SpOperatorAddress: spOperatorAddress,
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
