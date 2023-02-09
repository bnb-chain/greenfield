package keeper

import (
	"context"
	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// check stream record
	streamRecord, found := k.Keeper.GetStreamRecord(ctx, msg.From)
	if !found {
		return nil, types.ErrStreamRecordNotFound
	}
	// check whether creator can withdraw
	if msg.Creator != msg.From {
		paymentAccount, found := k.Keeper.GetPaymentAccount(ctx, msg.From)
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
	change := types.NewDefaultStreamRecordChangeWithAddr(msg.From).WithStaticBalanceChange(msg.Amount.Neg())
	err := k.UpdateStreamRecord(ctx, &streamRecord, change)
	if err != nil {
		return nil, err
	}
	// bank transfer
	creator, _ := sdk.AccAddressFromHexUnsafe(msg.Creator)
	coins := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, msg.Amount))
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, coins)
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawResponse{}, nil
}
