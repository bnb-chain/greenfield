package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCompleteStorageProviderExit = "complete_storage_provider_exit"

var _ sdk.Msg = &MsgCompleteStorageProviderExit{}

func NewMsgCompleteStorageProviderExit(operator sdk.AccAddress) *MsgCompleteStorageProviderExit {
	return &MsgCompleteStorageProviderExit{
		StorageProvider: operator.String(),
	}
}

func (msg *MsgCompleteStorageProviderExit) Route() string {
	return RouterKey
}

func (msg *MsgCompleteStorageProviderExit) Type() string {
	return TypeMsgCompleteStorageProviderExit
}

func (msg *MsgCompleteStorageProviderExit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
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
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	return nil
}
