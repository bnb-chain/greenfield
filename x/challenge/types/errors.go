package types

import (
	"cosmossdk.io/errors"
)

// x/challenge module sentinel errors
var (
	ErrUnknownSp               = errors.Register(ModuleName, 1, "unknown storage provider")
	ErrUnknownBucketObject     = errors.Register(ModuleName, 2, "unknown bucket info or object info")
	ErrInvalidSpStatus         = errors.Register(ModuleName, 3, "invalid storage provider status")
	ErrInvalidObjectStatus     = errors.Register(ModuleName, 4, "invalid object status to challenge")
	ErrNotStoredOnSp           = errors.Register(ModuleName, 5, "the object is not stored on the storage provider")
	ErrExistsRecentSlash       = errors.Register(ModuleName, 6, "the storage provider and object info had been slashed recently")
	ErrInvalidSegmentIndex     = errors.Register(ModuleName, 7, "the segment/piece index is invalid")
	ErrInvalidChallengeId      = errors.Register(ModuleName, 8, "invalid challenge id")
	ErrInvalidVoteResult       = errors.Register(ModuleName, 9, "invalid vote result")
	ErrInvalidVoteValidatorSet = errors.Register(ModuleName, 10, "invalid validator set")
	ErrInvalidVoteAggSignature = errors.Register(ModuleName, 11, "invalid bls signature")
	ErrDuplicatedSlash         = errors.Register(ModuleName, 12, "duplicated slash in cooling-off period")
	ErrInvalidBlsPubKey        = errors.Register(ModuleName, 13, "invalid bls public key")
	ErrNotEnoughVotes          = errors.Register(ModuleName, 14, "attest votes are not enough")
	ErrNotChallenger           = errors.Register(ModuleName, 15, "not a valid challenger")
	ErrNotInturnChallenger     = errors.Register(ModuleName, 16, "challenger is not in turn")
	ErrInvalidParams           = errors.Register(ModuleName, 17, "invalid params")
	ErrFailToSubmit            = errors.Register(ModuleName, 18, "fail to submit challenge")
)
