package types

import (
	"regexp"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "storage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_storage"
)

type RawID math.Uint

var (
	BucketPrefix      = []byte{0x11}
	ObjectPrefix      = []byte{0x12}
	GroupPrefix       = []byte{0x13}
	GroupMemberPrefix = []byte{0x14} // TODO(fynn): will be deprecated after permission module ready

	BucketByIDPrefix = []byte{0x21}
	ObjectByIDPrefix = []byte{0x22}
	GroupByIDPrefix  = []byte{0x23}

	BucketSequencePrefix = []byte{0x31}
	ObjectSequencePrefix = []byte{0x32}
	GroupSequencePrefix  = []byte{0x33}

	validBucketName = regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)
	ipAddress       = regexp.MustCompile(`^(\d+\.){3}\d+$`)
)

// GetBucketStoreKey return the bucket store key
func GetBucketStoreKey(bucketName string) []byte {
	objectHashKey := sdk.Keccak256([]byte(bucketName))
	return append(BucketPrefix, objectHashKey...)
}

func GetObjectStoreKey(bucketName string, objectName string) []byte {
	bucketKey := sdk.Keccak256([]byte(bucketName))
	objectKey := sdk.Keccak256([]byte(objectName))
	return append(ObjectPrefix, append(bucketKey, objectKey...)...)
}

func GetGroupStoreKey(owner sdk.AccAddress, groupName string) []byte {
	groupKey := sdk.Keccak256([]byte(groupName))
	return append(GroupPrefix, append(owner.Bytes(), groupKey...)...)
}

func GetGroupMemberKey(groupId math.Uint, memberAcc sdk.AccAddress) []byte {
	return append(GroupMemberPrefix, append(groupId.Bytes(), memberAcc.Bytes()...)...)
}

func GetBucketByIDStoreKey(bucketId math.Uint) []byte {
	return append(BucketByIDPrefix, EncodeSequence(bucketId)...)
}

func GetObjectByIDStoreKey(objectId math.Uint) []byte {
	return append(ObjectByIDPrefix, EncodeSequence(objectId)...)
}

func GetGroupByIDStoreKey(groupId math.Uint) []byte {
	return append(GroupByIDPrefix, EncodeSequence(groupId)...)
}
