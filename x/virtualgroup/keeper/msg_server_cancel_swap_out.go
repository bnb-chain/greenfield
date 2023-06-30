package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CancelSwapOut(goCtx context.Context, msg *types.MsgCancelSwapOut) (*types.MsgCancelSwapOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgCancelSwapOutResponse{}, nil
}
