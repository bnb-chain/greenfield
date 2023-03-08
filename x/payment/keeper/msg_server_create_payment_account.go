package keeper

import (
	"context"
	"encoding/binary"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) CreatePaymentAccount(goCtx context.Context, msg *types.MsgCreatePaymentAccount) (*types.MsgCreatePaymentAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get current count
	creator := sdk.MustAccAddressFromHex(msg.Creator)
	countRecord, _ := k.Keeper.GetPaymentAccountCount(ctx, creator)
	count := countRecord.Count
	// get payment account count limit
	params := k.Keeper.GetParams(ctx)
	if count >= params.PaymentAccountCountLimit {
		return nil, errorsmod.Wrapf(types.ErrReachPaymentAccountLimit, "current count: %d", count)
	}
	// TODO: charge fee
	// calculate the addr
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, count)
	paymentAccountAddr := sdk.AccAddress(address.Derive(creator.Bytes(), b)).String()
	newCount := count + 1
	k.Keeper.SetPaymentAccountCount(ctx, &types.PaymentAccountCount{
		Owner: msg.Creator,
		Count: newCount,
	})
	k.Keeper.SetPaymentAccount(ctx, &types.PaymentAccount{
		Addr:       paymentAccountAddr,
		Owner:      msg.Creator,
		Refundable: true,
	})
	return &types.MsgCreatePaymentAccountResponse{}, nil
}
