package types

import (
	"math/big"
	time "time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	gnfdcommon "github.com/bnb-chain/greenfield/types/common"
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

func SafeBigInt(input *big.Int) *big.Int {
	if input == nil {
		return big.NewInt(0)
	}
	return input
}

type CrossChainPackage struct {
	OperationType uint8
	Package       []byte
}

func (p CrossChainPackage) MustSerialize() []byte {
	return append([]byte{p.OperationType}, p.Package...)
}

func DeserializeRawCrossChainPackage(serializedPackage []byte) (*CrossChainPackage, error) {
	tp := CrossChainPackage{
		OperationType: serializedPackage[0],
		Package:       serializedPackage[1:],
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

type GeneralMirrorSynPackageStruct struct {
	Id    *big.Int
	Owner common.Address
}

type MirrorBucketAckPackage struct {
	Status uint8
	Id     *big.Int
}

var (
	generalMirrorSynPackageStructType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Id", Type: "uint256"},
		{Name: "Owner", Type: "address"},
	})

	generalMirrorSynPackageArgs = abi.Arguments{
		{Type: generalMirrorSynPackageStructType},
	}

	generalMirrorAckPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Status", Type: "uint8"},
		{Name: "Id", Type: "uint256"},
	})

	generalMirrorAckPackageArgs = abi.Arguments{
		{Type: generalMirrorAckPackageType},
	}
)

func (pkg *MirrorBucketSynPackage) Serialize() ([]byte, error) {
	return generalMirrorSynPackageArgs.Pack(&GeneralMirrorSynPackageStruct{
		SafeBigInt(pkg.Id),
		common.BytesToAddress(pkg.Owner),
	})
}

func deserializeMirrorSynPackage(serializedPackage []byte) (*GeneralMirrorSynPackageStruct, error) {
	unpacked, err := generalMirrorSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror syn package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], GeneralMirrorSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(GeneralMirrorSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect mirror syn package failed")
	}
	return &pkgStruct, nil
}

func DeserializeMirrorBucketSynPackage(serializedPackage []byte) (interface{}, error) {
	pkgStruct, err := deserializeMirrorSynPackage(serializedPackage)
	if err != nil {
		return nil, err
	}

	tp := MirrorBucketSynPackage{
		pkgStruct.Id,
		pkgStruct.Owner.Bytes(),
	}
	return &tp, nil
}

func (pkg *MirrorBucketAckPackage) Serialize() ([]byte, error) {
	return generalMirrorAckPackageArgs.Pack(&MirrorBucketAckPackage{
		pkg.Status,
		SafeBigInt(pkg.Id),
	})
}

func DeserializeMirrorBucketAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalMirrorAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror bucket ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], MirrorBucketAckPackage{})
	tp, ok := unpackedStruct.(MirrorBucketAckPackage)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect mirror bucket ack package failed")
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

func (pkg *MirrorObjectSynPackage) Serialize() ([]byte, error) {
	return generalMirrorSynPackageArgs.Pack(&GeneralMirrorSynPackageStruct{
		SafeBigInt(pkg.Id),
		common.BytesToAddress(pkg.Owner),
	})
}

func DeserializeMirrorObjectSynPackage(serializedPackage []byte) (interface{}, error) {
	pkgStruct, err := deserializeMirrorSynPackage(serializedPackage)
	if err != nil {
		return nil, err
	}

	tp := MirrorObjectSynPackage{
		pkgStruct.Id,
		pkgStruct.Owner.Bytes(),
	}
	return &tp, nil
}

func (pkg *MirrorObjectAckPackage) Serialize() ([]byte, error) {
	return generalMirrorAckPackageArgs.Pack(&MirrorObjectAckPackage{
		pkg.Status,
		SafeBigInt(pkg.Id),
	})
}

func DeserializeMirrorObjectAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalMirrorAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror object ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], MirrorObjectAckPackage{})
	tp, ok := unpackedStruct.(MirrorObjectAckPackage)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect mirror object ack package failed")
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

func (pkg *MirrorGroupSynPackage) Serialize() ([]byte, error) {
	return generalMirrorSynPackageArgs.Pack(&GeneralMirrorSynPackageStruct{
		SafeBigInt(pkg.Id),
		common.BytesToAddress(pkg.Owner),
	})
}

