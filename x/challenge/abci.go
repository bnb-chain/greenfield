package challenge

import (
	"fmt"

	"github.com/bnb-chain/greenfield/x/challenge/keeper"
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

func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	// reset count of challenge in current block to zero
	keeper.ResetChallengeCount(ctx)

	// delete expired challenges at this height
	events := make([]proto.Message, 0)
	height := uint64(ctx.BlockHeight())
	expirePeriod := keeper.ChallengeExpirePeriod(ctx)
	minHeight := height
	if height > expirePeriod {
		minHeight = height - expirePeriod
	}
	challenges := keeper.GetAllOngoingChallenge(ctx)
	for _, elem := range challenges {
		if elem.Height < minHeight {
			events = append(events, &types.EventExpireChallenge{
				ChallengeId: elem.Id,
			})
			keeper.RemoveOngoingChallenge(ctx, elem.Id)
		}
	}

	_ = ctx.EventManager().EmitTypedEvents(events...)

	// delete too old slashes at this height
	coolingOff := keeper.SlashCoolingOffPeriod(ctx)
	minHeight = height - coolingOff*coolingOffMultiplier
	slashes := keeper.GetAllRecentSlash(ctx)
	for _, elem := range slashes {
		if elem.Height < minHeight {
			keeper.RemoveRecentSlash(ctx, elem.Id)
		}
	}
}

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	// emit new challenge events if more challenges are needed
	count := keeper.GetChallengeCount(ctx)
	needed := keeper.EventCountPerBlock(ctx)
	events := make([]proto.Message, 0)

	if count >= needed {
		return
	}

	objectMap := make(map[string]struct{})                 // for de-duplication
	iteration, maxIteration := uint64(0), 2*(needed-count) // to prevent endless loop
	for count < needed && iteration < maxIteration {
		iteration++

		// TODO: random object to challenge
		randomObjectKey := []byte{}
		objectInfo, found := keeper.StorageKeeper.GetObjectAfterKey(ctx, randomObjectKey)
		if !found { // there is no object info yet
			return
		}

		if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_IN_SERVICE {
			continue
		}

		// random index
		randomSegmentIndex := uint32(1)

		// random sp address
		bucket, _ := keeper.StorageKeeper.GetBucket(ctx, objectInfo.ObjectName)
		randomSpOperatorAddress := ""
		var redundancyIndex int32
		if randomSegmentIndex == 0 { //primary sp
			randomSpOperatorAddress = bucket.PrimarySpAddress
			redundancyIndex = int32(-1)
		} else { //secondary sp
			secondarySpAddresses := objectInfo.SecondarySpAddresses
			randomSpOperatorAddress = secondarySpAddresses[randomSegmentIndex-1]
			redundancyIndex = int32(randomSegmentIndex - 1)
		}
		addr, _ := sdk.AccAddressFromHexUnsafe(randomSpOperatorAddress)
		sp, found := keeper.SpKeeper.GetStorageProvider(ctx, addr)
		if !found || sp.Status != sptypes.STATUS_IN_SERVICE {
			continue
		}

		mapKey := fmt.Sprintf("%s-%d", randomSpOperatorAddress, objectInfo.Id)
		if _, ok := objectMap[mapKey]; ok {
			continue
		}

		objectMap[mapKey] = struct{}{}
		challengeId, _ := keeper.GetChallengeID(ctx)
		objectKey := storagetypes.GetObjectKey(bucket.BucketName, objectInfo.ObjectName)
		challenge := types.Challenge{
			Id:                challengeId,
			SpOperatorAddress: randomSpOperatorAddress,
			ObjectKey:         objectKey,
			SegmentIndex:      randomSegmentIndex,
			Height:            uint64(ctx.BlockHeight()),
			ChallengerAddress: "",
		}
		keeper.SetOngoingChallenge(ctx, challenge)
		keeper.SetChallengeID(ctx, challengeId+1)
		events = append(events, &types.EventStartChallenge{
			ChallengeId:       challenge.Id,
			ObjectId:          objectInfo.Id.Uint64(),
			SegmentIndex:      randomSegmentIndex,
			SpOperatorAddress: randomSpOperatorAddress,
			RedundancyIndex:   redundancyIndex,
		})

		count++
	}
	_ = ctx.EventManager().EmitTypedEvents(events...)
}
