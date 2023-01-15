package keeper

import (
	"context"

    "github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) MockSealObject(goCtx context.Context,  msg *types.MsgMockSealObject) (*types.MsgMockSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgMockSealObjectResponse{}, nil
}
