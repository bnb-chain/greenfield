package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// bank transfer
	creator := sdk.MustAccAddressFromHex(msg.Creator)
	to := sdk.MustAccAddressFromHex(msg.To)
	depositAmount := msg.Amount
	coinsToDeposit := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, depositAmount))
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, coinsToDeposit)
	if err != nil {
		return nil, err
	}
	if ctx.IsUpgraded(upgradetypes.Nagqu) && k.IsPaymentAccount(ctx, to) {
		balanceOfToAccount := k.bankKeeper.GetBalance(ctx, to, k.GetParams(ctx).FeeDenom)
		if balanceOfToAccount.IsPositive() {
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, to, types.ModuleName, sdk.NewCoins(balanceOfToAccount))
			if err != nil {
				return nil, err
			}
			depositAmount = depositAmount.Add(balanceOfToAccount.Amount)
		}
	}

	// change payment record
	streamRecord, found := k.GetStreamRecord(ctx, to)

	if !found {
		// if not found, check whether the account exists, if exists, create a new record, otherwise, return error
		_, paymentAccountExists := k.GetPaymentAccount(ctx, to)
		if !paymentAccountExists && !k.accountKeeper.HasAccount(ctx, to) {
			return nil, types.ErrReceiveAccountNotExist
		}
		streamRecord.Account = msg.To
		streamRecord.CrudTimestamp = ctx.BlockTime().Unix()
		streamRecord.StaticBalance = depositAmount
		k.SetStreamRecord(ctx, streamRecord)
	} else {
		if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE {
			// add static balance
			change := types.NewDefaultStreamRecordChangeWithAddr(to).WithStaticBalanceChange(depositAmount)
			err = k.UpdateStreamRecord(ctx, streamRecord, change)
			if err != nil {
				return nil, err
			}
			k.SetStreamRecord(ctx, streamRecord)
		} else if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN {
			// deposit and try resume the account
			err = k.TryResumeStreamRecord(ctx, streamRecord, depositAmount)
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
		Amount: depositAmount,
	}
	err = ctx.EventManager().EmitTypedEvents(&event)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositResponse{}, nil
}
