package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bnb-chain/greenfield/types/s3util"
)

const TypeMsgCancelMigrateBucket = "cancel_migrate_bucket"

var _ sdk.Msg = &MsgCancelMigrateBucket{}

func NewMsgCancelMigrateBucket(operator sdk.AccAddress, bucketName string) *MsgCancelMigrateBucket {
	return &MsgCancelMigrateBucket{
		Operator:   operator.String(),
		BucketName: bucketName,
	}
}

func (msg *MsgCancelMigrateBucket) Route() string {
	return RouterKey
}

func (msg *MsgCancelMigrateBucket) Type() string {
	return TypeMsgCancelMigrateBucket
}

func (msg *MsgCancelMigrateBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCancelMigrateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCancelMigrateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid creator address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}
	return nil
}
