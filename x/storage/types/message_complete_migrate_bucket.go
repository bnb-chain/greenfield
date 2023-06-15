package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCompleteMigrateBucket = "complete_migrate_bucket"

var _ sdk.Msg = &MsgCompleteMigrateBucket{}

func NewMsgCompleteMigrateBucket(creator string) *MsgCompleteMigrateBucket {
	return &MsgCompleteMigrateBucket{
		Creator: creator,
	}
}

func (msg *MsgCompleteMigrateBucket) Route() string {
	return RouterKey
}

func (msg *MsgCompleteMigrateBucket) Type() string {
	return TypeMsgCompleteMigrateBucket
}

func (msg *MsgCompleteMigrateBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCompleteMigrateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCompleteMigrateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
