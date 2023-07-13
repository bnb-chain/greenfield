package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgStorageProviderExit = "storage_provider_exit"

var _ sdk.Msg = &MsgStorageProviderExit{}

func NewMsgStorageProviderExit(operatorAddress sdk.AccAddress) *MsgStorageProviderExit {
	return &MsgStorageProviderExit{
		StorageProvider: operatorAddress.String(),
	}
}

func (msg *MsgStorageProviderExit) Route() string {
	return RouterKey
}

func (msg *MsgStorageProviderExit) Type() string {
	return TypeMsgStorageProviderExit
}

func (msg *MsgStorageProviderExit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgStorageProviderExit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgStorageProviderExit) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
