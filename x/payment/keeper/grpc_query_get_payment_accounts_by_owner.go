package keeper

import (
	"context"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) GetPaymentAccountsByOwner(goCtx context.Context, req *types.QueryGetPaymentAccountsByOwnerRequest) (*types.QueryGetPaymentAccountsByOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	countRecord, found := k.GetPaymentAccountCount(ctx, req.Owner)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}
	count := countRecord.Count
	user := sdk.MustAccAddressFromHex(req.Owner)
	var paymentAccounts []string
	var i uint64
	for i = 0; i < count; i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, i)
		paymentAccount := sdk.AccAddress(address.Derive(user.Bytes(), b)).String()
		paymentAccounts = append(paymentAccounts, paymentAccount)
	}

	return &types.QueryGetPaymentAccountsByOwnerResponse{
		PaymentAccounts: paymentAccounts,
	}, nil
}
