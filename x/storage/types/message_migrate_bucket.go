package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bnb-chain/greenfield/types/s3util"
)

const TypeMsgMigrateBucket = "migrate_bucket"

var _ sdk.Msg = &MsgMigrateBucket{}

func NewMsgMigrateBucket(operator string) *MsgMigrateBucket {
	return &MsgMigrateBucket{
		Operator: operator,
	}
}

func (msg *MsgMigrateBucket) Route() string {
	return RouterKey
}

func (msg *MsgMigrateBucket) Type() string {
	return TypeMsgMigrateBucket
}

func (msg *MsgMigrateBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgMigrateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMigrateBucket) GetApprovalBytes() []byte {
	fakeMsg := proto.Clone(msg).(*MsgMigrateBucket)
	fakeMsg.DstPrimarySpApproval.Sig = nil
	return fakeMsg.GetSignBytes()
}

func (msg *MsgMigrateBucket) ValidateBasic() error {
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
