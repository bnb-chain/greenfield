package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/challenge module sentinel errors
var (
	ErrInvalidGenesis = sdkerrors.Register(ModuleName, 1, "invalid genesis state")

	ErrInvalidBucketName   = sdkerrors.Register(ModuleName, 2, "invalid bucket name")
	ErrInvalidObjectName   = sdkerrors.Register(ModuleName, 3, "invalid object name")
	ErrInvalidIndex        = sdkerrors.Register(ModuleName, 4, "invalid segment/piece index")
	ErrUnknownSp           = sdkerrors.Register(ModuleName, 5, "unknown storage provider")
	ErrUnknownObject       = sdkerrors.Register(ModuleName, 6, "unknown object info")
	ErrInvalidSpStatus     = sdkerrors.Register(ModuleName, 7, "invalid storage provider status")
	ErrInvalidObjectStatus = sdkerrors.Register(ModuleName, 8, "invalid object status to challenge")
	ErrNotStoredOnSp       = sdkerrors.Register(ModuleName, 9, "the object is not stored on the storage provider")
	ErrExistsRecentSlash   = sdkerrors.Register(ModuleName, 10, "the storage provider and object info had been slashed recently")
	ErrInvalidSegmentIndex = sdkerrors.Register(ModuleName, 11, "the segment/piece index is invalid")

	ErrInvalidVoteResult       = sdkerrors.Register(ModuleName, 12, "invalid vote result")
	ErrInvalidVoteValidatorSet = sdkerrors.Register(ModuleName, 13, "invalid validator set")
	ErrInvalidVoteAggSignature = sdkerrors.Register(ModuleName, 14, "invalid bls signature")
	ErrUnknownChallenge        = sdkerrors.Register(ModuleName, 15, "unknown challenge")
	ErrDuplicatedSlash         = sdkerrors.Register(ModuleName, 16, "duplicated slash in cooling-off period")
	ErrInvalidBlsPubKey        = sdkerrors.Register(ModuleName, 17, "invalid bls public key")
	ErrNotEnoughVotes          = sdkerrors.Register(ModuleName, 18, "attest votes are not enough")

	ErrInvalidChallengeId = sdkerrors.Register(ModuleName, 19, "invalid challenge id")
)
