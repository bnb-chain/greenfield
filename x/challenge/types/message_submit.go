package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bnb-chain/greenfield/types/s3util"
)

const TypeMsgSubmit = "submit"

var _ sdk.Msg = &MsgSubmit{}

func NewMsgSubmit(challenger sdk.AccAddress, spOperatorAddress sdk.AccAddress, bucketName, objectName string, randomIndex bool, segmentIndex uint32) *MsgSubmit {
	return &MsgSubmit{
		Challenger:        challenger.String(),
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Challenger)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmit) GetSignBytes() []byte {
	panic("GetSignBytes")
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmit) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Challenger)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid challenger address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp operator address (%s)", err)
	}

	if err = s3util.CheckValidBucketName(msg.BucketName); err != nil {
		return err
	}

	if err = s3util.CheckValidObjectName(msg.ObjectName); err != nil {
		return err
	}

	return nil
}
