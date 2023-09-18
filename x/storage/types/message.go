package types

import (
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gogoproto/proto"

	grn2 "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/common"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/types/s3util"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
)

const (
	// For bucket
	TypeMsgCreateBucket     = "create_bucket"
	TypeMsgDeleteBucket     = "delete_bucket"
	TypeMsgUpdateBucketInfo = "update_bucket_info"
	TypeMsgMirrorBucket     = "mirror_bucket"

	// For object
	TypeMsgCopyObject         = "copy_object"
	TypeMsgCreateObject       = "create_object"
	TypeMsgDeleteObject       = "delete_object"
	TypeMsgSealObject         = "seal_object"
	TypeMsgRejectSealObject   = "reject_seal_object"
	TypeMsgCancelCreateObject = "cancel_create_object"
	TypeMsgMirrorObject       = "mirror_object"
	TypeMsgDiscontinueObject  = "discontinue_object"
	TypeMsgDiscontinueBucket  = "discontinue_bucket"
	TypeMsgUpdateObjectInfo   = "update_object_info"

	// For group
	TypeMsgCreateGroup       = "create_group"
	TypeMsgDeleteGroup       = "delete_group"
	TypeMsgLeaveGroup        = "leave_group"
	TypeMsgUpdateGroupMember = "update_group_member"
	TypeMsgUpdateGroupExtra  = "update_group_extra"
	TypeMsgMirrorGroup       = "mirror_group"
	TypeMsgRenewGroupMember  = "renew_group_member"

	MaxGroupExtraInfoLimit = 512

	// For permission policy
	TypeMsgPutPolicy    = "put_policy"
	TypeMsgDeletePolicy = "delete_policy"

	MaxGroupMemberLimitOnce = 20

	// For discontinue
	MaxDiscontinueReasonLen = 128
	MaxDiscontinueObjects   = 128
)

var (
	// For bucket
	_ sdk.Msg = &MsgCreateBucket{}
	_ sdk.Msg = &MsgDeleteBucket{}
	_ sdk.Msg = &MsgUpdateBucketInfo{}
	_ sdk.Msg = &MsgMirrorBucket{}
	_ sdk.Msg = &MsgDiscontinueBucket{}

	// For object
	_ sdk.Msg = &MsgCreateObject{}
	_ sdk.Msg = &MsgSealObject{}
	_ sdk.Msg = &MsgRejectSealObject{}
	_ sdk.Msg = &MsgCopyObject{}
	_ sdk.Msg = &MsgDeleteObject{}
	_ sdk.Msg = &MsgCancelCreateObject{}
	_ sdk.Msg = &MsgMirrorObject{}
	_ sdk.Msg = &MsgDiscontinueObject{}
	_ sdk.Msg = &MsgUpdateObjectInfo{}

	// For group
	_ sdk.Msg = &MsgCreateGroup{}
	_ sdk.Msg = &MsgDeleteGroup{}
	_ sdk.Msg = &MsgUpdateGroupMember{}
	_ sdk.Msg = &MsgUpdateGroupExtra{}
	_ sdk.Msg = &MsgLeaveGroup{}
	_ sdk.Msg = &MsgMirrorGroup{}

	// For permission policy
	_ sdk.Msg = &MsgPutPolicy{}
	_ sdk.Msg = &MsgDeletePolicy{}

	// For params
	_ sdk.Msg = &MsgUpdateParams{}

	// The max timestamp in underlying package `google.golang.org/protobuf/types/known/timestamppb` is 9999-12-31T23:59:59Z
	// https://pkg.go.dev/google.golang.org/protobuf/types/known/timestamppb#Timestamp
	MaxTimeStamp, _ = time.Parse(time.RFC3339, "9999-12-31T23:59:59Z")
)