func DeserializeMirrorGroupSynPackage(serializedPackage []byte) (interface{}, error) {
	pkgStruct, err := deserializeMirrorSynPackage(serializedPackage)
	if err != nil {
		return nil, err
	}

	tp := MirrorGroupSynPackage{
		pkgStruct.Id,
		pkgStruct.Owner.Bytes(),
	}
	return &tp, nil
}

func (pkg *MirrorGroupAckPackage) Serialize() ([]byte, error) {
	return generalMirrorAckPackageArgs.Pack(&MirrorGroupAckPackage{
		pkg.Status,
		SafeBigInt(pkg.Id),
	})
}

func DeserializeMirrorGroupAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalMirrorAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize mirror group ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], MirrorGroupAckPackage{})
	tp, ok := unpackedStruct.(MirrorGroupAckPackage)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect mirror group ack package failed")
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

type CreateBucketSynPackageStruct struct {
	Creator                        common.Address
	BucketName                     string
	Visibility                     uint32
	PaymentAddress                 common.Address
	PrimarySpAddress               common.Address
	PrimarySpApprovalExpiredHeight uint64
	PrimarySpApprovalSignature     []byte
	ChargedReadQuota               uint64
	ExtraData                      []byte
}

var (
	createBucketSynPackageStructType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Creator", Type: "address"},
		{Name: "BucketName", Type: "string"},
		{Name: "Visibility", Type: "uint32"},
		{Name: "PaymentAddress", Type: "address"},
		{Name: "PrimarySpAddress", Type: "address"},
		{Name: "PrimarySpApprovalExpiredHeight", Type: "uint64"},
		{Name: "PrimarySpApprovalSignature", Type: "bytes"},
		{Name: "ChargedReadQuota", Type: "uint64"},
		{Name: "ExtraData", Type: "bytes"},
	})

	createBucketSynPackageStructArgs = abi.Arguments{
		{Type: createBucketSynPackageStructType},
	}
)

func (p CreateBucketSynPackage) MustSerialize() []byte {
	encodedBytes, err := createBucketSynPackageStructArgs.Pack(&CreateBucketSynPackageStruct{
		Creator:                        common.BytesToAddress(p.Creator),
		BucketName:                     p.BucketName,
		Visibility:                     p.Visibility,
		PaymentAddress:                 common.BytesToAddress(p.PaymentAddress),
		PrimarySpAddress:               common.BytesToAddress(p.PrimarySpAddress),
		PrimarySpApprovalExpiredHeight: p.PrimarySpApprovalExpiredHeight,
		PrimarySpApprovalSignature:     p.PrimarySpApprovalSignature,
		ChargedReadQuota:               p.ChargedReadQuota,
		ExtraData:                      p.ExtraData,
	})
	if err != nil {
		panic("encode create bucket syn package error")
	}
	return encodedBytes
}

func (p CreateBucketSynPackage) ValidateBasic() error {
	msg := MsgCreateBucket{
		Creator:          p.Creator.String(),
		BucketName:       p.BucketName,
		Visibility:       VisibilityType(p.Visibility),
		PaymentAddress:   p.PaymentAddress.String(),
		PrimarySpAddress: p.PrimarySpAddress.String(),
		PrimarySpApproval: &gnfdcommon.Approval{
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
		PrimarySpApproval: &gnfdcommon.Approval{
			ExpiredHeight: p.PrimarySpApprovalExpiredHeight,
			Sig:           p.PrimarySpApprovalSignature,
		},
		ChargedReadQuota: p.ChargedReadQuota,
	}
	return msg.GetApprovalBytes()
}

func DeserializeCreateBucketSynPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := createBucketSynPackageStructArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create bucket syn package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], CreateBucketSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(CreateBucketSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect create bucket syn package failed")
	}

	tp := CreateBucketSynPackage{
		pkgStruct.Creator.Bytes(),
		pkgStruct.BucketName,
		pkgStruct.Visibility,
		pkgStruct.PaymentAddress.Bytes(),
		pkgStruct.PrimarySpAddress.Bytes(),
		pkgStruct.PrimarySpApprovalExpiredHeight,
		pkgStruct.PrimarySpApprovalSignature,
		pkgStruct.ChargedReadQuota,
		pkgStruct.ExtraData,
	}
	return &tp, nil
}

