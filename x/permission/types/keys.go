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
	ParamsKey = []byte{0x01}

	BucketPolicyForAccountPrefix = []byte{0x11}
	ObjectPolicyForAccountPrefix = []byte{0x12}
	GroupPolicyForAccountPrefix  = []byte{0x13}
	GroupMemberPrefix            = []byte{0x14}
	GroupMemberExtraPrefix       = []byte{0x15}

	BucketPolicyForGroupPrefix = []byte{0x21}
	ObjectPolicyForGroupPrefix = []byte{0x22}

	PolicyByIDPrefix           = []byte{0x31}
	GroupMemberByIDPrefix      = []byte{0x32}
	GroupMemberExtraByIDPrefix = []byte{0x33}

	PolicySequencePrefix      = []byte{0x41}
	GroupMemberSequencePrefix = []byte{0x42}
)

func PolicyForAccountPrefix(resourceID math.Uint, resourceType resource.ResourceType) []byte {
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
	return key
}

func GetPolicyForAccountKey(resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) []byte {
	key := PolicyForAccountPrefix(resourceID, resourceType)
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

func GroupMembersPrefix(groupID math.Uint) []byte {
	return append(GroupMemberPrefix, groupID.Bytes()...)
}

func GetGroupMemberKey(groupID math.Uint, member sdk.AccAddress) []byte {
	return append(GroupMemberPrefix, append(groupID.Bytes(), member.Bytes()...)...)
}

func GetGroupMemberByIDKey(memberID math.Uint) []byte {
	return append(GroupMemberByIDPrefix, memberID.Bytes()...)
}

func GroupMembersExtraPrefix(groupID math.Uint) []byte {
	return append(GroupMemberExtraPrefix, groupID.Bytes()...)
}

func GetGroupMemberExtraKey(groupID math.Uint, member sdk.AccAddress) []byte {
	return append(GroupMemberExtraPrefix, append(groupID.Bytes(), member.Bytes()...)...)
}

func GetGroupMemberExtraByIDKey(memberID math.Uint) []byte {
	return append(GroupMemberExtraByIDPrefix, memberID.Bytes()...)
}
