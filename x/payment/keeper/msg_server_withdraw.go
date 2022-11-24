package keeper

import (
	"context"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check stream record
	streamRecord, found := k.Keeper.GetStreamRecord(ctx, msg.From)
	if !found {
		return nil, types.ErrStreamRecordNotFound
	}
	k.UpdateStreamRecord(ctx, &streamRecord)
	if streamRecord.StaticBalance < msg.Amount {
		return nil, types.ErrInsufficientBalance
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
			return nil, types.ErrPaymentAccountNotRefundable
		}
	}
	// bank transfer
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	coins := sdk.NewCoins(sdk.NewCoin(types.Denom, sdk.NewInt(msg.Amount)))
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, coins)
	if err != nil {
		return nil, err
	}
	// change stream record
	streamRecord.StaticBalance -= msg.Amount
	k.SetStreamRecord(ctx, streamRecord)

	return &types.MsgWithdrawResponse{}, nil
}
