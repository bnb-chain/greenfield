package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) PaymentAccountsByOwner(goCtx context.Context, req *types.QueryPaymentAccountsByOwnerRequest) (*types.QueryPaymentAccountsByOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromHexUnsafe(req.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}
	countRecord, found := k.GetPaymentAccountCount(ctx, owner)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}
	count := countRecord.Count
	var paymentAccounts []string
	var i uint64
	for i = 0; i < count; i++ {
		paymentAccount := k.DerivePaymentAccountAddress(owner, i)
		paymentAccounts = append(paymentAccounts, paymentAccount.String())
	}

	return &types.QueryPaymentAccountsByOwnerResponse{
		PaymentAccounts: paymentAccounts,
	}, nil
}
