package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgMockDeleteObject = "mock_delete_object"

var _ sdk.Msg = &MsgMockDeleteObject{}

func NewMsgMockDeleteObject(operator string, bucketName string, objectName string) *MsgMockDeleteObject {
  return &MsgMockDeleteObject{
		Operator: operator,
    BucketName: bucketName,
    ObjectName: objectName,
	}
}

func (msg *MsgMockDeleteObject) Route() string {
  return RouterKey
}

func (msg *MsgMockDeleteObject) Type() string {
  return TypeMsgMockDeleteObject
}

func (msg *MsgMockDeleteObject) GetSigners() []sdk.AccAddress {
  operator, err := sdk.AccAddressFromBech32(msg.Operator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{operator}
}

func (msg *MsgMockDeleteObject) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgMockDeleteObject) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Operator)
  	if err != nil {
  		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
  	}
  return nil
}

