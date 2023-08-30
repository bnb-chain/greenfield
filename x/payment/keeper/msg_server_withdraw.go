package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	creator := sdk.MustAccAddressFromHex(msg.Creator)
	from := sdk.MustAccAddressFromHex(msg.From)

	params := k.GetParams(ctx)
	if ctx.IsUpgraded(upgradetypes.Nagqu) {
		if msg.Amount.GTE(params.WithdrawTimeLockThreshold) {
			// check whether there is delayed withdrawal, if there is delayed withdrawal, must withdraw it firstly
			delayedWithdrawal, found := k.GetDelayedWithdrawalRecord(ctx, creator)
			if found {
				if !delayedWithdrawal.Amount.Equal(msg.Amount) {
					return nil, errors.Wrapf(types.ErrIncorrectWithdrawAmount, "withdrawal amount should be equal to the delayed %s", delayedWithdrawal.Amount)
				}
				params := k.GetParams(ctx)
				now := ctx.BlockTime().Unix()
				end := delayedWithdrawal.Timestamp + int64(params.WithdrawTimeLockDuration)
				if now <= end {
					return nil, errors.Wrapf(types.ErrNotReachTimeLockDuration, "withdrawal should be after %d", end)
				}
				k.RemoveDelayedWithdrawalRecord(ctx, creator)

				// withdraw it from module account directly
				err := k.bankTransfer(ctx, creator, from, msg.Amount)
				if err != nil {
					return nil, err
				}
				return &types.MsgWithdrawResponse{}, nil
			}
		}
	}

	// check stream record
	streamRecord, found := k.Keeper.GetStreamRecord(ctx, from)
	if !found {
		return nil, types.ErrStreamRecordNotFound
	}
	// check status
	if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN {
		return nil, errors.Wrapf(types.ErrInvalidStreamAccountStatus, "stream record is frozen")
	}
	// check whether creator can withdraw
	if !creator.Equals(from) {
		paymentAccount, found := k.Keeper.GetPaymentAccount(ctx, from)
		if !found {
			return nil, types.ErrPaymentAccountNotFound
		}
		owner := sdk.MustAccAddressFromHex(paymentAccount.Owner)
		if !creator.Equals(owner) {
			return nil, types.ErrNotPaymentAccountOwner
		}
		if !paymentAccount.Refundable {
			return nil, types.ErrPaymentAccountAlreadyNonRefundable
		}
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(from).WithStaticBalanceChange(msg.Amount.Neg())
	err := k.UpdateStreamRecord(ctx, streamRecord, change)
	if err != nil {
		return nil, err
	}
	k.SetStreamRecord(ctx, streamRecord)
	if streamRecord.StaticBalance.IsNegative() {
		return nil, errors.Wrapf(types.ErrInsufficientBalance, "static balance: %s after withdraw", streamRecord.StaticBalance)
	}

	if ctx.IsUpgraded(upgradetypes.Nagqu) {
		if msg.Amount.GTE(params.WithdrawTimeLockThreshold) {
			delayedWithdrawal := &types.DelayedWithdrawalRecord{
				Timestamp: ctx.BlockTime().Unix(),
				Addr:      creator.String(),
				Amount:    msg.Amount,
			}
			k.SetDelayedWithdrawalRecord(ctx, delayedWithdrawal)
			return &types.MsgWithdrawResponse{}, nil // TODO: how to let the user know that his/her withdrawal is delayed?
		}
	}

	// bank transfer
	err = k.bankTransfer(ctx, creator, from, msg.Amount)
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawResponse{}, nil
}

func (k msgServer) bankTransfer(ctx sdk.Context, creator, from sdk.AccAddress, amount math.Int) error {
	coins := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, amount))
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, coins)
	if err != nil {
		return err
	}
	// emit event
	err = ctx.EventManager().EmitTypedEvents(&types.EventWithdraw{
		From:   from.String(),
		To:     creator.String(),
		Amount: amount,
	})
	return err
}
