package types

import (
	"fmt"

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
		panic(fmt.Sprintf("GetPolicyForAccountKey Invalid Resource Type, %s", resourceType.String()))
	}
	key = append(key, resourceID.Bytes()...)
	key = append(key, addr.Bytes()...)
	return key
}

func GetPolicyForGroupKey(resourceID math.Uint, resourceType resource.ResourceType) []byte {
	var key []byte
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		key = BucketPolicyForGroupPrefix
	case resource.RESOURCE_TYPE_OBJECT:
		key = ObjectPolicyForGroupPrefix
	default:
		panic(fmt.Sprintf("GetPolicyForGroupKey Invalid Resource Type, %s", resourceType.String()))
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
