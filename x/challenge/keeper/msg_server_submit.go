package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Submit(goCtx context.Context, msg *types.MsgSubmit) (*types.MsgSubmitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spOperatorAddress, err := sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return nil, err
	}

	// check sp status
	sp, found := k.SpKeeper.GetStorageProvider(ctx, spOperatorAddress)
	if !found {
		return nil, types.ErrUnknownSp
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return nil, types.ErrInvalidSpStatus
	}

	// check sp recent slash

	// check object & read needed data
	objectInfo, found := k.StorageKeeper.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrUnknownObject
	}
	if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_IN_SERVICE {
		return nil, types.ErrInvalidObjectStatus
	}

	objectKey := storagetypes.GetObjectKey(msg.BucketName, msg.ObjectName)
	objectId := objectInfo.Id

	index := msg.Index
	if msg.RandomIndex {
		//TODO: random index

	}

	challengeId, err := k.GetChallengeID(ctx)
	if err != nil {
		return nil, err
	}
	challenge := types.Challenge{
		Id:                challengeId,
		SpOperatorAddress: msg.SpOperatorAddress,
		ObjectKey:         objectKey,
		Index:             msg.Index,
		Height:            uint64(ctx.BlockHeight()),
		ChallengerAddress: msg.Creator,
	}

	k.SetOngoingChallenge(ctx, challenge)
	k.SetChallengeID(ctx, challengeId+1)
	k.IncrChallengeCount(ctx)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventStartChallenge{
		ChallengeId:       challengeId,
		SpOperatorAddress: msg.SpOperatorAddress,
		ObjectId:          objectId.Uint64(),
		Index:             index,
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitResponse{}, nil
}
