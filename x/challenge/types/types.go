package types

const (
	// ChallengeResultSucceed stands for the result of succeed challenge
	ChallengeResultSucceed = uint32(1)

	// ChallengeResultFailed stands for the result of failed challenge
	ChallengeResultFailed = uint32(2)
)

// RedundancyIndexPrimary defines the redundancy index for primary storage provider (asked by storage provider api)
const RedundancyIndexPrimary = int32(-1)

// BlsSignatureLength defines the length of bls signature
const BlsSignatureLength = 96
