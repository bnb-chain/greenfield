package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/challenge module sentinel errors
var (
	ErrUnknownSp               = sdkerrors.Register(ModuleName, 1, "unknown storage provider")
	ErrUnknownObject           = sdkerrors.Register(ModuleName, 2, "unknown object info")
	ErrInvalidSpStatus         = sdkerrors.Register(ModuleName, 3, "invalid storage provider status")
	ErrInvalidObjectStatus     = sdkerrors.Register(ModuleName, 4, "invalid object status to challenge")
	ErrNotStoredOnSp           = sdkerrors.Register(ModuleName, 5, "the object is not stored on the storage provider")
	ErrExistsRecentSlash       = sdkerrors.Register(ModuleName, 6, "the storage provider and object info had been slashed recently")
	ErrInvalidSegmentIndex     = sdkerrors.Register(ModuleName, 7, "the segment/piece index is invalid")
	ErrInvalidChallengeId      = sdkerrors.Register(ModuleName, 8, "invalid challenge id")
	ErrInvalidVoteResult       = sdkerrors.Register(ModuleName, 9, "invalid vote result")
	ErrInvalidVoteValidatorSet = sdkerrors.Register(ModuleName, 10, "invalid validator set")
	ErrInvalidVoteAggSignature = sdkerrors.Register(ModuleName, 11, "invalid bls signature")
	ErrDuplicatedSlash         = sdkerrors.Register(ModuleName, 12, "duplicated slash in cooling-off period")
	ErrInvalidBlsPubKey        = sdkerrors.Register(ModuleName, 13, "invalid bls public key")
	ErrNotEnoughVotes          = sdkerrors.Register(ModuleName, 14, "attest votes are not enough")
)
