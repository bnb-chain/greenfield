package keeper

import (
	"context"
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/types/address"

	errorsmod "cosmossdk.io/errors"
	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreatePaymentAccount(goCtx context.Context, msg *types.MsgCreatePaymentAccount) (*types.MsgCreatePaymentAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get current count
	countRecord, _ := k.Keeper.GetPaymentAccountCount(ctx, msg.Creator)
	count := countRecord.Count
	// get payment account count limit
	params := k.Keeper.GetParams(ctx)
	if count >= params.PaymentAccountCountLimit {
		return nil, errorsmod.Wrapf(types.ErrReachPaymentAccountLimit, "current count: %d", count)
	}
	creator := sdk.MustAccAddressFromHex(msg.Creator)
	// TODO: charge fee
	// calculate the addr
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, count)
	paymentAccountAddr := sdk.AccAddress(address.Derive(creator.Bytes(), b)).String()
	newCount := count + 1
	k.Keeper.SetPaymentAccountCount(ctx, types.PaymentAccountCount{
		Owner: msg.Creator,
		Count: newCount,
	})
	k.Keeper.SetPaymentAccount(ctx, types.PaymentAccount{
		Addr:       paymentAccountAddr,
		Owner:      msg.Creator,
		Refundable: true,
	})
	err := ctx.EventManager().EmitTypedEvents(&types.EventCreatePaymentAccount{
		Addr:  paymentAccountAddr,
		Owner: msg.Creator,
		Index: count,
	})
	if err != nil {
		return nil, err
	}
	return &types.MsgCreatePaymentAccountResponse{Count: newCount, Addr: paymentAccountAddr}, nil
}
