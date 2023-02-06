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
	// OngoingChallengeKeyPrefix is the prefix to retrieve all OngoingChallenge
	OngoingChallengeKeyPrefix = "OngoingChallenge/value/"

	RecentSlashKey      = "RecentSlash/value/"
	RecentSlashCountKey = "RecentSlash/count/"
)

// OngoingChallengeKey returns the store key to retrieve a OngoingChallenge from the index fields
func OngoingChallengeKey(
	id uint64,
) []byte {
	var key []byte

	challengeIdBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(challengeIdBytes, uint64(id))
	key = append(key, challengeIdBytes...)
	key = append(key, []byte("/")...)

	return key
}