// NewMsgCreateBucket creates a new MsgCreateBucket instance.
func NewMsgCreateBucket(
	creator sdk.AccAddress, bucketName string, Visibility VisibilityType,
	primarySPAddress sdk.AccAddress, paymentAddress sdk.AccAddress, timeoutHeight uint64, sig []byte, chargedReadQuota uint64) *MsgCreateBucket {
	return &MsgCreateBucket{
		Creator:           creator.String(),
		BucketName:        bucketName,
		Visibility:        Visibility,
		PaymentAddress:    paymentAddress.String(),
		PrimarySpAddress:  primarySPAddress.String(),
		PrimarySpApproval: &common.Approval{ExpiredHeight: timeoutHeight, Sig: sig},
		ChargedReadQuota:  chargedReadQuota,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgCreateBucket) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgCreateBucket) Type() string {
	return TypeMsgCreateBucket
}

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
func (msg *MsgCreateBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgCreateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetApprovalBytes returns the message bytes of approval info.
func (msg *MsgCreateBucket) GetApprovalBytes() []byte {
	fakeMsg := proto.Clone(msg).(*MsgCreateBucket)
	fakeMsg.PrimarySpApproval.Sig = []byte{}
	return fakeMsg.GetSignBytes()
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCreateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.PrimarySpAddress)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid primary sp address (%s)", err)
	}

	if msg.PrimarySpApproval == nil {
		return ErrInvalidApproval.Wrap("Empty approvals are not allowed.")
	}

	// PaymentAddress is optional, use creator by default if not set.
	if msg.PaymentAddress != "" {
		if _, err := sdk.AccAddressFromHexUnsafe(msg.PaymentAddress); err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid store payment address (%s)", err)
		}
	}

	if msg.Visibility == VISIBILITY_TYPE_UNSPECIFIED {
		return errors.Wrapf(ErrInvalidVisibility, "Unspecified visibility is not allowed.")
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}
	return nil
}

// NewMsgDeleteBucket creates a new MsgDeleteBucket instance
func NewMsgDeleteBucket(operator sdk.AccAddress, bucketName string) *MsgDeleteBucket {
	return &MsgDeleteBucket{
		Operator:   operator.String(),
		BucketName: bucketName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgDeleteBucket) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgDeleteBucket) Type() string {
	return TypeMsgDeleteBucket
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgDeleteBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg *MsgDeleteBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgDeleteBucket) ValidateBasic() error {
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

// NewMsgUpdateBucketInfo creates a new MsgBucketReadQuota instance.
func NewMsgUpdateBucketInfo(operator sdk.AccAddress, bucketName string, chargedReadQuota *uint64, paymentAcc sdk.AccAddress, visibility VisibilityType) *MsgUpdateBucketInfo {
	msgUpdateBucketInfo := &MsgUpdateBucketInfo{
		Operator:   operator.String(),
		BucketName: bucketName,
		Visibility: visibility,
	}
	if paymentAcc != nil {
		msgUpdateBucketInfo.PaymentAddress = paymentAcc.String()
	}
	if chargedReadQuota != nil {
		msgUpdateBucketInfo.ChargedReadQuota = &common.UInt64Value{Value: *chargedReadQuota}
	}

	return msgUpdateBucketInfo
}

func (msg *MsgUpdateBucketInfo) Route() string {
	return RouterKey
}

func (msg *MsgUpdateBucketInfo) Type() string {
	return TypeMsgUpdateBucketInfo
}

func (msg *MsgUpdateBucketInfo) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateBucketInfo) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateBucketInfo) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err = s3util.CheckValidBucketName(msg.BucketName); err != nil {
		return err
	}

	if msg.PaymentAddress != "" {
		_, err = sdk.AccAddressFromHexUnsafe(msg.PaymentAddress)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid payment address (%s)", err)
		}
	}

	return nil
}

