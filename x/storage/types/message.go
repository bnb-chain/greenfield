package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	// For bucket
	TypeMsgCreateBucket = "create_bucket"
	TypeMsgDeleteBucket = "delete_bucket"

	// For object
	TypeMsgCopyObject       = "copy_object"
	TypeMsgCreateObject     = "create_object"
	TypeMsgDeleteObject     = "delete_object"
	TypeMsgSealObject       = "seal_object"
	TypeMsgRejectSealObject = "reject_seal_object"

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
	_ sdk.Msg = &MsgRejectSealObject{}

	// For group
	_ sdk.Msg = &MsgCreateGroup{}
	_ sdk.Msg = &MsgDeleteGroup{}
	_ sdk.Msg = &MsgLeaveGroup{}
	_ sdk.Msg = &MsgUpdateGroupMember{}
)

// NewMsgCreateBucket creates a new MsgCreateBucket instance.
func NewMsgCreateBucket(
	creator sdk.AccAddress, bucketName string, isPublic bool,
	primarySPAddress sdk.AccAddress, paymentAddress sdk.AccAddress, primarySPApproval []byte) *MsgCreateBucket {
	return &MsgCreateBucket{
		Creator:                    creator.String(),
		BucketName:                 bucketName,
		IsPublic:                   isPublic,
		PaymentAddress:             paymentAddress.String(),
		PrimarySpAddress:           primarySPAddress.String(),
		PrimarySpApprovalSignature: primarySPApproval,
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
	isPublic bool, expectChecksums [][]byte, contentType string, primarySPApproval []byte,
	secondarySPAccs []sdk.AccAddress) *MsgCreateObject {

	var secSPAddrs []string
	for _, secondarySP := range secondarySPAccs {
		secSPAddrs = append(secSPAddrs, secondarySP.String())
	}

	return &MsgCreateObject{
		Creator:                    creator.String(),
		BucketName:                 bucketName,
		ObjectName:                 objectName,
		PayloadSize:                payloadSize,
		IsPublic:                   isPublic,
		ContentType:                contentType,
		PrimarySpApprovalSignature: primarySPApproval,
		ExpectChecksums:            expectChecksums,
		ExpectSecondarySpAddresses: secSPAddrs,
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

	if err := CheckValidExpectChecksums(msg.ExpectChecksums); err != nil {
		return sdkerrors.Wrapf(ErrInvalidChcecksum, "invalid checksum (%s)", err)
	}

	if err := CheckValidContentType(msg.ContentType); err != nil {
		return sdkerrors.Wrapf(ErrInvalidContentType, "invalid checksum (%s)", err)
	}

	if msg.PrimarySpApprovalSignature == nil {
		return sdkerrors.Wrapf(ErrInvalidSPSignature, "empty sp signature")
	}

	for _, spAddress := range msg.ExpectSecondarySpAddresses {
		if _, err := sdk.AccAddressFromHexUnsafe(spAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp address (%s) in expect secondary SPs", err)
		}
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

func NewMsgSealObject(
	operator sdk.AccAddress, bucketName string, objectName string,
	secondarySPAccs []sdk.AccAddress, secondarySpSignatures [][]byte) *MsgSealObject {

	var secondarySPAddresses []string
	for _, secondarySP := range secondarySPAccs {
		secondarySPAddresses = append(secondarySPAddresses, secondarySP.String())
	}

	return &MsgSealObject{
		Operator:              operator.String(),
		BucketName:            bucketName,
		ObjectName:            objectName,
		SecondarySpAddresses:  secondarySPAddresses,
		SecondarySpSignatures: secondarySpSignatures,
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if _, err := sdk.AccAddressFromHexUnsafe(msg.Operator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidBucketName(msg.BucketName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidBucketName, "invalid bucket name (%s)", err)
	}

	if err := CheckValidObjectName(msg.ObjectName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidObjectName, "invalid object name (%s)", err)
	}

	if len(msg.SecondarySpAddresses) != 6 {
		return sdkerrors.Wrapf(ErrInvalidSPAddress, "Missing SP expect: (d%), but (d%)", 6, len(msg.SecondarySpAddresses))
	}

	for _, addr := range msg.SecondarySpAddresses {
		_, err := sdk.AccAddressFromHexUnsafe(addr)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid secondary sp address (%s)", err)
		}
	}

	if len(msg.SecondarySpSignatures) != 6 {
		return sdkerrors.Wrapf(ErrInvalidSPSignature, "Missing SP signatures")
	}

	for _, sig := range msg.SecondarySpSignatures {
		if sig == nil && len(sig) != ethcrypto.SignatureLength {
			return sdkerrors.Wrapf(ErrInvalidSPSignature, "invalid SP signatures")
		}
	}

	return nil
}

func NewMsgCopyObject(
	operator sdk.AccAddress, srcBucketName string, dstBucketName string,
	srcObjectName string, dstObjectName string, dstPrimarySPApproval []byte) *MsgCopyObject {
	return &MsgCopyObject{
		Operator:                      operator.String(),
		SrcBucketName:                 srcBucketName,
		DstBucketName:                 dstBucketName,
		SrcObjectName:                 srcObjectName,
		DstObjectName:                 dstObjectName,
		DstPrimarySpApprovalSignature: dstPrimarySPApproval,
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

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCopyObject) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

func NewMsgCreateGroup(creator sdk.AccAddress, groupName string, membersAcc []sdk.AccAddress) *MsgCreateGroup {
	var members []string
	for _, member := range membersAcc {
		members = append(members, member.String())
	}
	return &MsgCreateGroup{
		Creator:   creator.String(),
		GroupName: groupName,
		Members:   members,
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidGroupName(msg.GroupName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidGroupName, "invalid groupName (%s)", err)
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := CheckValidGroupName(msg.GroupName); err != nil {
		return sdkerrors.Wrapf(ErrInvalidGroupName, "invalid groupName (%s)", err)
	}
	return nil
}

func NewMsgUpdateGroupMember(
	operator sdk.AccAddress, groupName string, membersToAdd []sdk.AccAddress,
	membersToDelete []sdk.AccAddress) *MsgUpdateGroupMember {
	var membersAddrToAdd, membersAddrToDelete []string
	for _, member := range membersToAdd {
		membersAddrToAdd = append(membersAddrToAdd, member.String())
	}
	for _, member := range membersToDelete {
		membersAddrToDelete = append(membersAddrToDelete, member.String())
	}
	return &MsgUpdateGroupMember{
		Operator:        operator.String(),
		GroupName:       groupName,
		MembersToAdd:    membersAddrToAdd,
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
