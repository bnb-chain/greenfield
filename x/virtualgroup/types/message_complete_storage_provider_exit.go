package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
	operator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		spAddr, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
		if err != nil {
			panic(err)
		}
		return []sdk.AccAddress{spAddr}
	}
	// the operator address will be validated in runtime after harfork and treated as signer
	return []sdk.AccAddress{operator}
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
