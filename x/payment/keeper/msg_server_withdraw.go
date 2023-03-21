package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// check stream record
	from := sdk.MustAccAddressFromHex(msg.From)
	streamRecord, found := k.Keeper.GetStreamRecord(ctx, from)
	if !found {
		return nil, types.ErrStreamRecordNotFound
	}
	// check whether creator can withdraw
	if msg.Creator != msg.From {
		paymentAccount, found := k.Keeper.GetPaymentAccount(ctx, from)
		if !found {
			return nil, types.ErrPaymentAccountNotFound
		}
		if paymentAccount.Owner != msg.Creator {
			return nil, types.ErrNotPaymentAccountOwner
		}
		if !paymentAccount.Refundable {
			return nil, types.ErrPaymentAccountAlreadyNonRefundable
		}
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(from).WithStaticBalanceChange(msg.Amount.Neg())
	err := k.UpdateStreamRecord(ctx, streamRecord, change, false)
	k.SetStreamRecord(ctx, streamRecord)
	if err != nil {
		return nil, err
	}
	if streamRecord.StaticBalance.IsNegative() {
		return nil, errors.Wrapf(types.ErrInsufficientBalance, "static balance: %s after withdraw", streamRecord.StaticBalance)
	}
	// bank transfer
	creator := sdk.MustAccAddressFromHex(msg.Creator)
	coins := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, msg.Amount))
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, coins)
	if err != nil {
		return nil, err
	}
	// emit event
	err = ctx.EventManager().EmitTypedEvents(&types.EventWithdraw{
		From:   msg.From,
		To:     msg.Creator,
		Amount: msg.Amount,
	})
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawResponse{}, nil
}
