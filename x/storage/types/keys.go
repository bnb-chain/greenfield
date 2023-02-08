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

var (
	BucketPrefix      = []byte{0x10}
	ObjectPrefix      = []byte{0x11}
	GroupPrefix       = []byte{0x12}
	GroupMemberPrefix = []byte{0x13} // TODO(fynn): will be deprecated after permission module ready

	validBucketName = regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)
	ipAddress       = regexp.MustCompile(`^(\d+\.){3}\d+$`)
)

func GetBucketKey(bucketName string) []byte {
	return sdk.Keccak256([]byte(bucketName))
}

func GetObjectKey(bucketName string, objectName string) []byte {
	sdk.Keccak256()
	bucketKey := sdk.Keccak256([]byte(bucketName))
	objectKey := sdk.Keccak256([]byte(objectName))
	return append(bucketKey, objectKey...)
}

func GetGroupKey(owner sdk.AccAddress, groupName string) []byte {
	groupKey := sdk.Keccak256([]byte(groupName))
	return append([]byte(owner.String()), groupKey...)
}

func GetGroupMemberKey(groupId math.Uint, memberAcc sdk.AccAddress) []byte {
	return append(MustMarshalUint(groupId), []byte(memberAcc.String())...)
}
