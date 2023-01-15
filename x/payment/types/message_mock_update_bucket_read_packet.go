package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgMockUpdateBucketReadPacket = "mock_update_bucket_read_packet"

var _ sdk.Msg = &MsgMockUpdateBucketReadPacket{}

func NewMsgMockUpdateBucketReadPacket(operator string, bucketName string, readPacket int32) *MsgMockUpdateBucketReadPacket {
	return &MsgMockUpdateBucketReadPacket{
		Operator:   operator,
		BucketName: bucketName,
		ReadPacket: readPacket,
	}
}

func (msg *MsgMockUpdateBucketReadPacket) Route() string {
	return RouterKey
}

func (msg *MsgMockUpdateBucketReadPacket) Type() string {
	return TypeMsgMockUpdateBucketReadPacket
}

func (msg *MsgMockUpdateBucketReadPacket) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgMockUpdateBucketReadPacket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMockUpdateBucketReadPacket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	return nil
}
