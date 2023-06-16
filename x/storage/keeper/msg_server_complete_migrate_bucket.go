package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CompleteMigrateBucket(goCtx context.Context, msg *types.MsgCompleteMigrateBucket) (*types.MsgCompleteMigrateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgCompleteMigrateBucketResponse{}, nil
}
