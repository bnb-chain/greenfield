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

func (k Keeper) PaymentAccountCounts(c context.Context, req *types.QueryPaymentAccountCountsRequest) (*types.QueryPaymentAccountCountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if err := query.CheckOffsetQueryNotAllowed(ctx, req.Pagination); err != nil {
		return nil, err
	}

	var paymentAccountCounts []types.PaymentAccountCount
	store := ctx.KVStore(k.storeKey)
	paymentAccountCountStore := prefix.NewStore(store, types.PaymentAccountCountKeyPrefix)

	pageRes, err := query.Paginate(paymentAccountCountStore, req.Pagination, func(key []byte, value []byte) error {
		var paymentAccountCount types.PaymentAccountCount
		if err := k.cdc.Unmarshal(value, &paymentAccountCount); err != nil {
			return err
		}
		paymentAccountCount.Owner = sdk.AccAddress(key).String()

		paymentAccountCounts = append(paymentAccountCounts, paymentAccountCount)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPaymentAccountCountsResponse{PaymentAccountCounts: paymentAccountCounts, Pagination: pageRes}, nil
}

func (k Keeper) PaymentAccountCount(c context.Context, req *types.QueryPaymentAccountCountRequest) (*types.QueryPaymentAccountCountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	owner, err := sdk.AccAddressFromHexUnsafe(req.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner")
	}
	val, found := k.GetPaymentAccountCount(
		ctx,
		owner,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryPaymentAccountCountResponse{PaymentAccountCount: *val}, nil
}
