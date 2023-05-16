package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) VersionedParams(c context.Context, req *types.QueryVersionedParamsRequest) (*types.QueryVersionedParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	versionedParams, err := k.GetVersionedParamsWithTs(ctx, req.GetTimestamp())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryVersionedParamsResponse{VersionedParams: versionedParams}, nil
}
