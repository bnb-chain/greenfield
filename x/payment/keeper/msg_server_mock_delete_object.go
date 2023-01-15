package keeper

import (
	"context"

    "github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) MockDeleteObject(goCtx context.Context,  msg *types.MsgMockDeleteObject) (*types.MsgMockDeleteObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgMockDeleteObjectResponse{}, nil
}
