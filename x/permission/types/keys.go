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
	BucketPolicyToAccountPrefix = []byte{0x11}
	ObjectPolicyToAccountPrefix = []byte{0x12}
	GroupPolicyToAccountPrefix  = []byte{0x13}

	BucketPolicyToGroupPrefix = []byte{0x21}
	ObjectPolicyToGroupPrefix = []byte{0x22}

	PolicyByIDPrefix = []byte{0x31}

	PolicySequencePrefix = []byte{0x41}
)

func GetPolicyToAccountKey(resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) []byte {
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		return GetBucketPolicyToAccountKey(resourceID, addr)
	case resource.RESOURCE_TYPE_OBJECT:
		return GetObjectPolicyToAccountKey(resourceID, addr)
	case resource.RESOURCE_TYPE_GROUP:
		return GetGroupPolicyToAccountKey(resourceID, addr)
	default:
		return nil
	}
}

func GetBucketPolicyToAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(BucketPolicyToAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetObjectPolicyToAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(ObjectPolicyToAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetGroupPolicyToAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(GroupPolicyToAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetPolicyToGroupKey(resourceID math.Uint, resourceType resource.ResourceType) []byte {
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		return GetBucketPolicyToGroupKey(resourceID)
	case resource.RESOURCE_TYPE_OBJECT:
		return GetObjectPolicyToGroupKey(resourceID)
	default:
		return nil
	}
}

func GetBucketPolicyToGroupKey(resourceID math.Uint) []byte {
	return append(BucketPolicyToGroupPrefix, resourceID.Bytes()...)
}

func GetObjectPolicyToGroupKey(resourceID math.Uint) []byte {
	return append(ObjectPolicyToGroupPrefix, resourceID.Bytes()...)
}

func GetPolicyByIDKey(policyID math.Uint) []byte {
	return append(PolicyByIDPrefix, policyID.Bytes()...)
}
