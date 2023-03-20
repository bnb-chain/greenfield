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

var (
	BucketPolicyForAccountPrefix = []byte{0x11}
	ObjectPolicyForAccountPrefix = []byte{0x12}
	GroupPolicyForAccountPrefix  = []byte{0x13}
	GroupMemberPrefix            = []byte{0x14}

	BucketPolicyForGroupPrefix = []byte{0x21}
	ObjectPolicyForGroupPrefix = []byte{0x22}

	PolicyByIDPrefix      = []byte{0x31}
	GroupMemberByIDPrefix = []byte{0x32}

	PolicySequencePrefix      = []byte{0x41}
	GroupMemberSequencePrefix = []byte{0x42}
)

func GetPolicyForAccountKey(resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) []byte {
	var key []byte
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		key = BucketPolicyForAccountPrefix
	case resource.RESOURCE_TYPE_OBJECT:
		key = ObjectPolicyForAccountPrefix
	case resource.RESOURCE_TYPE_GROUP:
		key = GroupPolicyForAccountPrefix
	default:
		// todo(quality): better to return error explicitly
		return nil
	}
<<<<<<< HEAD
	key = append(key, resourceID.Bytes()...)
	key = append(key, addr.Bytes()...)
	return key
=======
}

// todo(quality): The only usage of `GetBucketPolicyForAccountKey` is in `GetPolicyForAccountKey`,
// and the patterns are the same. Recommend to delete these functions and implement them in `GetPolicyForAccountKey`.

//func GetPolicyForAccountKey(resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) []byte {
//	var key []byte
//	switch resourceType {
//	case resource.RESOURCE_TYPE_BUCKET:
//		key = BucketPolicyForAccountPrefix
//	case resource.RESOURCE_TYPE_OBJECT:
//		key = ObjectPolicyForAccountPrefix
//	case resource.RESOURCE_TYPE_GROUP:
//		key = GroupPolicyForAccountPrefix
//	default:
//		// todo(quality): better to return error explicitly
//		return nil
//	}
//	key = append(key, resourceID.Bytes()...)
//	key = append(key, addr.Bytes()...)
//	return key
//}

func GetBucketPolicyForAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(BucketPolicyForAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetObjectPolicyForAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(ObjectPolicyForAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
}

func GetGroupPolicyForAccountKey(resourceID math.Uint, addr sdk.AccAddress) []byte {
	return append(GroupPolicyForAccountPrefix, append(resourceID.Bytes(), addr.Bytes()...)...)
>>>>>>> 7384bc55 (chore: refine permission module)
}

func GetPolicyForGroupKey(resourceID math.Uint, resourceType resource.ResourceType) []byte {
	var key []byte
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		key = BucketPolicyForGroupPrefix
	case resource.RESOURCE_TYPE_OBJECT:
		key = ObjectPolicyForGroupPrefix
	default:
		return nil
	}
	key = append(key, resourceID.Bytes()...)
	return key
}

func GetPolicyByIDKey(policyID math.Uint) []byte {
	return append(PolicyByIDPrefix, policyID.Bytes()...)
}

func GetGroupMemberKey(groupID math.Uint, member sdk.AccAddress) []byte {
	return append(GroupMemberPrefix, append(groupID.Bytes(), member.Bytes()...)...)
}
func GetGroupMemberByIDKey(memberID math.Uint) []byte {
	return append(GroupMemberByIDPrefix, memberID.Bytes()...)
}
