package keeper

import (
	"context"

    "github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) MockUpdateBucketReadPacket(goCtx context.Context,  msg *types.MsgMockUpdateBucketReadPacket) (*types.MsgMockUpdateBucketReadPacketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgMockUpdateBucketReadPacketResponse{}, nil
}
