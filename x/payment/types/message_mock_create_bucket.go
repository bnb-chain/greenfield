package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgMockCreateBucket = "mock_create_bucket"

var _ sdk.Msg = &MsgMockCreateBucket{}

func NewMsgMockCreateBucket(operator string, bucketName string, readPaymentAccount string, storePaymentAccount string, spAddress string, readPacket ReadPacket) *MsgMockCreateBucket {
	return &MsgMockCreateBucket{
		Operator:            operator,
		BucketName:          bucketName,
		ReadPaymentAccount:  readPaymentAccount,
		StorePaymentAccount: storePaymentAccount,
		SpAddress:           spAddress,
		ReadPacket:          readPacket,
	}
}

func (msg *MsgMockCreateBucket) Route() string {
	return RouterKey
}

func (msg *MsgMockCreateBucket) Type() string {
	return TypeMsgMockCreateBucket
}

func (msg *MsgMockCreateBucket) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgMockCreateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMockCreateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	return nil
}
