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

	// AttestedChallengesPrefix is the prefix to record the latest attested challenges.
	AttestedChallengesPrefix = []byte{0x15}

	// AttestedChallengesSizeKey is the key to record the size of latest attested challenges.
	AttestedChallengesSizeKey = []byte{0x16}

	// AttestedChallengesCursorKey is the key to retrieve the latest attested challenges.
	AttestedChallengesCursorKey = []byte{0x17}

	// SlashAmountKeyPrefix is the prefix to count the amount of Slash for a sp.
	SlashAmountKeyPrefix = []byte{0x18}
)