// NewMsgCreateObject creates a new MsgCreateObject instance.
func NewMsgCreateObject(
	creator sdk.AccAddress, bucketName string, objectName string, payloadSize uint64,
	Visibility VisibilityType, expectChecksums [][]byte, contentType string, redundancyType RedundancyType, timeoutHeight uint64, sig []byte) *MsgCreateObject {

	return &MsgCreateObject{
		Creator:           creator.String(),
		BucketName:        bucketName,
		ObjectName:        objectName,
		PayloadSize:       payloadSize,
		Visibility:        Visibility,
		ContentType:       contentType,
		PrimarySpApproval: &common.Approval{ExpiredHeight: timeoutHeight, Sig: sig},
		ExpectChecksums:   expectChecksums,
		RedundancyType:    redundancyType,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgCreateObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgCreateObject) Type() string {
	return TypeMsgCreateObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgCreateObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgCreateObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCreateObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.PrimarySpApproval == nil {
		return errors.Wrapf(ErrInvalidApproval, "Empty approvals are not allowed.")
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.ObjectName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidExpectChecksums(msg.ExpectChecksums)
	if err != nil {
		return err
	}

	err = s3util.CheckValidContentType(msg.ContentType)
	if err != nil {
		return err
	}

	if msg.Visibility == VISIBILITY_TYPE_UNSPECIFIED {
		return errors.Wrapf(ErrInvalidVisibility, "Unspecified visibility is not allowed.")
	}
	return nil
}

// GetApprovalBytes returns the message bytes of approval info.
func (msg *MsgCreateObject) GetApprovalBytes() []byte {
	fakeMsg := proto.Clone(msg).(*MsgCreateObject)
	fakeMsg.PrimarySpApproval.Sig = []byte{}
	return fakeMsg.GetSignBytes()
}

func NewMsgCancelCreateObject(operator sdk.AccAddress, bucketName string, objectName string) *MsgCancelCreateObject {
	return &MsgCancelCreateObject{
		Operator:   operator.String(),
		BucketName: bucketName,
		ObjectName: objectName,
	}
}

func (msg *MsgCancelCreateObject) Route() string {
	return RouterKey
}

func (msg *MsgCancelCreateObject) Type() string {
	return TypeMsgCancelCreateObject
}

func (msg *MsgCancelCreateObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCancelCreateObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCancelCreateObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.ObjectName)
	if err != nil {
		return err
	}

	return nil
}

func NewMsgDeleteObject(operator sdk.AccAddress, bucketName string, objectName string) *MsgDeleteObject {
	return &MsgDeleteObject{
		Operator:   operator.String(),
		BucketName: bucketName,
		ObjectName: objectName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgDeleteObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgDeleteObject) Type() string {
	return TypeMsgDeleteObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgDeleteObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgDeleteObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgDeleteObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.ObjectName)
	if err != nil {
		return err
	}
	return nil
}

func NewMsgSealObject(
	operator sdk.AccAddress, bucketName string, objectName string, globalVirtualGroupID uint32,
	secondarySpBlsSignatures []byte) *MsgSealObject {

	return &MsgSealObject{
		Operator:                    operator.String(),
		BucketName:                  bucketName,
		ObjectName:                  objectName,
		GlobalVirtualGroupId:        globalVirtualGroupID,
		SecondarySpBlsAggSignatures: secondarySpBlsSignatures,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgSealObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgSealObject) Type() string {
	return TypeMsgSealObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgSealObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgSealObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgSealObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.ObjectName)
	if err != nil {
		return err
	}

	if len(msg.GetSecondarySpBlsAggSignatures()) != sdk.BLSSignatureLength {
		return errors.Wrap(gnfderrors.ErrInvalidBlsSignature,
			fmt.Sprintf("length of signature should be %d", sdk.BLSSignatureLength),
		)
	}

	return nil
}

func NewMsgCopyObject(
	operator sdk.AccAddress, srcBucketName string, dstBucketName string,
	srcObjectName string, dstObjectName string, timeoutHeight uint64, sig []byte) *MsgCopyObject {
	return &MsgCopyObject{
		Operator:             operator.String(),
		SrcBucketName:        srcBucketName,
		DstBucketName:        dstBucketName,
		SrcObjectName:        srcObjectName,
		DstObjectName:        dstObjectName,
		DstPrimarySpApproval: &common.Approval{ExpiredHeight: timeoutHeight, Sig: sig},
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgCopyObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgCopyObject) Type() string {
	return TypeMsgCopyObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgCopyObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCopyObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetApprovalBytes returns the message bytes of approval info.
func (msg *MsgCopyObject) GetApprovalBytes() []byte {
	fakeMsg := proto.Clone(msg).(*MsgCopyObject)
	fakeMsg.DstPrimarySpApproval.Sig = []byte{}
	return fakeMsg.GetSignBytes()
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCopyObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.DstPrimarySpApproval == nil {
		return errors.Wrapf(ErrInvalidApproval, "Empty approvals are not allowed.")
	}

	err = s3util.CheckValidBucketName(msg.SrcBucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.SrcObjectName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidBucketName(msg.DstBucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.DstObjectName)
	if err != nil {
		return err
	}
	return nil
}

func NewMsgRejectUnsealedObject(operator sdk.AccAddress, bucketName string, objectName string) *MsgRejectSealObject {
	return &MsgRejectSealObject{
		Operator:   operator.String(),
		BucketName: bucketName,
		ObjectName: objectName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgRejectSealObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgRejectSealObject) Type() string {
	return TypeMsgRejectSealObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgRejectSealObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgRejectSealObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgRejectSealObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.ObjectName)
	if err != nil {
		return err
	}

	return nil
}

func NewMsgDiscontinueObject(operator sdk.AccAddress, bucketName string, objectIds []Uint, reason string) *MsgDiscontinueObject {
	return &MsgDiscontinueObject{
		Operator:   operator.String(),
		BucketName: bucketName,
		ObjectIds:  objectIds,
		Reason:     strings.TrimSpace(reason),
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgDiscontinueObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgDiscontinueObject) Type() string {
	return TypeMsgDiscontinueObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgDiscontinueObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgDiscontinueObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgDiscontinueObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	if len(msg.ObjectIds) == 0 || len(msg.ObjectIds) > MaxDiscontinueObjects {
		return errors.Wrapf(ErrInvalidObjectIds, "length of ids is %d", len(msg.ObjectIds))
	}

	if len(msg.Reason) > MaxDiscontinueReasonLen {
		return errors.Wrapf(ErrInvalidReason, "reason is too long with length %d", len(msg.Reason))
	}

	return nil
}

func NewMsgDiscontinueBucket(operator sdk.AccAddress, bucketName string, reason string) *MsgDiscontinueBucket {
	return &MsgDiscontinueBucket{
		Operator:   operator.String(),
		BucketName: bucketName,
		Reason:     strings.TrimSpace(reason),
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgDiscontinueBucket) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgDiscontinueBucket) Type() string {
	return TypeMsgDiscontinueBucket
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgDiscontinueBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgDiscontinueBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgDiscontinueBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	if len(msg.Reason) > MaxDiscontinueReasonLen {
		return errors.Wrapf(ErrInvalidReason, "reason is too long with length %d", len(msg.Reason))
	}

	return nil
}

func NewMsgUpdateObjectInfo(
	operator sdk.AccAddress, bucketName string, objectName string,
	visibility VisibilityType) *MsgUpdateObjectInfo {
	return &MsgUpdateObjectInfo{
		Operator:   operator.String(),
		BucketName: bucketName,
		ObjectName: objectName,
		Visibility: visibility,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgUpdateObjectInfo) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgUpdateObjectInfo) Type() string {
	return TypeMsgUpdateObjectInfo
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgUpdateObjectInfo) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateObjectInfo) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgUpdateObjectInfo) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.ObjectName)
	if err != nil {
		return err
	}

	if msg.Visibility == VISIBILITY_TYPE_UNSPECIFIED {
		return errors.Wrapf(ErrInvalidVisibility, "Unspecified visibility is not allowed.")
	}

	return nil
}

func NewMsgCreateGroup(creator sdk.AccAddress, groupName string, extra string) *MsgCreateGroup {
	return &MsgCreateGroup{
		Creator:   creator.String(),
		GroupName: groupName,
		Extra:     extra,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgCreateGroup) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgCreateGroup) Type() string {
	return TypeMsgCreateGroup
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgCreateGroup) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgCreateGroup) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCreateGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	err = s3util.CheckValidGroupName(msg.GroupName)
	if err != nil {
		return gnfderrors.ErrInvalidGroupName.Wrapf("invalid groupName (%s)", err)
	}

	if len(msg.Extra) > MaxGroupExtraInfoLimit {
		return errors.Wrapf(gnfderrors.ErrInvalidParameter, "extra is too long with length %d, limit to %d", len(msg.Extra), MaxGroupExtraInfoLimit)
	}
	return nil
}