type CreateBucketAckPackage struct {
	Status    uint8
	Id        *big.Int
	Creator   sdk.AccAddress
	ExtraData []byte
}

type GeneralCreateAckPackageStruct struct {
	Status    uint8
	Id        *big.Int
	Creator   common.Address
	ExtraData []byte
}

var (
	generalCreateAckPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Status", Type: "uint8"},
		{Name: "Id", Type: "uint256"},
		{Name: "Creator", Type: "address"},
		{Name: "ExtraData", Type: "bytes"},
	})

	generalCreateAckPackageArgs = abi.Arguments{
		{Type: generalCreateAckPackageType},
	}
)

func (p CreateBucketAckPackage) MustSerialize() []byte {
	encodedBytes, err := generalCreateAckPackageArgs.Pack(&GeneralCreateAckPackageStruct{
		Status:    p.Status,
		Id:        SafeBigInt(p.Id),
		Creator:   common.BytesToAddress(p.Creator),
		ExtraData: p.ExtraData,
	})
	if err != nil {
		panic("encode create bucket ack package error")
	}
	return encodedBytes
}

func DeserializeCreateBucketAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalCreateAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create bucket ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], GeneralCreateAckPackageStruct{})
	pkgStruct, ok := unpackedStruct.(GeneralCreateAckPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect create bucket ack package failed")
	}

	tp := CreateBucketAckPackage{
		Status:    pkgStruct.Status,
		Id:        pkgStruct.Id,
		Creator:   pkgStruct.Creator.Bytes(),
		ExtraData: pkgStruct.ExtraData,
	}
	return &tp, nil
}

type DeleteBucketSynPackage struct {
	Operator  sdk.AccAddress
	Id        *big.Int
	ExtraData []byte
}

type GeneralDeleteSynPackageStruct struct {
	Operator  common.Address
	Id        *big.Int
	ExtraData []byte
}

var (
	generalDeleteSynPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Operator", Type: "address"},
		{Name: "Id", Type: "uint256"},
		{Name: "ExtraData", Type: "bytes"},
	})

	generalDeleteSynPackageArgs = abi.Arguments{
		{Type: generalDeleteSynPackageType},
	}
)

func (p DeleteBucketSynPackage) MustSerialize() []byte {
	encodedBytes, err := generalDeleteSynPackageArgs.Pack(&GeneralDeleteSynPackageStruct{
		Operator:  common.BytesToAddress(p.Operator),
		Id:        SafeBigInt(p.Id),
		ExtraData: p.ExtraData,
	})
	if err != nil {
		panic("encode delete bucket sync package error")
	}
	return encodedBytes
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
	unpacked, err := generalDeleteSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete bucket syn package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], GeneralDeleteSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(GeneralDeleteSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect delete bucket syn package failed")
	}

	tp := DeleteBucketSynPackage{
		Operator:  pkgStruct.Operator.Bytes(),
		Id:        pkgStruct.Id,
		ExtraData: pkgStruct.ExtraData,
	}
	return &tp, nil
}

type DeleteBucketAckPackage struct {
	Status    uint8
	Id        *big.Int
	ExtraData []byte
}

var (
	generalDeleteAckPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Status", Type: "uint8"},
		{Name: "Id", Type: "uint256"},
		{Name: "ExtraData", Type: "bytes"},
	})

	generalDeleteAckPackageArgs = abi.Arguments{
		{Type: generalDeleteAckPackageType},
	}
)

func (p DeleteBucketAckPackage) MustSerialize() []byte {
	encodedBytes, err := generalDeleteAckPackageArgs.Pack(&DeleteBucketAckPackage{
		p.Status,
		SafeBigInt(p.Id),
		p.ExtraData,
	})
	if err != nil {
		panic("encode delete bucket ack package error")
	}
	return encodedBytes
}

func DeserializeDeleteBucketAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalDeleteAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete bucket ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], DeleteBucketAckPackage{})
	tp, ok := unpackedStruct.(DeleteBucketAckPackage)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect delete bucket ack package failed")
	}
	return &tp, nil
}

type CreateGroupSynPackage struct {
	Creator   sdk.AccAddress
	GroupName string
	ExtraData []byte
}

