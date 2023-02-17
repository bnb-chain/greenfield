package keeper

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/bits-and-blooms/bitset"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/prysmaticlabs/prysm/crypto/bls"
)

func (k msgServer) Attest(goCtx context.Context, msg *types.MsgAttest) (*types.MsgAttestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check challenge
	challenge, found := k.GetOngoingChallenge(ctx, msg.ChallengeId)
	if !found {
		return nil, types.ErrUnknownChallenge
	}

	//check object, and get object info
	objectKey := challenge.ObjectKey
	objectInfo, found := k.StorageKeeper.GetObjectWithKey(ctx, objectKey)
	if !found {
		return nil, types.ErrUnknownObject
	}
	objectSize := objectInfo.PayloadSize
	if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_IN_SERVICE {
		return nil, types.ErrInvalidObjectStatus
	}

	// check attest validators and signatures
	validators, err := k.verifySignature(ctx, msg)
	if err != nil {
		return nil, err
	}

	// check slash
	recentSlashes := k.GetAllRecentSlash(ctx)
	for _, slash := range recentSlashes {
		if slash.SpOperatorAddress == challenge.SpOperatorAddress &&
			bytes.Equal(slash.ObjectKey, challenge.GetObjectKey()) {
			return nil, types.ErrDuplicatedSlash
		}
	}

	// do slash & reward
	k.doSlashAndRewards(ctx, uint64(objectSize), challenge, msg.Creator, validators)

	k.RemoveOngoingChallenge(ctx, msg.ChallengeId)
	slash := types.Slash{
		SpOperatorAddress: challenge.SpOperatorAddress,
		ObjectKey:         challenge.ObjectKey,
		Height:            uint64(ctx.BlockHeight()),
	}
	k.AppendRecentSlash(ctx, slash)

	return &types.MsgAttestResponse{}, nil
}

func (k Keeper) verifySignature(ctx sdk.Context, attest *types.MsgAttest) ([]string, error) {
	historicalInfo, ok := k.stakingKeeper.GetHistoricalInfo(ctx, ctx.BlockHeight())
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrInvalidVoteValidatorSet, "fail to get validators")
	}
	validators := historicalInfo.Valset

	validatorsBitSet := bitset.From(attest.VoteValidatorSet)
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

	aggSig, err := bls.SignatureFromBytes(attest.VoteAggSignature)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidVoteAggSignature, fmt.Sprintf("BLS signature converts failed: %v", err))
	}

	if !aggSig.FastAggregateVerify(votedPubKeys, attest.GetBlsSignBytes()) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidVoteAggSignature, "Signature verify failed")
	}

	return signedRelayers, nil
}

func (k msgServer) calculateSlashAmount(ctx sdk.Context, objectSize uint64) sdkmath.Int {
	perKb := k.SlashAmountSizeRate(ctx)
	decSize := sdk.NewDecFromBigInt(new(big.Int).SetUint64(objectSize))
	decRoot, err := decSize.ApproxSqrt()
	if err != nil {
		panic(err)
	}
	slashAmount := decRoot.MulMut(perKb).TruncateInt()

	min := k.SlashAmountMin(ctx)
	if slashAmount.LT(min) {
		return min
	}
	max := k.SlashAmountMax(ctx)
	if slashAmount.GT(max) {
		return max
	}
	return slashAmount
}

func (k msgServer) calculateRewardAmounts(ctx sdk.Context, total sdkmath.Int, challenger, submitter string, validators int64) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	challengerReward := sdk.ZeroInt()
	eachValidatorReward := sdk.ZeroInt()
	submitterReward := sdk.ZeroInt()
	if challenger == "" { // the challenge is submitted by challenger
		challengerRatio := k.RewardChallengerRatio(ctx)
		challengerReward = challengerRatio.MulInt(total).TruncateInt()

		validatorRatio := k.RewardValidatorRatio(ctx)
		eachValidatorReward = validatorRatio.MulInt(total).QuoInt64(validators).TruncateInt()
		for i := int64(0); i < validators; i++ {
			total = total.Sub(eachValidatorReward)
		}

		submitterReward = total.Sub(challengerReward)
	} else { // the challenge is triggered by blockchain automatically
		eachValidatorReward = total.Quo(sdk.NewIntFromUint64(uint64(validators)))
		for i := int64(0); i < validators; i++ {
			total = total.Sub(eachValidatorReward)
		}
		submitterReward = total.Sub(challengerReward)
	}
	return challengerReward, eachValidatorReward, submitterReward
}

func (k msgServer) doSlashAndRewards(ctx sdk.Context, objectSize uint64,
	challenge types.Challenge, submitter string, validators []string) {
	slashAmount := k.calculateSlashAmount(ctx, objectSize)
	challengerReward, eachValidatorReward, submitterReward := k.calculateRewardAmounts(ctx, slashAmount,
		challenge.ChallengerAddress, submitter, int64(len(validators)))

	denom := k.SlashDenom(ctx)
	rewards := make([]sptypes.RewardInfo, 0)
	rewards = append(rewards, sptypes.RewardInfo{
		Address: challenge.ChallengerAddress,
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: challengerReward,
		},
	})
	for _, val := range validators {
		rewards = append(rewards, sptypes.RewardInfo{
			Address: val,
			Amount: sdk.Coin{
				Denom:  denom,
				Amount: eachValidatorReward,
			},
		})
	}
	rewards = append(rewards, sptypes.RewardInfo{
		Address: submitter,
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: submitterReward,
		}})
	spOperatorAddress, err := sdk.AccAddressFromHexUnsafe(challenge.SpOperatorAddress)
	if err != nil {
		panic(err)
	}
	k.SpKeeper.Slash(ctx, spOperatorAddress, rewards)

	event := types.EventCompleteChallenge{
		ChallengeId:            challenge.Id,
		Result:                 0,
		SpOperatorAddress:      challenge.SpOperatorAddress,
		SlashAmount:            slashAmount.String(),
		ChallengerAddress:      challenge.ChallengerAddress,
		ChallengerRewardAmount: challengerReward.String(),
		SubmitterAddress:       submitter,
		SubmitterRewardAmount:  submitterReward.String(),
		ValidatorAddresses:     validators,
		ValidatorRewardAmount:  eachValidatorReward.String(),
	}
	_ = ctx.EventManager().EmitTypedEvents(&event)
}
