package keeper

import (
	"context"
	"math/big"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

func (k msgServer) Attest(goCtx context.Context, msg *types.MsgAttest) (*types.MsgAttestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	submitter, err := sdk.AccAddressFromHexUnsafe(msg.Submitter)
	if err != nil {
		return nil, err
	}
	spOperator, err := sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return nil, err
	}
	challenger := sdk.AccAddress{}
	if msg.ChallengerAddress != "" {
		challenger, err = sdk.AccAddressFromHexUnsafe(msg.ChallengerAddress)
		if err != nil {
			return nil, err
		}
	}

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

	if msg.VoteResult == types.CHALLENGE_SUCCEED {
		// check slash
		if k.ExistsSlash(ctx, spOperator, msg.ObjectId) {
			return nil, types.ErrDuplicatedSlash
		}

		// do slash & reward
		objectSize := objectInfo.PayloadSize
		err = k.doSlashAndRewards(ctx, msg.ChallengeId, msg.VoteResult, objectSize, spOperator, submitter, challenger, validators)
		if err != nil {
			return nil, err
		}

		slash := types.Slash{
			SpOperatorAddress: spOperator,
			ObjectId:          msg.ObjectId,
			Height:            uint64(ctx.BlockHeight()),
		}
		k.SaveSlash(ctx, slash)
	} else {
		// check whether it is a heartbeat attest
		heartbeatInterval := k.HeartbeatInterval(ctx)
		if msg.ChallengeId%heartbeatInterval != 0 {
			return nil, errors.Wrapf(types.ErrInvalidChallengeId, "heart challenge should be submitted at interval %d", heartbeatInterval)
		}

		// reward validators & tx submitter
		err = k.doHeartbeatAndRewards(ctx, msg.ChallengeId, msg.VoteResult, spOperator, submitter, challenger)
		if err != nil {
			return nil, err
		}

	}
	k.SetAttestChallengeId(ctx, msg.ChallengeId)

	return &types.MsgAttestResponse{}, nil
}

func (k msgServer) calculateSlashAmount(ctx sdk.Context, objectSize uint64) sdkmath.Int {
	sizeRate := k.SlashAmountSizeRate(ctx)
	objectSizeInGB := sdk.NewDecFromBigInt(new(big.Int).SetUint64(objectSize)).QuoRoundUp(sdk.NewDec(1073741824))
	slashAmount := objectSizeInGB.MulMut(sizeRate).MulMut(sdk.NewDec(1e18)).TruncateInt()

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

func (k msgServer) calculateSlashRewards(ctx sdk.Context, total sdkmath.Int, challenger sdk.AccAddress, validators int64) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	challengerReward := sdkmath.ZeroInt()
	var eachValidatorReward sdkmath.Int

	threshold := k.RewardSubmitterThreshold(ctx)
	submitterReward := k.RewardSubmitterRatio(ctx).Mul(sdk.NewDecFromInt(total)).TruncateInt()
	if submitterReward.GT(threshold) {
		submitterReward = threshold
	}
	total = total.Sub(submitterReward)

	if challenger.Equals(sdk.AccAddress{}) { // the challenge is triggered by blockchain automatically
		eachValidatorReward = total.Quo(sdk.NewIntFromUint64(uint64(validators)))
		for i := int64(0); i < validators; i++ {
			total = total.Sub(eachValidatorReward)
		}
		// send remaining to submitter
		submitterReward = submitterReward.Add(total)
	} else { // the challenge is submitted by challenger
		validatorRatio := k.RewardValidatorRatio(ctx)
		eachValidatorReward = validatorRatio.MulInt(total).QuoInt64(validators).TruncateInt()
		for i := int64(0); i < validators; i++ {
			total = total.Sub(eachValidatorReward)
		}
		// the left is rewarded to challenger
		challengerReward = total
	}
	return challengerReward, eachValidatorReward, submitterReward
}

func (k msgServer) doSlashAndRewards(ctx sdk.Context, challengeId uint64, voteResult types.VoteResult, objectSize uint64,
	spOperator, submitter, challenger sdk.AccAddress, validators []string) error {

	slashAmount := k.calculateSlashAmount(ctx, objectSize)
	challengerReward, eachValidatorReward, submitterReward := k.calculateSlashRewards(ctx, slashAmount,
		challenger, int64(len(validators)))

	denom := k.SpKeeper.DepositDenomForSP(ctx)
	rewards := make([]sptypes.RewardInfo, 0)
	if !challenger.Equals(sdk.AccAddress{}) {
		rewards = append(rewards, sptypes.RewardInfo{
			Address: challenger.String(),
			Amount: sdk.Coin{
				Denom:  denom,
				Amount: challengerReward,
			},
		})
	}
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
		Address: submitter.String(),
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: submitterReward,
		}})

	err := k.SpKeeper.Slash(ctx, spOperator, rewards)
	if err != nil {
		return err
	}

	event := types.EventAttestChallenge{
		ChallengeId:            challengeId,
		Result:                 voteResult,
		SpOperatorAddress:      spOperator.String(),
		SlashAmount:            slashAmount.String(),
		ChallengerAddress:      challenger.String(),
		ChallengerRewardAmount: challengerReward.String(),
		SubmitterAddress:       submitter.String(),
		SubmitterRewardAmount:  submitterReward.String(),
		ValidatorRewardAmount:  eachValidatorReward.MulRaw(int64(len(validators))).String(),
	}
	return ctx.EventManager().EmitTypedEvents(&event)
}

func (k msgServer) calculateHeartbeatRewards(ctx sdk.Context, total sdkmath.Int) (sdkmath.Int, sdkmath.Int) {
	threshold := k.RewardSubmitterThreshold(ctx)
	submitterReward := k.RewardSubmitterRatio(ctx).Mul(sdk.NewDecFromInt(total)).TruncateInt()
	if submitterReward.GT(threshold) {
		submitterReward = threshold
	}

	return total.Sub(submitterReward), submitterReward
}

func (k msgServer) doHeartbeatAndRewards(ctx sdk.Context, challengeId uint64, voteResult types.VoteResult,
	spOperator, submitter, challenger sdk.AccAddress) error {
	totalAmount := k.paymentKeeper.QueryValidatorRewards(ctx)

	validatorReward, submitterReward := sdkmath.NewInt(0), sdkmath.NewInt(0)
	if !totalAmount.IsZero() {
		validatorReward, submitterReward = k.calculateHeartbeatRewards(ctx, totalAmount)
		if validatorReward.IsPositive() && submitterReward.IsPositive() {
			distModuleAcc := authtypes.NewModuleAddress(distributiontypes.ModuleName)
			err := k.paymentKeeper.TransferValidatorRewards(ctx, distModuleAcc, validatorReward)
			if err != nil {
				return err
			}
			err = k.paymentKeeper.TransferValidatorRewards(ctx, submitter, submitterReward)
			if err != nil {
				return err
			}
		}
	}
	return ctx.EventManager().EmitTypedEvents(&types.EventAttestChallenge{
		ChallengeId:            challengeId,
		Result:                 voteResult,
		SpOperatorAddress:      spOperator.String(),
		SlashAmount:            "",
		ChallengerAddress:      challenger.String(),
		ChallengerRewardAmount: "",
		SubmitterAddress:       submitter.String(),
		SubmitterRewardAmount:  submitterReward.String(),
		ValidatorRewardAmount:  validatorReward.String(),
	})
}
