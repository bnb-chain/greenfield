package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreatePaymentAccount(goCtx context.Context, msg *types.MsgCreatePaymentAccount) (*types.MsgCreatePaymentAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get current count
	creator := sdk.MustAccAddressFromHex(msg.Creator)
	countRecord, _ := k.GetPaymentAccountCount(ctx, creator)
	count := countRecord.Count
	// get payment account count limit
	params := k.GetParams(ctx)
	if count >= params.PaymentAccountCountLimit {
		return nil, errorsmod.Wrapf(types.ErrReachPaymentAccountLimit, "current count: %d, limit: %d", count, params.PaymentAccountCountLimit)
	}
	paymentAccountAddr := k.DerivePaymentAccountAddress(creator, count).String()
	newCount := count + 1
	k.SetPaymentAccountCount(ctx, &types.PaymentAccountCount{
		Owner: msg.Creator,
		Count: newCount,
	})
	k.SetPaymentAccount(ctx, &types.PaymentAccount{
		Addr:       paymentAccountAddr,
		Owner:      msg.Creator,
		Refundable: true,
	})
	return &types.MsgCreatePaymentAccountResponse{}, nil
}
