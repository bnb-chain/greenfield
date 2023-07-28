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

func (k Keeper) PaymentAccounts(c context.Context, req *types.QueryPaymentAccountsRequest) (*types.QueryPaymentAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if err := query.CheckOffsetQueryNotAllowed(ctx, req.Pagination); err != nil {
		return nil, err
	}

	var paymentAccounts []types.PaymentAccount
	store := ctx.KVStore(k.storeKey)
	paymentAccountStore := prefix.NewStore(store, types.PaymentAccountKeyPrefix)

	pageRes, err := query.Paginate(paymentAccountStore, req.Pagination, func(key []byte, value []byte) error {
		var paymentAccount types.PaymentAccount
		if err := k.cdc.Unmarshal(value, &paymentAccount); err != nil {
			return err
		}
		paymentAccount.Addr = sdk.AccAddress(key).String()

		paymentAccounts = append(paymentAccounts, paymentAccount)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPaymentAccountsResponse{PaymentAccounts: paymentAccounts, Pagination: pageRes}, nil
}

func (k Keeper) PaymentAccount(c context.Context, req *types.QueryGetPaymentAccountRequest) (*types.QueryGetPaymentAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromHexUnsafe(req.Addr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}
	val, found := k.GetPaymentAccount(
		ctx,
		addr,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPaymentAccountResponse{PaymentAccount: *val}, nil
}
