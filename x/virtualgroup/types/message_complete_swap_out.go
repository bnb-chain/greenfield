package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCompleteSwapOut = "complete_swap_out"

var _ sdk.Msg = &MsgCompleteSwapOut{}

func NewMsgCompleteSwapOut(creator string) *MsgCompleteSwapOut {
	return &MsgCompleteSwapOut{
		StorageProvider: creator,
	}
}

func (msg *MsgCompleteSwapOut) Route() string {
	return RouterKey
}

func (msg *MsgCompleteSwapOut) Type() string {
	return TypeMsgCompleteSwapOut
}

func (msg *MsgCompleteSwapOut) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.StorageProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCompleteSwapOut) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCompleteSwapOut) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.StorageProvider)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
