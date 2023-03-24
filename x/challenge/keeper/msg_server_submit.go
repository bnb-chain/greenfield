package keeper

import (
	"context"

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
	sp, found := k.SpKeeper.GetStorageProvider(ctx, spOperator)
	if !found {
		return nil, types.ErrUnknownSp
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return nil, types.ErrInvalidSpStatus
	}

	// check object & read needed data
	objectInfo, found := k.StorageKeeper.GetObjectInfo(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrUnknownObject
	}
	if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_SEALED {
		return nil, types.ErrInvalidObjectStatus
	}

	// check whether the sp stores the object info
	stored := false
	for _, sp := range objectInfo.GetSecondarySpAddresses() {
		if spOperator.Equals(sdk.MustAccAddressFromHex(sp)) {
			stored = true
			break
		}
	}
	if !stored {
		bucket, _ := k.StorageKeeper.GetBucketInfo(ctx, msg.BucketName)
		if !spOperator.Equals(sdk.MustAccAddressFromHex(bucket.GetPrimarySpAddress())) {
			return nil, types.ErrNotStoredOnSp
		}
	}

	// check sp recent slash
	if k.ExistsSlash(ctx, spOperator, objectInfo.Id) {
		return nil, types.ErrExistsRecentSlash
	}

	// generate redundancy index
	redundancyIndex := types.RedundancyIndexPrimary
	for i, sp := range objectInfo.GetSecondarySpAddresses() {
		if spOperator.Equals(sdk.MustAccAddressFromHex(sp)) {
			redundancyIndex = int32(i)
			break
		}
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
	expiredHeight := k.Keeper.ChallengeKeepAlivePeriod(ctx) + uint64(ctx.BlockHeight())
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
