package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// Submit handles user's request for submitting a challenge.
func (k msgServer) Submit(goCtx context.Context, msg *types.MsgSubmit) (*types.MsgSubmitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spOperator := sdk.MustAccAddressFromHex(msg.SpOperatorAddress)
	challenger := sdk.MustAccAddressFromHex(msg.Challenger)

	// check sp status
	bucketInfo, found := k.StorageKeeper.GetBucketInfo(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrUnknownBucketObject
	}
	sp := k.StorageKeeper.MustGetPrimarySPForBucket(ctx, bucketInfo)
	if sp.Status != sptypes.STATUS_IN_SERVICE && sp.Status != sptypes.STATUS_GRACEFUL_EXITING {
		return nil, types.ErrInvalidSpStatus
	}

	// check object & read needed data
	objectInfo, found := k.StorageKeeper.GetObjectInfo(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrUnknownBucketObject
	}
	if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_SEALED {
		return nil, types.ErrInvalidObjectStatus
	}

	// check whether the sp stores the object info, generate redundancy index
	stored := false
	redundancyIndex := types.RedundancyIndexPrimary

	if spOperator.Equals(sdk.MustAccAddressFromHex(sp.OperatorAddress)) {
		stored = true
	}

	if !stored {
		gvg, found := k.StorageKeeper.GetObjectGVG(ctx, bucketInfo.Id, objectInfo.LocalVirtualGroupId)
		if !found {
			return nil, errors.Wrapf(types.ErrCannotFindGVG, "no GVG binding for LVG: %d", objectInfo.LocalVirtualGroupId)
		}

		// check secondary sp
		for i, spId := range gvg.SecondarySpIds {
			tmpSp, found := k.SpKeeper.GetStorageProvider(ctx, spId)
			if !found {
				return nil, errors.Wrapf(types.ErrUnknownSp, "cannot find storage provider: %d", spId)
			}
			if spOperator.Equals(sdk.MustAccAddressFromHex(tmpSp.OperatorAddress)) {
				redundancyIndex = int32(i)
				stored = true
				break
			}
		}
	}
	if !stored {
		return nil, types.ErrNotStoredOnSp
	}

	// check sp recent slash
	if k.ExistsSlash(ctx, sp.Id, objectInfo.Id) {
		return nil, types.ErrExistsRecentSlash
	}

	// generate segment index
	segmentIndex := msg.SegmentIndex
	segments := CalculateSegments(objectInfo.PayloadSize, k.Keeper.StorageKeeper.MaxSegmentSize(ctx))
	if msg.RandomIndex {
		segmentIndex = RandomSegmentIndex(ctx.BlockHeader().RandaoMix, segments)
	} else {
		if uint64(segmentIndex) > segments-1 {
			return nil, types.ErrInvalidSegmentIndex
		}
	}

	k.IncrChallengeCountCurrentBlock(ctx)
	challengeId := k.GetChallengeId(ctx) + 1
	expiredHeight := k.Keeper.GetParams(ctx).ChallengeKeepAlivePeriod + uint64(ctx.BlockHeight())
	k.SaveChallenge(ctx, types.Challenge{
		Id:            challengeId,
		ExpiredHeight: expiredHeight,
	})

	if err := ctx.EventManager().EmitTypedEvents(&types.EventStartChallenge{
		ChallengeId:       challengeId,
		ObjectId:          objectInfo.Id,
		SegmentIndex:      segmentIndex,
		SpOperatorAddress: spOperator.String(),
		RedundancyIndex:   redundancyIndex,
		ChallengerAddress: challenger.String(),
		ExpiredHeight:     expiredHeight,
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitResponse{}, nil
}
