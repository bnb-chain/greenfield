package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreatePaymentAccount = "create_payment_account"

var _ sdk.Msg = &MsgCreatePaymentAccount{}

func NewMsgCreatePaymentAccount(creator string) *MsgCreatePaymentAccount {
	return &MsgCreatePaymentAccount{
		Creator: creator,
	}
}

func (msg *MsgCreatePaymentAccount) Route() string {
	return RouterKey
}

func (msg *MsgCreatePaymentAccount) Type() string {
	return TypeMsgCreatePaymentAccount
}

func (msg *MsgCreatePaymentAccount) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreatePaymentAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreatePaymentAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