func NewMsgDeleteGroup(operator sdk.AccAddress, groupName string) *MsgDeleteGroup {
	return &MsgDeleteGroup{
		Operator:  operator.String(),
		GroupName: groupName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgDeleteGroup) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgDeleteGroup) Type() string {
	return TypeMsgDeleteGroup
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgDeleteGroup) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgDeleteGroup) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgDeleteGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	err = s3util.CheckValidGroupName(msg.GroupName)
	if err != nil {
		return errors.Wrapf(gnfderrors.ErrInvalidGroupName, "invalid groupName (%s)", err)
	}
	return nil
}

func NewMsgLeaveGroup(member sdk.AccAddress, groupOwner sdk.AccAddress, groupName string) *MsgLeaveGroup {
	return &MsgLeaveGroup{
		Member:     member.String(),
		GroupOwner: groupOwner.String(),
		GroupName:  groupName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgLeaveGroup) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgLeaveGroup) Type() string {
	return TypeMsgLeaveGroup
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgLeaveGroup) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Member)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgLeaveGroup) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgLeaveGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Member)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.GroupOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid group owner (%s)", err)
	}

	err = s3util.CheckValidGroupName(msg.GroupName)
	if err != nil {
		return err
	}
	return nil
}

func NewMsgUpdateGroupMember(
	operator sdk.AccAddress, groupOwner sdk.AccAddress, groupName string, membersToAdd []*MsgGroupMember,
	membersToDelete []sdk.AccAddress) *MsgUpdateGroupMember {
	var membersAddrToDelete []string
	for _, member := range membersToDelete {
		membersAddrToDelete = append(membersAddrToDelete, member.String())
	}
	return &MsgUpdateGroupMember{
		Operator:        operator.String(),
		GroupOwner:      groupOwner.String(),
		GroupName:       groupName,
		MembersToAdd:    membersToAdd,
		MembersToDelete: membersAddrToDelete,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgUpdateGroupMember) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgUpdateGroupMember) Type() string {
	return TypeMsgUpdateGroupMember
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgUpdateGroupMember) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgUpdateGroupMember) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgUpdateGroupMember) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.GroupOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid group owner address (%s)", err)
	}

	err = s3util.CheckValidGroupName(msg.GroupName)
	if err != nil {
		return err
	}

	if len(msg.MembersToAdd)+len(msg.MembersToDelete) > MaxGroupMemberLimitOnce {
		return gnfderrors.ErrInvalidParameter.Wrapf("Once update group member limit exceeded")
	}
	for _, member := range msg.MembersToAdd {
		_, err = sdk.AccAddressFromHexUnsafe(member.Member)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid member address (%s)", err)
		}
		if member.ExpirationTime != nil && member.ExpirationTime.UTC().After(MaxTimeStamp) {
			return gnfderrors.ErrInvalidParameter.Wrapf("Expiration time is bigger than max timestamp [%s]", MaxTimeStamp)
		}

	}
	for _, member := range msg.MembersToDelete {
		_, err = sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid member address (%s)", err)
		}
	}
	return nil
}