type CreateGroupSynPackageStruct struct {
	Creator   common.Address
	GroupName string
	ExtraData []byte
}

var (
	createGroupSynPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Creator", Type: "address"},
		{Name: "GroupName", Type: "string"},
		{Name: "ExtraData", Type: "bytes"},
	})

	createGroupSynPackageArgs = abi.Arguments{
		{Type: createGroupSynPackageType},
	}
)

func (p CreateGroupSynPackage) ValidateBasic() error {
	msg := MsgCreateGroup{
		Creator:   p.Creator.String(),
		GroupName: p.GroupName,
	}
	return msg.ValidateBasic()
}

func (p CreateGroupSynPackage) MustSerialize() []byte {
	encodedBytes, err := createGroupSynPackageArgs.Pack(&CreateGroupSynPackageStruct{
		Creator:   common.BytesToAddress(p.Creator),
		GroupName: p.GroupName,
		ExtraData: p.ExtraData,
	})
	if err != nil {
		panic("encode create group syn package error")
	}
	return encodedBytes
}

func DeserializeCreateGroupSynPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := createGroupSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create group syn package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], CreateGroupSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(CreateGroupSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect create group syn package failed")
	}

	tp := CreateGroupSynPackage{
		pkgStruct.Creator.Bytes(),
		pkgStruct.GroupName,
		pkgStruct.ExtraData,
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
	encodedBytes, err := generalCreateAckPackageArgs.Pack(&GeneralCreateAckPackageStruct{
		Status:    p.Status,
		Id:        SafeBigInt(p.Id),
		Creator:   common.BytesToAddress(p.Creator),
		ExtraData: p.ExtraData,
	})
	if err != nil {
		panic("encode create group ack package error")
	}
	return encodedBytes
}

func DeserializeCreateGroupAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalCreateAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize create group ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], GeneralCreateAckPackageStruct{})
	pkgStruct, ok := unpackedStruct.(GeneralCreateAckPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect create group ack package failed")
	}

	tp := CreateGroupAckPackage{
		Status:    pkgStruct.Status,
		Id:        pkgStruct.Id,
		Creator:   pkgStruct.Creator.Bytes(),
		ExtraData: pkgStruct.ExtraData,
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
	unpacked, err := generalDeleteSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete object syn package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], GeneralDeleteSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(GeneralDeleteSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect delete object syn package failed")
	}

	tp := DeleteObjectSynPackage{
		Operator:  pkgStruct.Operator.Bytes(),
		Id:        pkgStruct.Id,
		ExtraData: pkgStruct.ExtraData,
	}
	return &tp, nil
}

type DeleteObjectAckPackage struct {
	Status    uint8
	Id        *big.Int
	ExtraData []byte
}

func (p DeleteObjectAckPackage) MustSerialize() []byte {
	encodedBytes, err := generalDeleteAckPackageArgs.Pack(&DeleteObjectAckPackage{
		p.Status,
		SafeBigInt(p.Id),
		p.ExtraData,
	})
	if err != nil {
		panic("encode delete object ack package error")
	}
	return encodedBytes
}

func DeserializeDeleteObjectAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalDeleteAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete object ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], DeleteObjectAckPackage{})
	tp, ok := unpackedStruct.(DeleteObjectAckPackage)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect delete object ack package failed")
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
	unpacked, err := generalDeleteSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete group syn package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], GeneralDeleteSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(GeneralDeleteSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect delete group syn package failed")
	}

	tp := DeleteGroupSynPackage{
		Operator:  pkgStruct.Operator.Bytes(),
		Id:        pkgStruct.Id,
		ExtraData: pkgStruct.ExtraData,
	}
	return &tp, nil
}

type DeleteGroupAckPackage struct {
	Status    uint8
	Id        *big.Int
	ExtraData []byte
}

func DeserializeDeleteGroupAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := generalDeleteAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete group ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], DeleteGroupAckPackage{})
	tp, ok := unpackedStruct.(DeleteGroupAckPackage)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect delete group ack package failed")
	}
	return &tp, nil
}

