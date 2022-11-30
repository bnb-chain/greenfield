package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// For bucket
	TypeMsgCreateBucket = "create_bucket"
	TypeMsgDeleteBucket = "delete_bucket"

	// For object
	TypeMsgCopyObject           = "copy_object"
	TypeMsgCreateObject         = "create_object"
	TypeMsgDeleteObject         = "delete_object"
	TypeMsgSealObject           = "seal_object"
	TypeMsgRejectUnsealedObject = "reject_unsealed_object"

	// For group
	TypeMsgCreateGroup       = "create_group"
	TypeMsgDeleteGroup       = "delete_group"
	TypeMsgLeaveGroup        = "leave_group"
	TypeMsgUpdateGroupMember = "update_group_member"
)

var (
	// For bucket
	_ sdk.Msg = &MsgCreateBucket{}
	_ sdk.Msg = &MsgDeleteBucket{}
	// For object
	_ sdk.Msg = &MsgCreateObject{}
	_ sdk.Msg = &MsgDeleteObject{}
	_ sdk.Msg = &MsgSealObject{}
	_ sdk.Msg = &MsgCopyObject{}
	_ sdk.Msg = &MsgRejectUnsealedObject{}

	// For group
	_ sdk.Msg = &MsgCreateGroup{}
	_ sdk.Msg = &MsgDeleteGroup{}
	_ sdk.Msg = &MsgLeaveGroup{}
	_ sdk.Msg = &MsgUpdateGroupMember{}
)

