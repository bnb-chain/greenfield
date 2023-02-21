package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateSpStoragePrice = "update_sp_storage_price"

var _ sdk.Msg = &MsgUpdateSpStoragePrice{}

func (msg *MsgUpdateSpStoragePrice) Route() string {
	return RouterKey
}

func (msg *MsgUpdateSpStoragePrice) Type() string {
	return TypeMsgUpdateSpStoragePrice
}

func (msg *MsgUpdateSpStoragePrice) GetSigners() []sdk.AccAddress {
	spAddr, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{spAddr}
}

func (msg *MsgUpdateSpStoragePrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateSpStoragePrice) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp address (%s)", err)
	}
	return nil
}
