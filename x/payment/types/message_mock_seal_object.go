package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgMockSealObject = "mock_seal_object"

var _ sdk.Msg = &MsgMockSealObject{}

func NewMsgMockSealObject(operator string, bucketName string, objectName string, secondarySPs []string) *MsgMockSealObject {
	return &MsgMockSealObject{
		Operator:     operator,
		BucketName:   bucketName,
		ObjectName:   objectName,
		SecondarySPs: secondarySPs,
	}
}

func (msg *MsgMockSealObject) Route() string {
	return RouterKey
}

func (msg *MsgMockSealObject) Type() string {
	return TypeMsgMockSealObject
}

func (msg *MsgMockSealObject) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgMockSealObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMockSealObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	return nil
}
