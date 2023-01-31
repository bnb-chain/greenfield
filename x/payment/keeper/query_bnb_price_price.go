package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/bnb-chain/bfs/x/payment/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BnbPricePriceAll(c context.Context, req *types.QueryAllBnbPricePriceRequest) (*types.QueryAllBnbPricePriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var bnbPricePrices []types.BnbPricePrice
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	bnbPricePriceStore := prefix.NewStore(store, types.KeyPrefix(types.BnbPricePriceKeyPrefix))

	pageRes, err := query.Paginate(bnbPricePriceStore, req.Pagination, func(key []byte, value []byte) error {
		var bnbPricePrice types.BnbPricePrice
		if err := k.cdc.Unmarshal(value, &bnbPricePrice); err != nil {
			return err
		}

		bnbPricePrices = append(bnbPricePrices, bnbPricePrice)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllBnbPricePriceResponse{BnbPricePrice: bnbPricePrices, Pagination: pageRes}, nil
}

func (k Keeper) BnbPricePrice(c context.Context, req *types.QueryGetBnbPricePriceRequest) (*types.QueryGetBnbPricePriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetBnbPricePrice(
	    ctx,
	    req.Time,
        )
	if !found {
	    return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBnbPricePriceResponse{BnbPricePrice: val}, nil
}