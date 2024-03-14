package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) ParamsByTimestamp(c context.Context, req *types.QueryParamsByTimestampRequest) (*types.QueryParamsByTimestampResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	ts := req.GetTimestamp()
	if ts == 0 {
		ts = ctx.BlockTime().Unix() + 1
	}

	params := k.GetParams(ctx)
	versionedParams, err := k.GetVersionedParamsWithTs(ctx, ts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	params.VersionedParams = versionedParams

	return &types.QueryParamsByTimestampResponse{Params: params}, nil
}
