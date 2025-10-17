package keeper

import (
	"cosmossdk.io/errors"
	"github.com/bits-and-blooms/bitset"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// BlsSignedMsg defined the interface of a bls signed message.
type BlsSignedMsg interface {
	// GetBlsSignBytes returns the bls signed message in bytes.
	GetBlsSignBytes(chainId string) [32]byte

	// GetVoteValidatorSet returns the validators who signed the message.
	GetVoteValidatorSet() []uint64

	// GetVoteAggSignature returns the aggregated bls signature.
	GetVoteAggSignature() []byte
}

// verifySignature verifies whether the signature is valid or not.
func (k Keeper) verifySignature(ctx sdk.Context, signedMsg BlsSignedMsg, validators []stakingtypes.Validator) ([]string, error) {
	validatorsBitSet := bitset.From(signedMsg.GetVoteValidatorSet())
	if validatorsBitSet.Count() > uint(len(validators)) {
		return nil, errors.Wrap(types.ErrInvalidVoteValidatorSet, "number of validator set is larger than validators")
	}

	signedChallengers := make([]string, 0, validatorsBitSet.Count())
	votedPubKeys := make([]bls.PublicKey, 0, validatorsBitSet.Count())
	for index, val := range validators {
		if !validatorsBitSet.Test(uint(index)) {
			continue
		}

		signedChallengers = append(signedChallengers, val.ChallengerAddress)
		votePubKey, err := bls.PublicKeyFromBytes(val.BlsKey)
		if err != nil {
			return nil, errors.Wrapf(types.ErrInvalidBlsPubKey, "BLS public key converts failed: %v", err)
		}
		votedPubKeys = append(votedPubKeys, votePubKey)
	}

	if len(votedPubKeys) <= len(validators)*2/3 {
		return nil, errors.Wrapf(types.ErrNotEnoughVotes, "Not enough validators voted, need: %d, voted: %d", len(validators)*2/3, len(votedPubKeys))
	}

	aggSig, err := bls.SignatureFromBytes(signedMsg.GetVoteAggSignature())
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidVoteAggSignature, "BLS signature converts failed: %v", err)
	}

	if !aggSig.FastAggregateVerify(votedPubKeys, signedMsg.GetBlsSignBytes(ctx.ChainID())) {
		return nil, errors.Wrap(types.ErrInvalidVoteAggSignature, "Signature verify failed")
	}

	return signedChallengers, nil
}
