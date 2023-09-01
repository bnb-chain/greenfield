package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) DelayedWithdrawal(goCtx context.Context, req *types.QueryDelayedWithdrawalRequest) (*types.QueryDelayedWithdrawalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	account, err := sdk.AccAddressFromHexUnsafe(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account")
	}
	delayedWithdrawal, found := k.GetDelayedWithdrawalRecord(
		ctx,
		account,
	)

	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryDelayedWithdrawalResponse{DelayedWithdrawal: *delayedWithdrawal}, nil
}
