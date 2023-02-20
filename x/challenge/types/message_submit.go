package types

import (
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSubmit = "submit"

var _ sdk.Msg = &MsgSubmit{}

func NewMsgSubmit(creator sdk.AccAddress, spOperatorAddress sdk.AccAddress, bucketName, objectName string, randomIndex bool, segmentIndex uint32) *MsgSubmit {
	return &MsgSubmit{
		Creator:           creator.String(),
		SpOperatorAddress: spOperatorAddress.String(),
		BucketName:        bucketName,
		ObjectName:        objectName,
		RandomIndex:       randomIndex,
		SegmentIndex:      segmentIndex,
	}
}

func (msg *MsgSubmit) Route() string {
	return RouterKey
}

func (msg *MsgSubmit) Type() string {
	return TypeMsgSubmit
}

func (msg *MsgSubmit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmit) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := storagetypes.CheckValidBucketName(msg.BucketName); err != nil {
		return err
	}

	if err := storagetypes.CheckValidObjectName(msg.ObjectName); err != nil {
		return err
	}

	if !msg.RandomIndex && msg.SegmentIndex > 5 {
		return sdkerrors.Wrapf(ErrInvalidIndex, "Index should be correctly provided when random index is disabled")
	}

	return nil
}
