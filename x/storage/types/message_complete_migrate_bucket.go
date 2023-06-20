package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bnb-chain/greenfield/types/s3util"
)

const TypeMsgCompleteMigrateBucket = "complete_migrate_bucket"

var _ sdk.Msg = &MsgCompleteMigrateBucket{}

func NewMsgCompleteMigrateBucket(operator string) *MsgCompleteMigrateBucket {
	return &MsgCompleteMigrateBucket{
		Operator: operator,
	}
}

func (msg *MsgCompleteMigrateBucket) Route() string {
	return RouterKey
}

func (msg *MsgCompleteMigrateBucket) Type() string {
	return TypeMsgCompleteMigrateBucket
}

func (msg *MsgCompleteMigrateBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCompleteMigrateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCompleteMigrateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}
	return nil
}
