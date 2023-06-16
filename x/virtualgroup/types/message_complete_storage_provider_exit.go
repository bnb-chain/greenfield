package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCompleteStorageProviderExit = "complete_storage_provider_exit"

var _ sdk.Msg = &MsgCompleteStorageProviderExit{}

func NewMsgCompleteStorageProviderExit(creator string) *MsgCompleteStorageProviderExit {
	return &MsgCompleteStorageProviderExit{
		Creator: creator,
	}
}

func (msg *MsgCompleteStorageProviderExit) Route() string {
	return RouterKey
}

func (msg *MsgCompleteStorageProviderExit) Type() string {
	return TypeMsgCompleteStorageProviderExit
}

func (msg *MsgCompleteStorageProviderExit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCompleteStorageProviderExit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCompleteStorageProviderExit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