func NewMsgUpdateGroupExtra(operator sdk.AccAddress, groupOwner sdk.AccAddress, groupName, extra string) *MsgUpdateGroupExtra {
	return &MsgUpdateGroupExtra{
		Operator:   operator.String(),
		GroupOwner: groupOwner.String(),
		GroupName:  groupName,
		Extra:      extra,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgUpdateGroupExtra) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgUpdateGroupExtra) Type() string {
	return TypeMsgUpdateGroupExtra
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgUpdateGroupExtra) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgUpdateGroupExtra) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgUpdateGroupExtra) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.GroupOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid group owner address (%s)", err)
	}

	err = s3util.CheckValidGroupName(msg.GroupName)
	if err != nil {
		return err
	}
	if len(msg.Extra) > MaxGroupExtraInfoLimit {
		return errors.Wrapf(gnfderrors.ErrInvalidParameter, "extra is too long with length %d, limit to %d", len(msg.Extra), MaxGroupExtraInfoLimit)
	}

	return nil
}

func NewMsgPutPolicy(operator sdk.AccAddress, resource string,
	principal *permtypes.Principal, statements []*permtypes.Statement, expirationTime *time.Time) *MsgPutPolicy {
	return &MsgPutPolicy{
		Operator:       operator.String(),
		Resource:       resource,
		Principal:      principal,
		Statements:     statements,
		ExpirationTime: expirationTime,
	}
}

