package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgStorageProviderForceExit = "storage_provider_force_exit"

var _ sdk.Msg = &MsgStorageProviderForceExit{}

func NewMsgStorageProviderForceExit(authority string, spAddress sdk.AccAddress) *MsgStorageProviderForceExit {
	return &MsgStorageProviderForceExit{
		Authority:       authority,
		StorageProvider: spAddress.String(),
	}
}

func (msg *MsgStorageProviderForceExit) Route() string {
	return RouterKey
}

func (msg *MsgStorageProviderForceExit) Type() string {
	return TypeMsgStorageProviderForceExit
}

func (msg *MsgStorageProviderForceExit) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgStorageProviderForceExit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *MsgStorageProviderForceExit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp address (%s)", err)
	}
	return nil
}
