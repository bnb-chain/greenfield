package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bnb-chain/greenfield/types/common"
	"github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/types/s3util"
)

const TypeMsgMigrateBucket = "migrate_bucket"

var _ sdk.Msg = &MsgMigrateBucket{}

func NewMsgMigrateBucket(operator sdk.AccAddress, bucketName string, dstPrimarySPID uint32) *MsgMigrateBucket {
	return &MsgMigrateBucket{
		Operator:             operator.String(),
		BucketName:           bucketName,
		DstPrimarySpId:       dstPrimarySPID,
		DstPrimarySpApproval: &common.Approval{},
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

	if msg.DstPrimarySpId == 0 {
		return errors.ErrInvalidMessage.Wrapf("Invalid dst primary sp id: %d", msg.DstPrimarySpId)
	}

	if msg.DstPrimarySpApproval == nil {
		return ErrInvalidApproval.Wrap("Empty approvals are not allowed.")
	}
	return nil
}

func (msg *MsgMigrateBucket) ValidateRuntime(ctx sdk.Context) error {
	err := msg.ValidateBasic()
	if err != nil {
		return err
	}

	if ctx.IsUpgraded(upgradetypes.Ural) {
		err = s3util.CheckValidBucketNameByCharacterLength(msg.BucketName)
	} else {
		err = s3util.CheckValidBucketName(msg.BucketName)
	}
	if err != nil {
		return err
	}

	return nil
}