func (msg *MsgPutPolicy) Route() string {
	return RouterKey
}

func (msg *MsgPutPolicy) Type() string {
	return TypeMsgPutPolicy
}

func (msg *MsgPutPolicy) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgPutPolicy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPutPolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	var grn grn2.GRN
	err = grn.ParseFromString(msg.Resource, true)
	if err != nil {
		return errors.Wrapf(gnfderrors.ErrInvalidGRN, "invalid greenfield resource name (%s)", err)
	}

	if msg.Principal == nil {
		return gnfderrors.ErrInvalidPrincipal.Wrapf("principal cannot be empty")
	}

	if msg.Principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_GROUP && grn.ResourceType() == resource.RESOURCE_TYPE_GROUP {
		return gnfderrors.ErrInvalidPrincipal.Wrapf("Not allow grant group's permission to another group")
	}

	err = msg.Principal.ValidateBasic()
	if err != nil {
		return err
	}

	for _, s := range msg.Statements {
		err = s.ValidateBasic(grn.ResourceType())
		if err != nil {
			return err
		}
	}

	return nil
}

func (msg *MsgPutPolicy) ValidateRuntime(ctx sdk.Context) error {
	var grn grn2.GRN
	_ = grn.ParseFromString(msg.Resource, true) // no error after ValidateBasic
	for _, s := range msg.Statements {
		err := s.ValidateRuntime(ctx, grn.ResourceType())
		if err != nil {
			return err
		}
	}

	return nil
}

func NewMsgDeletePolicy(operator sdk.AccAddress, resource string, principal *permtypes.Principal) *MsgDeletePolicy {
	return &MsgDeletePolicy{
		Operator:  operator.String(),
		Resource:  resource,
		Principal: principal,
	}
}

func (msg *MsgDeletePolicy) Route() string {
	return RouterKey
}

func (msg *MsgDeletePolicy) Type() string {
	return TypeMsgDeletePolicy
}

func (msg *MsgDeletePolicy) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeletePolicy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeletePolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	var grn grn2.GRN
	err = grn.ParseFromString(msg.Resource, false)
	if err != nil {
		return errors.Wrapf(gnfderrors.ErrInvalidGRN, "invalid greenfield resource name (%s)", err)
	}

	if msg.Principal == nil {
		return gnfderrors.ErrInvalidPrincipal.Wrapf("principal cannot be empty")
	}

	if msg.Principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_GROUP && grn.ResourceType() == resource.RESOURCE_TYPE_GROUP {
		return gnfderrors.ErrInvalidPrincipal.Wrapf("Not allow grant group's permission to another group")
	}

	return nil
}

