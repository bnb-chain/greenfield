package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgMockSetBucketPaymentAccount = "mock_set_bucket_payment_account"

var _ sdk.Msg = &MsgMockSetBucketPaymentAccount{}

func NewMsgMockSetBucketPaymentAccount(operator string, bucketName string, readPaymentAccount string, storePaymentAccount string) *MsgMockSetBucketPaymentAccount {
  return &MsgMockSetBucketPaymentAccount{
		Operator: operator,
    BucketName: bucketName,
    ReadPaymentAccount: readPaymentAccount,
    StorePaymentAccount: storePaymentAccount,
	}
}

func (msg *MsgMockSetBucketPaymentAccount) Route() string {
  return RouterKey
}

func (msg *MsgMockSetBucketPaymentAccount) Type() string {
  return TypeMsgMockSetBucketPaymentAccount
}

func (msg *MsgMockSetBucketPaymentAccount) GetSigners() []sdk.AccAddress {
  operator, err := sdk.AccAddressFromBech32(msg.Operator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{operator}
}

func (msg *MsgMockSetBucketPaymentAccount) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgMockSetBucketPaymentAccount) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Operator)
  	if err != nil {
  		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
  	}
  return nil
}

