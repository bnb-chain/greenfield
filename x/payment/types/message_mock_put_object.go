package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgMockPutObject = "mock_put_object"

var _ sdk.Msg = &MsgMockPutObject{}

func NewMsgMockPutObject(owner string, bucketName string, objectName string, size uint64, spAddr string) *MsgMockPutObject {
	return &MsgMockPutObject{
		Owner:      owner,
		BucketName: bucketName,
		ObjectName: objectName,
		Size_:      size,
		SpAddr:     spAddr,
	}
}

func (msg *MsgMockPutObject) Route() string {
	return RouterKey
}

func (msg *MsgMockPutObject) Type() string {
	return TypeMsgMockPutObject
}

func (msg *MsgMockPutObject) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromHexUnsafe(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

func (msg *MsgMockPutObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMockPutObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}
