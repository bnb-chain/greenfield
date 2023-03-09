package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (k msgServer) Submit(goCtx context.Context, msg *types.MsgSubmit) (*types.MsgSubmitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spOperatorAddr, err := sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return nil, err
	}

	// check sp status
	sp, found := k.SpKeeper.GetStorageProvider(ctx, spOperatorAddr)
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
		if strings.EqualFold(spOperatorAddr.String(), sp) {
			stored = true
			break
		}
	}
	if !stored {
		bucket, _ := k.StorageKeeper.GetBucketInfo(ctx, msg.BucketName)
		if !strings.EqualFold(spOperatorAddr.String(), bucket.GetPrimarySpAddress()) {
			return nil, types.ErrNotStoredOnSp
		}
	}

	// check sp recent slash
	if k.ExistsSlash(ctx, spOperatorAddr, objectInfo.Id) {
		return nil, types.ErrExistsRecentSlash
	}

	// generate redundancy index
	redundancyIndex := types.RedundancyIndexPrimary
	for i, sp := range objectInfo.GetSecondarySpAddresses() {
		if strings.EqualFold(spOperatorAddr.String(), sp) {
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

	challengeId := k.GetOngoingChallengeId(ctx)
	k.SetOngoingChallengeId(ctx, challengeId+1)
	k.IncrChallengeCountCurrentBlock(ctx)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventStartChallenge{
		ChallengeId:       challengeId,
		ObjectId:          objectInfo.Id,
		SegmentIndex:      segmentIndex,
		SpOperatorAddress: spOperatorAddr.String(),
		RedundancyIndex:   redundancyIndex,
		ChallengerAddress: msg.Challenger,
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitResponse{}, nil
}
