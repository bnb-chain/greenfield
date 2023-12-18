package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgStorageProviderForcedExit = "storage_provider_force_exit"

var _ sdk.Msg = &MsgStorageProviderForcedExit{}

func NewMsgStorageProviderForcedExit(authority string, spAddress sdk.AccAddress) *MsgStorageProviderForcedExit {
	return &MsgStorageProviderForcedExit{
		Authority:       authority,
		StorageProvider: spAddress.String(),
	}
}

func (msg *MsgStorageProviderForcedExit) Route() string {
	return RouterKey
}

func (msg *MsgStorageProviderForcedExit) Type() string {
	return TypeMsgStorageProviderForcedExit
}

func (msg *MsgStorageProviderForcedExit) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgStorageProviderForcedExit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *MsgStorageProviderForcedExit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}
	if _, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp address (%s)", err)
	}
	return nil
}
