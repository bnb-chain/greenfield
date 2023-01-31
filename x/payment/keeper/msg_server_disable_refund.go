package keeper

import (
	"context"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DisableRefund(goCtx context.Context, msg *types.MsgDisableRefund) (*types.MsgDisableRefundResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	paymentAccount, found := k.Keeper.GetPaymentAccount(ctx, msg.Addr)
	if !found {
		return nil, types.ErrPaymentAccountNotFound
	}
	if paymentAccount.Owner != msg.Creator {
		return nil, types.ErrNotPaymentAccountOwner
	}
	if !paymentAccount.Refundable {
		return nil, types.ErrPaymentAccountAlreadyNonRefundable
	}
	paymentAccount.Refundable = false
	k.Keeper.SetPaymentAccount(ctx, paymentAccount)
	return &types.MsgDisableRefundResponse{}, nil
}
