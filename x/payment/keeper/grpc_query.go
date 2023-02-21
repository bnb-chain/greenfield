package keeper

import (
	"context"
	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) QueryGetSpStoragePriceByTime(goCtx context.Context, req *types.QueryGetSpStoragePriceByTimeRequest) (*types.QueryGetSpStoragePriceByTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Timestamp <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid timestamp")
	}
	_, err := sdk.AccAddressFromHexUnsafe(req.SpAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid sp address")
	}
	spStoragePrice, err := k.GetSpStoragePriceByTime(ctx, req.SpAddr, req.Timestamp)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &types.QueryGetSpStoragePriceByTimeResponse{SpStoragePrice: spStoragePrice}, nil
}
