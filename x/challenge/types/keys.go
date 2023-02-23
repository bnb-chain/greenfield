package types

import (
	"encoding/binary"
)

const (
	// ModuleName defines the module name
	ModuleName = "challenge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_challenge"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	// OngoingChallengeKeyPrefix is the prefix to retrieve all Challenge
	OngoingChallengeKeyPrefix = "Challenge/ongoing/"

	// ChallengeIdKey is the key to retrieve the max id of challenge
	ChallengeIdKey = "Challenge/id/"

	// ChallengeCountKey is key to track the count of challenges in the current block
	ChallengeCountKey = "Challenge/count/"

	// HeartbeatChallengeIdKey is the key to record the latest heartbeat challenge id
	HeartbeatChallengeIdKey = "Challenge/heartbeat/"

	RecentSlashKey      = "Slash/value/"
	RecentSlashCountKey = "Slash/count/"
)

// OngoingChallengeKey returns the store key to retrieve a Challenge from the index fields
func OngoingChallengeKey(
	id uint64,
) []byte {
	var key []byte

	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, uint64(id))
	key = append(key, idBytes...)
	key = append(key, []byte("/")...)

	return key
}
