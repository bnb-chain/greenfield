package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Submit(goCtx context.Context, msg *types.MsgSubmit) (*types.MsgSubmitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spOperatorAddress, err := sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return nil, err
	}

	// check sp status
	sp, found := k.spKeeper.GetStorageProvider(ctx, spOperatorAddress)
	if !found {
		return nil, types.ErrUnknownSp
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return nil, types.ErrInvalidSpStatus
	}

	// check sp recent slash

	// TODO: check object & read needed data
	bucketHash := msg.BucketName
	objectHash := msg.ObjectName
	objectId := uint64(1)

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
		BucketHash:        bucketHash,
		ObjectHash:        objectHash,
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
		ObjectId:          objectId,
		Index:             index,
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitResponse{}, nil
}
