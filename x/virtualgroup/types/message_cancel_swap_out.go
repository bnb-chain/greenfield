package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCancelSwapOut = "cancel_swap_out"

var _ sdk.Msg = &MsgCancelSwapOut{}

func NewMsgCancelSwapOut(creator string) *MsgCancelSwapOut {
	return &MsgCancelSwapOut{
		Creator: creator,
	}
}

func (msg *MsgCancelSwapOut) Route() string {
	return RouterKey
}

func (msg *MsgCancelSwapOut) Type() string {
	return TypeMsgCancelSwapOut
}

func (msg *MsgCancelSwapOut) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCancelSwapOut) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCancelSwapOut) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
