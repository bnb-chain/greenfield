package keeper

import (
	"context"

    "github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) MockSetBucketPaymentAccount(goCtx context.Context,  msg *types.MsgMockSetBucketPaymentAccount) (*types.MsgMockSetBucketPaymentAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgMockSetBucketPaymentAccountResponse{}, nil
}
