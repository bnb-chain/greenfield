package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	types2 "github.com/bnb-chain/greenfield/x/sp/types"
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

func (k Keeper) QueryParamsByTimestamp(c context.Context, req *types.QueryParamsByTimestampRequest) (*types.QueryParamsByTimestampResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	ts := req.GetTimestamp()
	if ts == 0 {
		ts = ctx.BlockTime().Unix()
	}

	params := k.GetParams(ctx)
	versionedParams, err := k.GetVersionedParamsWithTs(ctx, ts)
	params.VersionedParams = versionedParams
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsByTimestampResponse{Params: params}, nil
}

func (k Keeper) QueryLockFee(c context.Context, req *types.QueryLockFeeRequest) (*types.QueryLockFeeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	createAt := req.GetCreateAt()
	if createAt == 0 {
		createAt = ctx.BlockTime().Unix()
	}

	primaryAcc, err := sdk.AccAddressFromHexUnsafe(req.PrimarySpAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid primary storage provider address")
	}

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, primaryAcc)
	if !found {
		return nil, types2.ErrStorageProviderNotFound
	}

	amount, err := k.GetObjectLockFee(ctx, sp.GetId(), createAt, req.PayloadSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLockFeeResponse{Amount: amount}, nil
}
