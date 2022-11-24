package keeper

import (
	"context"
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetPaymentAccountsByUser(goCtx context.Context, req *types.QueryGetPaymentAccountsByUserRequest) (*types.QueryGetPaymentAccountsByUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	countRecord, found := k.GetPaymentAccountCount(ctx, req.User)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}
	count := countRecord.Count
	user := sdk.MustAccAddressFromBech32(req.User)
	var paymentAccounts []string
	var i uint64
	for i = 0; i < count; i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, i)
		paymentAccount := sdk.AccAddress(address.Derive(user.Bytes(), b)).String()
		paymentAccounts = append(paymentAccounts, paymentAccount)
	}

	return &types.QueryGetPaymentAccountsByUserResponse{
		PaymentAccounts: paymentAccounts,
	}, nil
}
