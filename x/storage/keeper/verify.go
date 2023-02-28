package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) VerifyPaymentAccount(ctx sdk.Context, paymentAddress string, ownerAcc sdk.AccAddress) (sdk.AccAddress, error) {
	paymentAcc, err := sdk.AccAddressFromHexUnsafe(paymentAddress)
	if err == sdk.ErrEmptyHexAddress {
		return ownerAcc, nil
	} else if err != nil {
		return nil, err
	}

	if !k.paymentKeeper.IsPaymentAccountOwner(ctx, paymentAcc.String(), ownerAcc.String()) {
		return nil, paymenttypes.ErrNotPaymentAccountOwner
	}
	return paymentAcc, nil
}
