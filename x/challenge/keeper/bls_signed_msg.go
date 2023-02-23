package keeper

import (
	"fmt"

	"github.com/bits-and-blooms/bitset"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

type BlsSignedMsg interface {
	GetBlsSignBytes() [32]byte
	GetVoteValidatorSet() []uint64
	GetVoteAggSignature() []byte
}

func (k Keeper) verifySignature(ctx sdk.Context, signedMsg BlsSignedMsg) ([]string, error) {
	historicalInfo, ok := k.stakingKeeper.GetHistoricalInfo(ctx, ctx.BlockHeight())
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrInvalidVoteValidatorSet, "fail to get validators")
	}
	validators := historicalInfo.Valset

	validatorsBitSet := bitset.From(signedMsg.GetVoteValidatorSet())
	if validatorsBitSet.Count() > uint(len(validators)) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidVoteValidatorSet, "number of validator set is larger than validators")
	}

	signedRelayers := make([]string, 0, validatorsBitSet.Count())
	votedPubKeys := make([]bls.PublicKey, 0, validatorsBitSet.Count())
	for index, val := range validators {
		if !validatorsBitSet.Test(uint(index)) {
			continue
		}

		signedRelayers = append(signedRelayers, val.RelayerAddress)
		votePubKey, err := bls.PublicKeyFromBytes(val.RelayerBlsKey)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrInvalidBlsPubKey, fmt.Sprintf("BLS public key converts failed: %v", err))
		}
		votedPubKeys = append(votedPubKeys, votePubKey)
	}

	if len(votedPubKeys) <= len(validators)*2/3 {
		return nil, sdkerrors.Wrapf(types.ErrNotEnoughVotes, fmt.Sprintf("Not enough validators voted, need: %d, voted: %d", len(validators)*2/3, len(votedPubKeys)))
	}

	aggSig, err := bls.SignatureFromBytes(signedMsg.GetVoteAggSignature())
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidVoteAggSignature, fmt.Sprintf("BLS signature converts failed: %v", err))
	}

	if !aggSig.FastAggregateVerify(votedPubKeys, signedMsg.GetBlsSignBytes()) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidVoteAggSignature, "Signature verify failed")
	}

	return signedRelayers, nil
}
