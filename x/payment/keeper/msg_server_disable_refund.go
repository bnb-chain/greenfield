package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) DisableRefund(goCtx context.Context, msg *types.MsgDisableRefund) (*types.MsgDisableRefundResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr := sdk.MustAccAddressFromHex(msg.Addr)
	paymentAccount, found := k.Keeper.GetPaymentAccount(ctx, addr)
	if !found {
		return nil, types.ErrPaymentAccountNotFound
	}
	if paymentAccount.Owner != msg.Owner {
		return nil, types.ErrNotPaymentAccountOwner
	}
	if !paymentAccount.Refundable {
		return nil, types.ErrPaymentAccountAlreadyNonRefundable
	}
	paymentAccount.Refundable = false
	k.Keeper.SetPaymentAccount(ctx, paymentAccount)
	return &types.MsgDisableRefundResponse{}, nil
}
