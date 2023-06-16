package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgMigrateBucket = "migrate_bucket"

var _ sdk.Msg = &MsgMigrateBucket{}

func NewMsgMigrateBucket(creator string) *MsgMigrateBucket {
	return &MsgMigrateBucket{
		Creator: creator,
	}
}

func (msg *MsgMigrateBucket) Route() string {
	return RouterKey
}

func (msg *MsgMigrateBucket) Type() string {
	return TypeMsgMigrateBucket
}

func (msg *MsgMigrateBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgMigrateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMigrateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
