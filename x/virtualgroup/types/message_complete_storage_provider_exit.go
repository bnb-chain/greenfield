package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const TypeMsgCompleteStorageProviderExit = "complete_storage_provider_exit"

var _ sdk.Msg = &MsgCompleteStorageProviderExit{}

func NewMsgCompleteStorageProviderExit(operator, storageProvider sdk.AccAddress) *MsgCompleteStorageProviderExit {
	return &MsgCompleteStorageProviderExit{
		Operator:        operator.String(),
		StorageProvider: storageProvider.String(),
	}
}

func (msg *MsgCompleteStorageProviderExit) Route() string {
	return RouterKey
}

func (msg *MsgCompleteStorageProviderExit) Type() string {
	return TypeMsgCompleteStorageProviderExit
}

func (msg *MsgCompleteStorageProviderExit) GetSigners() []sdk.AccAddress {
	spAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)
	operator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err == nil {
		return []sdk.AccAddress{operator}
	}
	return []sdk.AccAddress{spAddr}
}

func (msg *MsgCompleteStorageProviderExit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCompleteStorageProviderExit) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid storage provider address (%s)", err)
	}
	return nil
}

func (msg *MsgCompleteStorageProviderExit) ValidateRuntime(ctx sdk.Context) error {
	if ctx.IsUpgraded(upgradetypes.Hulunbeier) {
		_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
		}
	}
	return nil
}
