package keeper

import (
	"context"

    "github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) MockCreateBucket(goCtx context.Context,  msg *types.MsgMockCreateBucket) (*types.MsgMockCreateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgMockCreateBucketResponse{}, nil
}
