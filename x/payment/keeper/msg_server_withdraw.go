package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

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
	if err != nil {
		return nil, err
	}
	if streamRecord.StaticBalance.IsNegative() {
		return nil, errors.Wrapf(types.ErrInsufficientBalance, "static balance: %s after withdraw", streamRecord.StaticBalance)
	}
	// bank transfer
	feeDenom := k.GetParams(ctx).FeeDenom
	validatorFeeRate := k.GetParams(ctx).ValidatorFeeRate
	toValidators := validatorFeeRate.MulInt(msg.Amount).TruncateInt()
	toCreator := msg.Amount.Sub(toValidators)

	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, distributiontypes.ModuleName,
		sdk.NewCoins(sdk.NewCoin(feeDenom, toValidators)))
	if err != nil {
		return nil, err
	}

	creator, _ := sdk.AccAddressFromHexUnsafe(msg.Creator)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator,
		sdk.NewCoins(sdk.NewCoin(feeDenom, toCreator)))
	if err != nil {
		return nil, err
	}
	// emit event
	err = ctx.EventManager().EmitTypedEvents(&types.EventWithdraw{
		From:   msg.From,
		To:     msg.Creator,
		Amount: toCreator,
	}, &types.EventWithdraw{
		From:   msg.From,
		To:     authtypes.NewModuleAddress(distributiontypes.ModuleName).String(),
		Amount: toValidators,
	})
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawResponse{}, nil
}
