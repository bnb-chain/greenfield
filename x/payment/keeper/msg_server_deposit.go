package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// bank transfer
	creator, _ := sdk.AccAddressFromHexUnsafe(msg.Creator)
	coins := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, msg.Amount))
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, coins)
	if err != nil {
		return nil, err
	}

	// change payment record
	streamRecord, found := k.GetStreamRecord(ctx, msg.To)

	// if not found, check whether the account exists, if so, create a new record, otherwise, return error
	if !found {
		_, paymentAccountExists := k.GetPaymentAccount(ctx, msg.To)
		toAcc, _ := sdk.AccAddressFromHexUnsafe(msg.To)
		if !paymentAccountExists && !k.accountKeeper.HasAccount(ctx, toAcc) {
			return nil, types.ErrReceiveAccountNotExist
		}
		streamRecord.Account = msg.To
		streamRecord.CrudTimestamp = ctx.BlockTime().Unix()
		streamRecord.StaticBalance = msg.Amount
		k.Keeper.SetStreamRecord(ctx, streamRecord)
		return &types.MsgDepositResponse{}, nil
	}

	// add static balance
	if streamRecord.Status == types.StreamPaymentAccountStatusNormal {
		change := types.NewDefaultStreamRecordChangeWithAddr(msg.To).WithStaticBalanceChange(msg.Amount)
		err = k.UpdateStreamRecord(ctx, streamRecord, change, false)
		k.SetStreamRecord(ctx, streamRecord)
		return &types.MsgDepositResponse{}, err
	}
	// status can only be normal or frozen
	if streamRecord.Status != types.StreamPaymentAccountStatusFrozen {
		return nil, types.ErrInvalidStreamAccountStatus
	}
	// deposit and try resume the account
	err = k.TryResumeStreamRecord(ctx, streamRecord, msg.Amount)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositResponse{}, err
}