func (msg *MsgDeletePolicy) ValidateRuntime(ctx sdk.Context) error {
	if ctx.IsUpgraded(upgradetypes.Xxxxx) {
		if err := msg.Principal.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

// NewMsgMirrorBucket creates a new MsgMirrorBucket instance
func NewMsgMirrorBucket(operator sdk.AccAddress, destChainId sdk.ChainID, id Uint, bucketName string) *MsgMirrorBucket {
	return &MsgMirrorBucket{
		Operator:    operator.String(),
		Id:          id,
		BucketName:  bucketName,
		DestChainId: uint32(destChainId),
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgMirrorBucket) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgMirrorBucket) Type() string {
	return TypeMsgMirrorBucket
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgMirrorBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg *MsgMirrorBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgMirrorBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !msg.Id.IsNil() && msg.Id.GT(sdk.NewUint(0)) {
		if msg.BucketName != "" {
			return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name should be empty")
		}
		return nil
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	return nil
}

// NewMsgMirrorObject creates a new MsgMirrorObject instance
func NewMsgMirrorObject(operator sdk.AccAddress, destChainId sdk.ChainID, id Uint, bucketName, objectName string) *MsgMirrorObject {
	return &MsgMirrorObject{
		Operator:    operator.String(),
		DestChainId: uint32(destChainId),
		Id:          id,
		BucketName:  bucketName,
		ObjectName:  objectName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgMirrorObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgMirrorObject) Type() string {
	return TypeMsgMirrorObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgMirrorObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgMirrorObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgMirrorObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !msg.Id.IsNil() && msg.Id.GT(sdk.NewUint(0)) {
		if msg.BucketName != "" {
			return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name should be empty")
		}
		if msg.ObjectName != "" {
			return errors.Wrap(gnfderrors.ErrInvalidObjectName, "Object name should be empty")
		}
		return nil
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	err = s3util.CheckValidObjectName(msg.ObjectName)
	if err != nil {
		return err
	}

	return nil
}

// NewMsgMirrorGroup creates a new MsgMirrorGroup instance
func NewMsgMirrorGroup(operator sdk.AccAddress, destChainId sdk.ChainID, id Uint, groupName string) *MsgMirrorGroup {
	return &MsgMirrorGroup{
		Operator:    operator.String(),
		DestChainId: uint32(destChainId),
		Id:          id,
		GroupName:   groupName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgMirrorGroup) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgMirrorGroup) Type() string {
	return TypeMsgMirrorGroup
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgMirrorGroup) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgMirrorGroup) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgMirrorGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !msg.Id.IsNil() && msg.Id.GT(sdk.NewUint(0)) {
		if msg.GroupName != "" {
			return errors.Wrap(gnfderrors.ErrInvalidGroupName, "Group name should be empty")
		}
		return nil
	}

	err = s3util.CheckValidGroupName(msg.GroupName)
	if err != nil {
		return gnfderrors.ErrInvalidGroupName.Wrapf("invalid groupName (%s)", err)
	}

	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(m.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

func NewMsgRenewGroupMember(
	operator sdk.AccAddress, groupOwner sdk.AccAddress, groupName string, members []*MsgGroupMember) *MsgRenewGroupMember {

	return &MsgRenewGroupMember{
		Operator:   operator.String(),
		GroupOwner: groupOwner.String(),
		GroupName:  groupName,
		Members:    members,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgRenewGroupMember) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgRenewGroupMember) Type() string {
	return TypeMsgRenewGroupMember
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgRenewGroupMember) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgRenewGroupMember) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgRenewGroupMember) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.GroupOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid group owner address (%s)", err)
	}

	err = s3util.CheckValidGroupName(msg.GroupName)
	if err != nil {
		return err
	}

	if len(msg.Members) > MaxGroupMemberLimitOnce {
		return gnfderrors.ErrInvalidParameter.Wrapf("Once renew group member limit exceeded")
	}
	for _, member := range msg.Members {
		_, err = sdk.AccAddressFromHexUnsafe(member.Member)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid member address (%s)", err)
		}
		if member.ExpirationTime != nil && member.ExpirationTime.UTC().After(MaxTimeStamp) {
			return gnfderrors.ErrInvalidParameter.Wrapf("Expiration time is bigger than max timestamp [%s]", MaxTimeStamp)
		}
	}

	return nil
}
