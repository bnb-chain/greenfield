package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDisableRefund = "disable_refund"

var _ sdk.Msg = &MsgDisableRefund{}

func NewMsgDisableRefund(owner string, addr string) *MsgDisableRefund {
	return &MsgDisableRefund{
		Owner: owner,
		Addr:  addr,
	}
}

func (msg *MsgDisableRefund) Route() string {
	return RouterKey
}

func (msg *MsgDisableRefund) Type() string {
	return TypeMsgDisableRefund
}

func (msg *MsgDisableRefund) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDisableRefund) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDisableRefund) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Owner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	_, err = sdk.AccAddressFromHexUnsafe(msg.Addr)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid payment address (%s)", err)
	}
	return nil
}
