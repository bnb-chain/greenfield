package types

import (
	"math/big"

	"cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/types/common"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	BucketChannel = "bucket"
	ObjectChannel = "object"
	GroupChannel  = "group"

	BucketChannelId sdk.ChannelID = 4
	ObjectChannelId sdk.ChannelID = 5
	GroupChannelId  sdk.ChannelID = 6

	// bucket operation types

	OperationMirrorBucket uint8 = 1
	OperationCreateBucket uint8 = 2
	OperationDeleteBucket uint8 = 3

	// object operation types

	OperationMirrorObject uint8 = 1
	// OperationCreateObject uint8 = 2 // not used
	OperationDeleteObject uint8 = 3

	// group operation types

	OperationMirrorGroup       uint8 = 1
	OperationCreateGroup       uint8 = 2
	OperationDeleteGroup       uint8 = 3
	OperationUpdateGroupMember uint8 = 4
)

type CrossChainPackage struct {
	OperationType uint8
	Package       []byte
}

func (p CrossChainPackage) MustSerialize() []byte {
	encodedBytes, err := rlp.EncodeToBytes(p)
	if err != nil {
		panic("encode delete cross chain package error")
	}
	return encodedBytes
}

func DeserializeRawCrossChainPackage(serializedPackage []byte) (*CrossChainPackage, error) {
	var tp CrossChainPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize raw cross chain package failed")
	}
	return &tp, nil
}

type DeserializeFunc func(serializedPackage []byte) (interface{}, error)

var DeserializeFuncMap = map[sdk.ChannelID]map[uint8][3]DeserializeFunc{
	BucketChannelId: {
		OperationMirrorBucket: {
			DeserializeMirrorBucketSynPackage,
			DeserializeMirrorBucketAckPackage,
			DeserializeMirrorBucketSynPackage,
		},
		OperationCreateBucket: {
			DeserializeCreateBucketSynPackage,
			DeserializeCreateBucketAckPackage,
			DeserializeCreateBucketSynPackage,
		},
		OperationDeleteBucket: {
			DeserializeDeleteBucketSynPackage,
			DeserializeDeleteBucketAckPackage,
			DeserializeDeleteBucketSynPackage,
		},
	},
	ObjectChannelId: {
		OperationMirrorObject: {
			DeserializeMirrorObjectSynPackage,
			DeserializeMirrorObjectAckPackage,
			DeserializeMirrorObjectSynPackage,
		},
		OperationDeleteObject: {
			DeserializeDeleteObjectSynPackage,
			DeserializeDeleteObjectAckPackage,
			DeserializeDeleteObjectSynPackage,
		},
	},
	GroupChannelId: {
		OperationMirrorGroup: {
			DeserializeMirrorGroupSynPackage,
			DeserializeMirrorGroupAckPackage,
			DeserializeMirrorGroupSynPackage,
		},
		OperationCreateGroup: {
			DeserializeCreateGroupSynPackage,
			DeserializeCreateGroupAckPackage,
			DeserializeCreateGroupSynPackage,
		},
		OperationDeleteGroup: {
			DeserializeDeleteGroupSynPackage,
			DeserializeDeleteGroupAckPackage,
			DeserializeDeleteGroupSynPackage,
		},
		OperationUpdateGroupMember: {
			DeserializeUpdateGroupMemberSynPackage,
			DeserializeUpdateGroupMemberAckPackage,
			DeserializeUpdateGroupMemberSynPackage,
		},
	},
}

func DeserializeCrossChainPackage(rawPack []byte, channelId sdk.ChannelID, packageType sdk.CrossChainPackageType) (interface{}, error) {
	if packageType >= 3 {
		return nil, ErrInvalidCrossChainPackage
	}

	pack, err := DeserializeRawCrossChainPackage(rawPack)
	if err != nil {
		return nil, err
	}

	operationMap, ok := DeserializeFuncMap[channelId][pack.OperationType]
	if !ok {
		return nil, ErrInvalidCrossChainPackage
	}

	return operationMap[packageType](pack.Package)
}

const (
	StatusSuccess = 0
	StatusFail    = 1
)

type MirrorBucketSynPackage struct {
	Id    *big.Int
	Owner sdk.AccAddress
}

type MirrorBucketAckPackage struct {
	Status uint8
	Id     *big.Int
}

func DeserializeMirrorBucketSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp MirrorBucketSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror bucket syn package failed")
	}
	return &tp, nil
}

func DeserializeMirrorBucketAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp MirrorBucketAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror bucket ack package failed")
	}
	return &tp, nil
}

