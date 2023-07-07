package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) OutFlows(c context.Context, req *types.QueryOutFlowsRequest) (*types.QueryOutFlowsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	account, err := sdk.AccAddressFromHexUnsafe(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account")
	}

	return &types.QueryOutFlowsResponse{OutFlows: k.GetOutFlows(ctx, account)}, nil
}
