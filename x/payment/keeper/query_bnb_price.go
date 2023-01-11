package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/bnb-chain/bfs/x/payment/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BnbPrice(c context.Context, req *types.QueryGetBnbPriceRequest) (*types.QueryGetBnbPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetBnbPrice(ctx)
	if !found {
	    return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBnbPriceResponse{BnbPrice: val}, nil
}