type MirrorObjectSynPackage struct {
	Id    *big.Int
	Owner sdk.AccAddress
}

type MirrorObjectAckPackage struct {
	Status uint8
	Id     *big.Int
}

func DeserializeMirrorObjectSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp MirrorObjectSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror object syn package failed")
	}
	return &tp, nil
}

func DeserializeMirrorObjectAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp MirrorObjectAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror object ack package failed")
	}
	return &tp, nil
}

type MirrorGroupSynPackage struct {
	Id    *big.Int
	Owner sdk.AccAddress
}

type MirrorGroupAckPackage struct {
	Status uint8
	Id     *big.Int
}

func DeserializeMirrorGroupSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp MirrorGroupSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror group syn package failed")
	}
	return &tp, nil
}

func DeserializeMirrorGroupAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp MirrorGroupAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror group ack package failed")
	}
	return &tp, nil
}

type CreateBucketSynPackage struct {
	Creator                        sdk.AccAddress
	BucketName                     string
	Visibility                     uint32
	PaymentAddress                 sdk.AccAddress
	PrimarySpAddress               sdk.AccAddress
	PrimarySpApprovalExpiredHeight uint64
	PrimarySpApprovalSignature     []byte
	ChargedReadQuota               uint64
	ExtraData                      []byte
}

func (p CreateBucketSynPackage) ValidateBasic() error {
	msg := MsgCreateBucket{
		Creator:          p.Creator.String(),
		BucketName:       p.BucketName,
		Visibility:       VisibilityType(p.Visibility),
		PaymentAddress:   p.PaymentAddress.String(),
		PrimarySpAddress: p.PrimarySpAddress.String(),
		PrimarySpApproval: &common.Approval{
			ExpiredHeight: p.PrimarySpApprovalExpiredHeight,
			Sig:           p.PrimarySpApprovalSignature,
		},
		ChargedReadQuota: p.ChargedReadQuota,
	}

	return msg.ValidateBasic()
}

func (p CreateBucketSynPackage) GetApprovalBytes() []byte {
	msg := MsgCreateBucket{
		Creator:          p.Creator.String(),
		BucketName:       p.BucketName,
		Visibility:       VisibilityType(p.Visibility),
		PaymentAddress:   p.PaymentAddress.String(),
		PrimarySpAddress: p.PrimarySpAddress.String(),
		PrimarySpApproval: &common.Approval{
			ExpiredHeight: p.PrimarySpApprovalExpiredHeight,
			Sig:           p.PrimarySpApprovalSignature,
		},
		ChargedReadQuota: p.ChargedReadQuota,
	}
	return msg.GetApprovalBytes()
}

func DeserializeCreateBucketSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp CreateBucketSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create bucket syn package failed")
	}
	return &tp, nil
}

type CreateBucketAckPackage struct {
	Status    uint8
	Id        *big.Int
	Creator   sdk.AccAddress
	ExtraData []byte
}

func (p CreateBucketAckPackage) MustSerialize() []byte {
	encodedBytes, err := rlp.EncodeToBytes(p)
	if err != nil {
		panic("encode create bucket ack package error")
	}
	return encodedBytes
}

func DeserializeCreateBucketAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp CreateBucketAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create bucket ack package failed")
	}
	return &tp, nil
}

type DeleteBucketSynPackage struct {
	Operator  sdk.AccAddress
	Id        *big.Int
	ExtraData []byte
}

func (p DeleteBucketSynPackage) ValidateBasic() error {
	if p.Operator.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	if p.Id == nil || p.Id.Cmp(big.NewInt(0)) < 0 {
		return ErrInvalidId
	}
	return nil
}

func DeserializeDeleteBucketSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp DeleteBucketSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete bucket syn package failed")
	}
	return &tp, nil
}

type DeleteBucketAckPackage struct {
	Status    uint8
	Id        *big.Int
	ExtraData []byte
}

func (p DeleteBucketAckPackage) MustSerialize() []byte {
	encodedBytes, err := rlp.EncodeToBytes(p)
	if err != nil {
		panic("encode delete bucket ack package error")
	}
	return encodedBytes
}

func DeserializeDeleteBucketAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp DeleteBucketAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete bucket ack package failed")
	}
	return &tp, nil
}

type CreateGroupSynPackage struct {
	Creator   sdk.AccAddress
	GroupName string
	ExtraData []byte
}

func (p CreateGroupSynPackage) ValidateBasic() error {
	msg := MsgCreateGroup{
		Creator:   p.Creator.String(),
		GroupName: p.GroupName,
	}
	return msg.ValidateBasic()
}

func DeserializeCreateGroupSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp CreateGroupSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create group syn package failed")
	}
	return &tp, nil
}

type CreateGroupAckPackage struct {
	Status    uint8
	Id        *big.Int
	Creator   sdk.AccAddress
	ExtraData []byte
}

func (p CreateGroupAckPackage) MustSerialize() []byte {
	encodedBytes, err := rlp.EncodeToBytes(p)
	if err != nil {
		panic("encode create group ack package error")
	}
	return encodedBytes
}

func DeserializeCreateGroupAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp CreateGroupAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create group ack package failed")
	}
	return &tp, nil
}

type DeleteObjectSynPackage struct {
	Operator  sdk.AccAddress
	Id        *big.Int
	ExtraData []byte
}

func (p DeleteObjectSynPackage) ValidateBasic() error {
	if p.Operator.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	if p.Id == nil || p.Id.Cmp(big.NewInt(0)) < 0 {
		return ErrInvalidId
	}
	return nil
}

func DeserializeDeleteObjectSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp DeleteObjectSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete object syn package failed")
	}
	return &tp, nil
}

type DeleteObjectAckPackage struct {
	Status    uint8
	Id        *big.Int
	ExtraData []byte
}

func (p DeleteObjectAckPackage) MustSerialize() []byte {
	encodedBytes, err := rlp.EncodeToBytes(p)
	if err != nil {
		panic("encode delete object ack package error")
	}
	return encodedBytes
}

func DeserializeDeleteObjectAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp DeleteObjectAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete object syn package failed")
	}
	return &tp, nil
}

type DeleteGroupSynPackage struct {
	Operator  sdk.AccAddress
	Id        *big.Int
	ExtraData []byte
}

func (p DeleteGroupSynPackage) ValidateBasic() error {
	if p.Operator.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	if p.Id == nil || p.Id.Cmp(big.NewInt(0)) < 0 {
		return ErrInvalidId
	}
	return nil
}

func DeserializeDeleteGroupSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp DeleteGroupSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete group syn package failed")
	}
	return &tp, nil
}

type DeleteGroupAckPackage struct {
	Status    uint8
	Id        *big.Int
	ExtraData []byte
}

func DeserializeDeleteGroupAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp DeleteGroupAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete group ack package failed")
	}
	return &tp, nil
}

func (p DeleteGroupAckPackage) MustSerialize() []byte {
	encodedBytes, err := rlp.EncodeToBytes(p)
	if err != nil {
		panic("encode delete group ack package error")
	}
	return encodedBytes
}

const (
	OperationAddGroupMember    uint8 = 0
	OperationDeleteGroupMember uint8 = 1
)

type UpdateGroupMemberSynPackage struct {
	Operator      sdk.AccAddress
	GroupId       *big.Int
	OperationType uint8
	Members       []sdk.AccAddress
	ExtraData     []byte
}

func (p UpdateGroupMemberSynPackage) GetMembers() []string {
	members := make([]string, 0, len(p.Members))
	for _, member := range p.Members {
		members = append(members, member.String())
	}
	return members
}

func (p UpdateGroupMemberSynPackage) ValidateBasic() error {
	if p.OperationType != OperationAddGroupMember && p.OperationType != OperationDeleteGroupMember {
		return ErrInvalidOperationType
	}

	if p.Operator.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	if p.GroupId == nil || p.GroupId.Cmp(big.NewInt(0)) < 0 {
		return ErrInvalidId
	}

	for _, member := range p.Members {
		if member.Empty() {
			return sdkerrors.ErrInvalidAddress
		}
	}
	return nil
}

func DeserializeUpdateGroupMemberSynPackage(serializedPackage []byte) (interface{}, error) {
	var tp UpdateGroupMemberSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize update group member syn package failed")
	}
	return &tp, nil
}

type UpdateGroupMemberAckPackage struct {
	Status        uint8
	Id            *big.Int
	Operator      sdk.AccAddress
	OperationType uint8
	Members       []sdk.AccAddress
	ExtraData     []byte
}

func DeserializeUpdateGroupMemberAckPackage(serializedPackage []byte) (interface{}, error) {
	var tp UpdateGroupMemberAckPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize update group member ack package failed")
	}
	return &tp, nil
}

func (p UpdateGroupMemberAckPackage) MustSerialize() []byte {
	encodedBytes, err := rlp.EncodeToBytes(p)
	if err != nil {
		panic("encode delete group ack package error")
	}
	return encodedBytes
}