func (p DeleteGroupAckPackage) MustSerialize() []byte {
	encodedBytes, err := generalDeleteAckPackageArgs.Pack(&DeleteGroupAckPackage{
		p.Status,
		SafeBigInt(p.Id),
		p.ExtraData,
	})
	if err != nil {
		panic("encode delete group ack package error")
	}
	return encodedBytes
}

const (
	OperationAddGroupMember    uint8 = 0
	OperationDeleteGroupMember uint8 = 1
	OperationRenewGroupMember  uint8 = 2
)

type UpdateGroupMemberSynPackage struct {
	Operator      sdk.AccAddress
	GroupId       *big.Int
	OperationType uint8
	Members       []sdk.AccAddress
	ExtraData     []byte
}

type UpdateGroupMemberSynPackageStruct struct {
	Operator      common.Address
	GroupId       *big.Int
	OperationType uint8
	Members       []common.Address
	ExtraData     []byte
}

var (
	updateGroupMemberSynPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Operator", Type: "address"},
		{Name: "GroupId", Type: "uint256"},
		{Name: "OperationType", Type: "uint8"},
		{Name: "Members", Type: "address[]"},
		{Name: "ExtraData", Type: "bytes"},
	})

	updateGroupMemberSynPackageArgs = abi.Arguments{
		{Type: updateGroupMemberSynPackageType},
	}
)

func (p UpdateGroupMemberSynPackage) GetMembers() []string {
	members := make([]string, 0, len(p.Members))
	for _, member := range p.Members {
		members = append(members, member.String())
	}
	return members
}

func (p UpdateGroupMemberSynPackage) MustSerialize() []byte {
	totalMember := len(p.Members)
	members := make([]common.Address, totalMember)
	for i, member := range p.Members {
		members[i] = common.BytesToAddress(member)
	}

	encodedBytes, err := updateGroupMemberSynPackageArgs.Pack(&UpdateGroupMemberSynPackageStruct{
		common.BytesToAddress(p.Operator),
		SafeBigInt(p.GroupId),
		p.OperationType,
		members,
		p.ExtraData,
	})
	if err != nil {
		panic("encode update group member syn package error")
	}
	return encodedBytes
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
	unpacked, err := updateGroupMemberSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize delete bucket ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], UpdateGroupMemberSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(UpdateGroupMemberSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect delete bucket ack package failed")
	}

	totalMember := len(pkgStruct.Members)
	members := make([]sdk.AccAddress, totalMember)
	for i, member := range pkgStruct.Members {
		members[i] = member.Bytes()
	}
	tp := UpdateGroupMemberSynPackage{
		pkgStruct.Operator.Bytes(),
		pkgStruct.GroupId,
		pkgStruct.OperationType,
		members,
		pkgStruct.ExtraData,
	}
	return &tp, nil
}

type UpdateGroupMemberV2SynPackage struct {
	Operator         sdk.AccAddress
	GroupId          *big.Int
	OperationType    uint8
	Members          []sdk.AccAddress
	ExtraData        []byte
	MemberExpiration []uint64
}

func RegisterUpdateGroupMemberV2SynPackageType() {
	DeserializeFuncMap[GroupChannelId] = map[uint8][3]DeserializeFunc{
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
			DeserializeUpdateGroupMemberV2SynPackage,
			DeserializeUpdateGroupMemberAckPackage,
			DeserializeUpdateGroupMemberV2SynPackage,
		},
	}
}

type UpdateGroupMemberV2SynPackageStruct struct {
	Operator         common.Address
	GroupId          *big.Int
	OperationType    uint8
	Members          []common.Address
	ExtraData        []byte
	MemberExpiration []uint64
}

var (
	updateGroupMemberV2SynPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Operator", Type: "address"},
		{Name: "GroupId", Type: "uint256"},
		{Name: "OperationType", Type: "uint8"},
		{Name: "Members", Type: "address[]"},
		{Name: "ExtraData", Type: "bytes"},
		{Name: "MemberExpiration", Type: "uint64[]"},
	})

	updateGroupMemberV2SynPackageArgs = abi.Arguments{
		{Type: updateGroupMemberV2SynPackageType},
	}
)

func (p UpdateGroupMemberV2SynPackage) GetMembers() []string {
	members := make([]string, 0, len(p.Members))
	for _, member := range p.Members {
		members = append(members, member.String())
	}
	return members
}

