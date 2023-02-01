package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BnbPriceAll(c context.Context, req *types.QueryAllBnbPriceRequest) (*types.QueryAllBnbPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var BnbPrices []types.BnbPrice
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	BnbPriceStore := prefix.NewStore(store, types.BnbPriceKeyPrefix)

	pageRes, err := query.Paginate(BnbPriceStore, req.Pagination, func(key []byte, value []byte) error {
		var BnbPrice types.BnbPrice
		if err := k.cdc.Unmarshal(value, &BnbPrice); err != nil {
			return err
		}

		BnbPrices = append(BnbPrices, BnbPrice)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllBnbPriceResponse{BnbPrice: BnbPrices, Pagination: pageRes}, nil
}

func (k Keeper) BnbPrice(c context.Context, req *types.QueryGetBnbPriceRequest) (*types.QueryGetBnbPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetBnbPrice(
		ctx,
		req.Time,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBnbPriceResponse{BnbPrice: val}, nil
}
