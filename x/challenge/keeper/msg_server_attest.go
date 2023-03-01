package keeper

import (
	"context"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

func (k msgServer) Attest(goCtx context.Context, msg *types.MsgAttest) (*types.MsgAttestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ongoingId := k.GetOngoingChallengeId(ctx)
	attestedId := k.GetAttestChallengeId(ctx)

	if msg.ChallengeId <= attestedId || msg.ChallengeId > ongoingId {
		return nil, types.ErrInvalidChallengeId
	}

	//check object, and get object info
	objectInfo, found := k.StorageKeeper.GetObjectInfoById(ctx, msg.ObjectId)
	if !found { // be noted: even the object info is not in service now, we will continue slash the storage provider
		return nil, types.ErrUnknownObject
	}

	// check attest validators and signatures
	validators, err := k.verifySignature(ctx, msg)
	if err != nil {
		return nil, err
	}

	// check slash
	if k.ExistsSlash(ctx, strings.ToLower(msg.SpOperatorAddress), msg.ObjectId) {
		return nil, types.ErrDuplicatedSlash
	}

	// do slash & reward
	objectSize := objectInfo.PayloadSize
	k.doSlashAndRewards(ctx, uint64(objectSize), msg, validators)

	slash := types.Slash{
		SpOperatorAddress: strings.ToLower(msg.SpOperatorAddress),
		ObjectId:          msg.ObjectId,
		Height:            uint64(ctx.BlockHeight()),
	}
	k.SaveSlash(ctx, slash)
	k.RemoveChallengeUntil(ctx, msg.ChallengeId)

	return &types.MsgAttestResponse{}, nil
}

func (k msgServer) calculateChallengeSlash(ctx sdk.Context, objectSize uint64) sdkmath.Int {
	sizeRate := k.SlashAmountSizeRate(ctx)
	decSize := sdk.NewDecFromBigInt(new(big.Int).SetUint64(objectSize))
	decRoot, err := decSize.ApproxSqrt()
	if err != nil {
		panic(err)
	}
	slashAmount := decRoot.MulMut(sizeRate).TruncateInt()

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

func (k msgServer) calculateChallengeRewards(ctx sdk.Context, total sdkmath.Int, challenger, submitter string, validators int64) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	challengerReward := sdkmath.ZeroInt()
	var eachValidatorReward sdkmath.Int
	var submitterReward sdkmath.Int
	if challenger != "" { // the challenge is submitted by challenger
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

func (k msgServer) doSlashAndRewards(ctx sdk.Context, objectSize uint64, msg *types.MsgAttest, validators []string) {
	submitter := msg.Creator
	challenger := ""
	challenge, found := k.GetChallenge(ctx, msg.ChallengeId)
	if found {
		challenger = challenge.ChallengerAddress
	}

	slashAmount := k.calculateChallengeSlash(ctx, objectSize)
	challengerReward, eachValidatorReward, submitterReward := k.calculateChallengeRewards(ctx, slashAmount,
		challenger, submitter, int64(len(validators)))

	denom := k.SpKeeper.DepositDenomForSP(ctx)
	rewards := make([]sptypes.RewardInfo, 0)
	rewards = append(rewards, sptypes.RewardInfo{
		Address: challenger,
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
	spOperatorAddress, err := sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		panic(err)
	}
	err = k.SpKeeper.Slash(ctx, spOperatorAddress, rewards)
	if err != nil {
		panic(err)
	}

	event := types.EventCompleteChallenge{
		ChallengeId:            msg.ChallengeId,
		Result:                 types.ChallengeResultSucceed,
		SpOperatorAddress:      msg.SpOperatorAddress,
		SlashAmount:            slashAmount.String(),
		ChallengerAddress:      challenger,
		ChallengerRewardAmount: challengerReward.String(),
		SubmitterAddress:       submitter,
		SubmitterRewardAmount:  submitterReward.String(),
		ValidatorAddresses:     validators,
		ValidatorRewardAmount:  eachValidatorReward.String(),
	}
	_ = ctx.EventManager().EmitTypedEvents(&event)
}