func (p UpdateGroupMemberV2SynPackage) GetMemberExpiration() []time.Time {
	memberExpiration := make([]time.Time, 0, len(p.MemberExpiration))
	for _, expiration := range p.MemberExpiration {
		memberExpiration = append(memberExpiration, time.Unix(int64(expiration), 0))
	}
	return memberExpiration
}

func (p UpdateGroupMemberV2SynPackage) MustSerialize() []byte {
	totalMember := len(p.Members)
	members := make([]common.Address, totalMember)
	for i, member := range p.Members {
		members[i] = common.BytesToAddress(member)
	}

	encodedBytes, err := updateGroupMemberSynPackageArgs.Pack(&UpdateGroupMemberSynPackageStruct{
		common.BytesToAddress(p.Operator),
		SafeBigInt(p.GroupId),
		p.OperationType,
		members,
		p.ExtraData,
	})
	if err != nil {
		panic("encode update group member syn package error")
	}
	return encodedBytes
}

func (p UpdateGroupMemberV2SynPackage) ValidateBasic() error {
	if p.OperationType != OperationAddGroupMember && p.OperationType != OperationDeleteGroupMember && p.OperationType != OperationRenewGroupMember {
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

	if len(p.Members) != len(p.MemberExpiration) {
		return ErrInvalidGroupMemberExpiration
	}

	return nil
}

func DeserializeUpdateGroupMemberV2SynPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := updateGroupMemberV2SynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize update group member sun package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], UpdateGroupMemberV2SynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(UpdateGroupMemberV2SynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect update group member sun package failed")
	}

	totalMember := len(pkgStruct.Members)
	members := make([]sdk.AccAddress, totalMember)
	for i, member := range pkgStruct.Members {
		members[i] = member.Bytes()
	}
	tp := UpdateGroupMemberV2SynPackage{
		pkgStruct.Operator.Bytes(),
		pkgStruct.GroupId,
		pkgStruct.OperationType,
		members,
		pkgStruct.ExtraData,
		pkgStruct.MemberExpiration,
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

type UpdateGroupMemberAckPackageStruct struct {
	Status        uint8
	Id            *big.Int
	Operator      common.Address
	OperationType uint8
	Members       []common.Address
	ExtraData     []byte
}

var (
	updateGroupMemberAckPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Status", Type: "uint8"},
		{Name: "Id", Type: "uint256"},
		{Name: "Operator", Type: "address"},
		{Name: "OperationType", Type: "uint8"},
		{Name: "Members", Type: "address[]"},
		{Name: "ExtraData", Type: "bytes"},
	})

	updateGroupMemberAckPackageArgs = abi.Arguments{
		{Type: updateGroupMemberAckPackageType},
	}
)

func DeserializeUpdateGroupMemberAckPackage(serializedPackage []byte) (interface{}, error) {
	unpacked, err := updateGroupMemberAckPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "deserialize update group member ack package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], UpdateGroupMemberAckPackageStruct{})
	pkgStruct, ok := unpackedStruct.(UpdateGroupMemberAckPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidCrossChainPackage, "reflect update group member ack package failed")
	}

	totalMember := len(pkgStruct.Members)
	members := make([]sdk.AccAddress, totalMember)
	for i, member := range pkgStruct.Members {
		members[i] = member.Bytes()
	}
	tp := UpdateGroupMemberAckPackage{
		pkgStruct.Status,
		pkgStruct.Id,
		pkgStruct.Operator.Bytes(),
		pkgStruct.OperationType,
		members,
		pkgStruct.ExtraData,
	}
	return &tp, nil
}

func (p UpdateGroupMemberAckPackage) MustSerialize() []byte {
	totalMember := len(p.Members)
	members := make([]common.Address, totalMember)
	for i, member := range p.Members {
		members[i] = common.BytesToAddress(member)
	}

	encodedBytes, err := updateGroupMemberAckPackageArgs.Pack(&UpdateGroupMemberAckPackageStruct{
		p.Status,
		SafeBigInt(p.Id),
		common.BytesToAddress(p.Operator),
		p.OperationType,
		members,
		p.ExtraData,
	})
	if err != nil {
		panic("encode update group member ack package error")
	}
	return encodedBytes
}
