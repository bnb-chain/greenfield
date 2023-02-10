package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var _ authz.Authorization = &DepositAuthorization{}

// NewDepositAuthorization creates a new DepositAuthorization object.
func NewDepositAuthorization(spAddress sdk.AccAddress, amount *sdk.Coin) (*DepositAuthorization, error) {
	a := DepositAuthorization{}
	a.SpAddress = spAddress.String()

	if amount != nil {
		a.MaxDeposit = amount
	}

	return &a, nil
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a DepositAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgDeposit{})
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a DepositAuthorization) ValidateBasic() error {
	if a.MaxDeposit != nil && a.MaxDeposit.IsNegative() {
		return sdkerrors.Wrapf(authz.ErrNegativeMaxTokens, "negative coin amount: %v", a.MaxDeposit)
	}
	return nil
}

// Accept implements Authorization.Accept.
func (a DepositAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	mDeposit, ok := msg.(*MsgDeposit)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidRequest.Wrap("msg type mismatch")
	}

	sp := a.GetSpAddress()
	if sp != mDeposit.SpAddress {
		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("cannot deposit to %s storage provider", mDeposit.SpAddress)
	}

	if a.MaxDeposit == nil {
		return authz.AcceptResponse{
			Accept: true, Delete: false,
			Updated: &DepositAuthorization{SpAddress: a.SpAddress},
		}, nil
	}

	limitLeft, err := a.MaxDeposit.SafeSub(mDeposit.Deposit)
	if err != nil {
		return authz.AcceptResponse{}, err
	}
	if limitLeft.IsZero() {
		return authz.AcceptResponse{Accept: true, Delete: true}, nil
	}
	return authz.AcceptResponse{
		Accept: true, Delete: false,
		Updated: &DepositAuthorization{SpAddress: a.SpAddress, MaxDeposit: &limitLeft},
	}, nil
}
