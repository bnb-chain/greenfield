package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bnb-chain/greenfield/types/s3util"
)

const TypeMsgRejectMigrateBucket = "reject_migrate_bucket"

var _ sdk.Msg = &MsgRejectMigrateBucket{}

func NewMsgRejectMigrateBucket(operator sdk.AccAddress, bucketName string) *MsgRejectMigrateBucket {
	return &MsgRejectMigrateBucket{
		Operator:   operator.String(),
		BucketName: bucketName,
	}
}

func (msg *MsgRejectMigrateBucket) Route() string {
	return RouterKey
}

func (msg *MsgRejectMigrateBucket) Type() string {
	return TypeMsgRejectMigrateBucket
}

func (msg *MsgRejectMigrateBucket) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgRejectMigrateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRejectMigrateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid operator address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}
	return nil
}
