package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CompleteSwapOut(goCtx context.Context, msg *types.MsgCompleteSwapOut) (*types.MsgCompleteSwapOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgCompleteSwapOutResponse{}, nil
}
