package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// bank transfer
	creator := sdk.MustAccAddressFromHex(msg.Creator)
	to := sdk.MustAccAddressFromHex(msg.To)
	coins := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, msg.Amount))
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, coins)
	if err != nil {
		return nil, err
	}
	// change payment record
	streamRecord, found := k.GetStreamRecord(ctx, msg.To)

	if !found {
		// if not found, check whether the account exists, if exists, create a new record, otherwise, return error
		_, paymentAccountExists := k.GetPaymentAccount(ctx, to.String())
		if !paymentAccountExists && !k.accountKeeper.HasAccount(ctx, to) {
			return nil, types.ErrReceiveAccountNotExist
		}
		streamRecord.Account = msg.To
		streamRecord.CrudTimestamp = ctx.BlockTime().Unix()
		streamRecord.StaticBalance = msg.Amount
		k.SetStreamRecord(ctx, streamRecord)
	} else {
		if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE {
			// add static balance
			change := types.NewDefaultStreamRecordChangeWithAddr(msg.To).WithStaticBalanceChange(msg.Amount)
			err = k.UpdateStreamRecord(ctx, streamRecord, change, false)
			if err != nil {
				return nil, err
			}
			k.SetStreamRecord(ctx, streamRecord)
		} else if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN {
			// deposit and try resume the account
			err = k.TryResumeStreamRecord(ctx, streamRecord, msg.Amount)
			if err != nil {
				return nil, err
			}
		} else {
			// status can only be normal or frozen
			return nil, types.ErrInvalidStreamAccountStatus
		}
	}

	event := types.EventDeposit{
		From:   creator.String(),
		To:     to.String(),
		Amount: msg.Amount,
	}
	err = ctx.EventManager().EmitTypedEvents(&event)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositResponse{}, nil
}
