package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) VerifyPaymentAccount(ctx sdk.Context, paymentAddress string, ownerAcc sdk.AccAddress) (sdk.AccAddress, error) {
	paymentAcc, err := sdk.AccAddressFromHexUnsafe(paymentAddress)
	if err == sdk.ErrEmptyHexAddress {
		return ownerAcc, nil
	} else if err != nil {
		return nil, err
	}

	// don't check if the payment account is owned by the owner account
	if !ctx.IsUpgraded(upgradetypes.Pawnee) {
		if !k.paymentKeeper.IsPaymentAccountOwner(ctx, paymentAcc, ownerAcc) {
			return nil, paymenttypes.ErrNotPaymentAccountOwner
		}
	}

	return paymentAcc, nil
}
