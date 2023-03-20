package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/types/resource"
)

const (
	// ModuleName defines the module name
	ModuleName = "permission"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_permission"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	BucketPolicyForAccountPrefix = []byte{0x11}
	ObjectPolicyForAccountPrefix = []byte{0x12}
	GroupPolicyForAccountPrefix  = []byte{0x13}

	BucketPolicyForGroupPrefix = []byte{0x21}
	ObjectPolicyForGroupPrefix = []byte{0x22}

	PolicyByIDPrefix = []byte{0x31}

	PolicySequencePrefix = []byte{0x41}

	GroupMemberPolicyPrefix = []byte{0x51}
)

func GetPolicyForAccountKey(resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) []byte {
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		return GetBucketPolicyForAccountKey(resourceID, addr)
	case resource.RESOURCE_TYPE_OBJECT:
		return GetObjectPolicyForAccountKey(resourceID, addr)
	case resource.RESOURCE_TYPE_GROUP:
		return GetGroupPolicyForAccountKey(resourceID, addr)
	default:
		return nil
	}
}

func GetBucketPolicyForAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(BucketPolicyForAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetObjectPolicyForAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(ObjectPolicyForAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetGroupPolicyForAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(GroupPolicyForAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetPolicyForGroupKey(resourceID math.Uint, resourceType resource.ResourceType) []byte {
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		return GetBucketPolicyForGroupKey(resourceID)
	case resource.RESOURCE_TYPE_OBJECT:
		return GetObjectPolicyForGroupKey(resourceID)
	default:
		return nil
	}
}

func GetBucketPolicyForGroupKey(resourceID math.Uint) []byte {
	return append(BucketPolicyForGroupPrefix, resourceID.Bytes()...)
}

func GetObjectPolicyForGroupKey(resourceID math.Uint) []byte {
	return append(ObjectPolicyForGroupPrefix, resourceID.Bytes()...)
}

func GetPolicyByIDKey(policyID math.Uint) []byte {
	return append(PolicyByIDPrefix, policyID.Bytes()...)
}

func GetGroupMemberPolicyPrefix(groupID math.Uint, member sdk.AccAddress) []byte {
	return append(GroupMemberPolicyPrefix, append(groupID.Bytes(), member.Bytes()...)...)
}
