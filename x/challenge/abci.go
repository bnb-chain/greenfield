package challenge

import (
	"fmt"
	"strings"

	k "github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
)

// coolingOffMultiplier is used to purge slashes. The purpose of it is for: if the cooling-off period is increase, e.g.
// by gov, then we still keep some recent slashes for the new parameter. However, if it is increased too large, then
// there is the possibility that some recent slashes for the new parameter will be gone, and we think this is acceptable.
const coolingOffMultiplier = 3

func BeginBlocker(ctx sdk.Context, keeper k.Keeper) {
	// reset count of challenge in current block to zero
	keeper.ResetChallengeCount(ctx)

	// delete expired challenges at this height
	events := make([]proto.Message, 0)
	expirePeriod := keeper.ChallengeExpirePeriod(ctx)
	height := uint64(ctx.BlockHeight()) - expirePeriod

	challenges := keeper.GetAllOngoingChallenge(ctx)
	for _, elem := range challenges {
		if elem.Height < height {
			events = append(events, &types.EventExpireChallenge{
				ChallengeId: elem.Id,
			})
			keeper.RemoveOngoingChallenge(ctx, elem.Id)
		}
	}

	_ = ctx.EventManager().EmitTypedEvents(events...)

	// delete too old slashes at this height
	coolingOff := keeper.SlashCoolingOffPeriod(ctx)
	height = uint64(ctx.BlockHeight()) - coolingOff*coolingOffMultiplier
	slashes := keeper.GetAllRecentSlash(ctx)
	for _, elem := range slashes {
		if elem.Height < height {
			keeper.RemoveRecentSlash(ctx, elem.Id)
		}
	}
}

func EndBlocker(ctx sdk.Context, keeper k.Keeper) {
	// emit new challenge events if more challenges are needed
	count := keeper.GetChallengeCount(ctx)
	needed := keeper.EventCountPerBlock(ctx)
	if count >= needed {
		return
	}

	events := make([]proto.Message, 0)                     // for events
	objectMap := make(map[string]struct{})                 // for de-duplication
	iteration, maxIteration := uint64(0), 2*(needed-count) // to prevent endless loop
	height := uint64(ctx.BlockHeight()) - keeper.SlashCoolingOffPeriod(ctx)
	for count < needed && iteration < maxIteration {
		iteration++

		// random object info
		objectKey := k.RandomObjectKey(ctx.BlockHeader().RandaoMix)
		objectInfo, found := keeper.StorageKeeper.GetObjectAfterKey(ctx, objectKey)
		if !found { // there is no object info yet, cannot generate challenges
			return
		}

		if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_IN_SERVICE {
			continue
		}

		// random redundancy index (sp address)
		var spOperatorAddress string
		secondarySpAddresses := objectInfo.SecondarySpAddresses

		redundancyIndex := k.RandomRedundancyIndex(ctx.BlockHeader().RandaoMix, uint64(len(secondarySpAddresses)+1))
		redundancyIndex--

		bucket, found := keeper.StorageKeeper.GetBucket(ctx, objectInfo.ObjectName)
		if !found {
			continue
		}
		if redundancyIndex == types.RedundancyIndexPrimary { // primary sp
			spOperatorAddress = bucket.PrimarySpAddress
		} else {
			spOperatorAddress = objectInfo.SecondarySpAddresses[redundancyIndex]
		}

		addr, err := sdk.AccAddressFromHexUnsafe(spOperatorAddress)
		if err != nil {
			continue
		}
		sp, found := keeper.SpKeeper.GetStorageProvider(ctx, addr)
		if !found || sp.Status != sptypes.STATUS_IN_SERVICE {
			continue
		}

		// check recent slash
		if keeper.ExistsSlash(ctx, height, strings.ToLower(spOperatorAddress), objectKey) {
			continue
		}

		// random segment/piece index
		segments := k.CalculateSegments(objectInfo.PayloadSize, keeper.StorageKeeper.MaxSegmentSize(ctx))
		segmentIndex := k.RandomSegmentIndex(ctx.BlockHeader().RandaoMix, segments)

		mapKey := fmt.Sprintf("%s-%d", spOperatorAddress, objectInfo.Id)
		if _, ok := objectMap[mapKey]; ok { // already generated for this pair
			continue
		}

		objectMap[mapKey] = struct{}{}
		challengeId, err := keeper.GetChallengeID(ctx)
		if err != nil {
			continue
		}
		challenge := types.Challenge{
			Id:                challengeId,
			SpOperatorAddress: spOperatorAddress,
			ObjectKey:         storagetypes.GetObjectKey(bucket.BucketName, objectInfo.ObjectName),
			SegmentIndex:      segmentIndex,
			Height:            uint64(ctx.BlockHeight()),
			ChallengerAddress: "",
		}
		keeper.SetOngoingChallenge(ctx, challenge)
		keeper.SetChallengeID(ctx, challengeId+1)
		events = append(events, &types.EventStartChallenge{
			ChallengeId:       challenge.Id,
			ObjectId:          objectInfo.Id.Uint64(),
			SegmentIndex:      segmentIndex,
			SpOperatorAddress: spOperatorAddress,
			RedundancyIndex:   redundancyIndex,
		})

		count++
	}
	_ = ctx.EventManager().EmitTypedEvents(events...)
}
