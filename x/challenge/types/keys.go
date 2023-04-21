package types

const (
	// ModuleName defines the module name
	ModuleName = "challenge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_challenge"

	// TStoreKey defines transient store key
	TStoreKey = "transient_challenge"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	ParamsKey = []byte{0x01}

	// ChallengeIdKey is the key to retrieve the id of challenge.
	ChallengeIdKey = []byte{0x11}

	// ChallengeKeyPrefix is the prefix to retrieve Challenge.
	ChallengeKeyPrefix = []byte{0x12}

	// SlashKeyPrefix is the prefix to retrieve Slash.
	SlashKeyPrefix = []byte{0x13}

	// CurrentBlockChallengeCountKey is key to track the count of challenges in the current block.
	// The data is stored in transient store.
	CurrentBlockChallengeCountKey = []byte{0x14}

	// AttestChallengeIdsPrefix is the prefix to record the latest attested challenge ids.
	AttestChallengeIdsPrefix = []byte{0x15}

	// AttestChallengeIdsSizeKey is the key to record the size of latest attested challenge ids.
	AttestChallengeIdsSizeKey = []byte{0x16}

	// AttestChallengeIdsCursorKey is the key to retrieve the latest attested challenge ids.
	AttestChallengeIdsCursorKey = []byte{0x17}
)
