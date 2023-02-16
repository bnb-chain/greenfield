package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) PaymentAccountCountAll(c context.Context, req *types.QueryAllPaymentAccountCountRequest) (*types.QueryAllPaymentAccountCountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var paymentAccountCounts []types.PaymentAccountCount
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	paymentAccountCountStore := prefix.NewStore(store, types.PaymentAccountCountKeyPrefix)

	pageRes, err := query.Paginate(paymentAccountCountStore, req.Pagination, func(key []byte, value []byte) error {
		var paymentAccountCount types.PaymentAccountCount
		if err := k.cdc.Unmarshal(value, &paymentAccountCount); err != nil {
			return err
		}

		paymentAccountCounts = append(paymentAccountCounts, paymentAccountCount)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPaymentAccountCountResponse{PaymentAccountCount: paymentAccountCounts, Pagination: pageRes}, nil
}

func (k Keeper) PaymentAccountCount(c context.Context, req *types.QueryGetPaymentAccountCountRequest) (*types.QueryGetPaymentAccountCountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetPaymentAccountCount(
		ctx,
		req.Owner,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPaymentAccountCountResponse{PaymentAccountCount: val}, nil
}
