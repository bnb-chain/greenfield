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

// GetBucketKey return the bucket name store key
func GetBucketKey(bucketName string) []byte {
	objectNameHash := sdk.Keccak256([]byte(bucketName))
	return append(BucketPrefix, objectNameHash...)
}

// GetObjectKey return the object name store key
func GetObjectKey(bucketName string, objectName string) []byte {
	bucketNameHash := sdk.Keccak256([]byte(bucketName))
	objectNameHash := sdk.Keccak256([]byte(objectName))
	return append(ObjectPrefix, append(bucketNameHash, objectNameHash...)...)
}

func GetObjectKeyOnlyBucketPrefix(bucketName string) []byte {
	return append(ObjectPrefix, sdk.Keccak256([]byte(bucketName))...)
}

// GetGroupKey return the group name store key
func GetGroupKey(owner sdk.AccAddress, groupName string) []byte {
	groupNameHash := sdk.Keccak256([]byte(groupName))
	return append(GroupPrefix, append(owner.Bytes(), groupNameHash...)...)
}

// GetGroupMemberKey return the group member name store key
func GetGroupMemberKey(groupId math.Uint, memberAcc sdk.AccAddress) []byte {
	return append(GroupMemberPrefix, append(groupId.Bytes(), memberAcc.Bytes()...)...)
}

// GetBucketByIDKey return the bucketID store key
func GetBucketByIDKey(bucketId math.Uint) []byte {
	return append(BucketByIDPrefix, EncodeSequence(bucketId)...)
}

// GetObjectByIDKey return the objectId store key
func GetObjectByIDKey(objectId math.Uint) []byte {
	return append(ObjectByIDPrefix, EncodeSequence(objectId)...)
}

// GetGroupByIDKey return the groupId store key
func GetGroupByIDKey(groupId math.Uint) []byte {
	return append(GroupByIDPrefix, EncodeSequence(groupId)...)
}
