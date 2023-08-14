package types

import (
	"fmt"
	"math/big"
	"time"

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

	FormatTimeBytesLength = 29 // len(sdk.FormatTimeBytes(time.Now()))
)

var (
	ParamsKey = []byte{0x01}

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

	PolicyQueueKeyPrefix = []byte{0x51}
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
	return append(GroupMemberPrefix, LengthPrefix(groupID)...)
}

func GetGroupMemberKey(groupID math.Uint, member sdk.AccAddress) []byte {
	return append(GroupMemberPrefix, append(LengthPrefix(groupID), member.Bytes()...)...)
}

func GetGroupMemberByIDKey(memberID math.Uint) []byte {
	return append(GroupMemberByIDPrefix, memberID.Bytes()...)
}

// PolicyPrefixQueue is the canonical key to store policy key.
//
// Key format:
// - <key_prefix><exp_bytes><policy_id_bytes>
func PolicyPrefixQueue(exp *time.Time, key []byte) []byte {
	policyByExpTimeKey := PolicyByExpTimeKey(exp)
	return append(policyByExpTimeKey, key...)
}

// PolicyByExpTimeKey returns a key with key prefix, expiry
//
// Key format:
// - <key_prefix><exp_bytes>
func PolicyByExpTimeKey(exp *time.Time) []byte {
	// no need of appending len(exp_bytes) here, `FormatTimeBytes` gives const length everytime.
	return append(PolicyQueueKeyPrefix, sdk.FormatTimeBytes(*exp)...)
}

func ParsePolicyIdFromQueueKey(key []byte) math.Uint {
	// key is of format:
	// <key_prefix><expiration_bytes(fixed length)><policy_id_bytes>
	bz := key[FormatTimeBytesLength+1:]
	return math.NewUintFromBigInt(new(big.Int).SetBytes(bz))
}

func LengthPrefix(id math.Uint) []byte {
	bz := id.Bytes()
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}

	return append([]byte{byte(bzLen)}, bz...)
}