// NewMsgCreateBucket creates a new MsgCreateBucket instance.
func NewMsgCreateBucket(
	creator sdk.AccAddress, bucketName string, isPublic bool,
	primarySPAddress sdk.AccAddress, paymentAddress sdk.AccAddress, primarySPSignature []byte) *MsgCreateBucket {
	return &MsgCreateBucket{
		Creator:            creator.String(),
		BucketName:         bucketName,
		IsPublic:           isPublic,
		PaymentAddress:     paymentAddress.String(),
		PrimarySpAddress:   primarySPAddress.String(),
		PrimarySpSignature: primarySPSignature,
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

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCreateBucket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if _, err := sdk.AccAddressFromHexUnsafe(msg.PrimarySpAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid primary sp address (%s)", err)
	}

	// PaymentAddress is optional, use creator by default if not set.
	if msg.PaymentAddress != "" {
		if _, err := sdk.AccAddressFromHexUnsafe(msg.PaymentAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid store payment address (%s)", err)
		}
	}

	if err := CheckValidBucketName(msg.BucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid bucket name (%s)", err)
	}
	return nil
}

// NewMsgDeleteBucket creates a new MsgDeleteBucket instance
func NewMsgDeleteBucket(creator string, bucketName string) *MsgDeleteBucket {
	return &MsgDeleteBucket{
		Creator:    creator,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
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
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidBucketName(msg.BucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid bucket name (%s)", err)
	}

	return nil
}

// NewMsgCreateObject creates a new MsgCreateObject instance.
func NewMsgCreateObject(
	creator sdk.AccAddress, bucketName string, objectName string, payloadSize uint64,
	isPublic bool, expectChecksum [][]byte, contentType string, primarySPSignature []byte,
	secondarySPs []sdk.AccAddress) *MsgCreateObject {
	var secondarySPAddresses []string
	if secondarySPs != nil {
		for _, secondarySP := range secondarySPs {
			secondarySPAddresses = append(secondarySPAddresses, secondarySP.String())
		}
	}

	return &MsgCreateObject{
		Creator:                    creator.String(),
		BucketName:                 bucketName,
		ObjectName:                 objectName,
		PayloadSize:                payloadSize,
		IsPublic:                   isPublic,
		ContentType:                contentType,
		PrimarySpSignature:         primarySPSignature,
		ExpectChecksum:             expectChecksum,
		ExpectSecondarySpAddresses: secondarySPAddresses,
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
	if _, err := sdk.AccAddressFromHexUnsafe(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidBucketName(msg.BucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid bucket name (%s)", err)
	}

	if err := CheckValidObjectName(msg.ObjectName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidObjectName, "invalid object name (%s)", err)
	}

	if err := CheckValidExpectChecksums(msg.ExpectChecksum); err != nil {
		return sdkerrors.Wrapf(ErrInvalidChcecksum, "invalid checksum (%s)", err)
	}

	if err := CheckValidContentType(msg.ContentType); err != nil {
		return sdkerrors.Wrapf(ErrInvalidContentType, "invalid checksum (%s)", err)
	}

	if msg.PrimarySpSignature == nil {
		return sdkerrors.Wrapf(ErrInvalidSPSignature, "empty sp signature")
	}

	for _, spAddress := range msg.ExpectSecondarySpAddresses {
		if _, err := sdk.AccAddressFromHexUnsafe(spAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp address (%s) in expect secondary SPs", err)
		}
	}
	return nil
}

func NewMsgDeleteObject(creator string, bucketName string, objectName string) *MsgDeleteObject {
	return &MsgDeleteObject{
		Creator:    creator,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
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
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidBucketName(msg.BucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid bucket name (%s)", err)
	}

	if err := CheckValidObjectName(msg.ObjectName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidObjectName, "invalid object name (%s)", err)
	}
	return nil
}

func NewMsgSealObject(creator sdk.AccAddress, bucketName string, objectName string, spSignatures [][]byte) *MsgSealObject {
	return &MsgSealObject{
		Creator:      creator.String(),
		BucketName:   bucketName,
		ObjectName:   objectName,
		SpSignatures: spSignatures,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
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
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if _, err := sdk.AccAddressFromHexUnsafe(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidBucketName(msg.BucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid bucket name (%s)", err)
	}

	if err := CheckValidObjectName(msg.ObjectName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidObjectName, "invalid object name (%s)", err)
	}

	if len(msg.SpSignatures) != 7 {
		return sdkerrors.Wrapf(ErrInvalidSPSignature, "Missing SP signatures")
	}

	for _, sig := range msg.SpSignatures {
		if sig == nil {
			return sdkerrors.Wrapf(ErrInvalidSPSignature, "Empty SP signatures")
		}
	}

	return nil
}

func NewMsgCopyObject(
	creator string, srcBucketName string, dstBucketName string,
	srcObjectName string, dstObjectName string) *MsgCopyObject {
	return &MsgCopyObject{
		Creator:       creator,
		SrcBucketName: srcBucketName,
		DstBucketName: dstBucketName,
		SrcObjectName: srcObjectName,
		DstObjectName: dstObjectName,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCopyObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCopyObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidBucketName(msg.SrcBucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid src bucket name (%s)", err)
	}

	if err := CheckValidObjectName(msg.SrcObjectName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidObjectName, "invalid src object name (%s)", err)
	}

	if err := CheckValidBucketName(msg.DstBucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid src bucket name (%s)", err)
	}

	if err := CheckValidObjectName(msg.DstObjectName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidObjectName, "invalid src object name (%s)", err)
	}
	return nil
}

func NewMsgRejectUnsealedObject(creator string, bucketName string, objectName string) *MsgRejectUnsealedObject {
	return &MsgRejectUnsealedObject{
		Creator:    creator,
		BucketName: bucketName,
		ObjectName: objectName,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgRejectUnsealedObject) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgRejectUnsealedObject) Type() string {
	return TypeMsgRejectUnsealedObject
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgRejectUnsealedObject) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgRejectUnsealedObject) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgRejectUnsealedObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

func NewMsgCreateGroup(creator string, groupName string) *MsgCreateGroup {
	return &MsgCreateGroup{
		Creator:   creator,
		GroupName: groupName,
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidGroupName(msg.GroupName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidGroupName, "invalid groupName (%s)", err)
	}
	return nil
}

func NewMsgDeleteGroup(creator string, groupName string) *MsgDeleteGroup {
	return &MsgDeleteGroup{
		Creator:   creator,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
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
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidGroupName(msg.GroupName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidGroupName, "invalid groupName (%s)", err)
	}
	return nil
}

func NewMsgLeaveGroup(creator string, groupName string) *MsgLeaveGroup {
	return &MsgLeaveGroup{
		Creator:   creator,
		GroupName: groupName,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
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
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidGroupName(msg.GroupName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidGroupName, "invalid groupName (%s)", err)
	}
	return nil
}

func NewMsgUpdateGroupMember(creator string, groupName string) *MsgUpdateGroupMember {
	return &MsgUpdateGroupMember{
		Creator:   creator,
		GroupName: groupName,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
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
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
