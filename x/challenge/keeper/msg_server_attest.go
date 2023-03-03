package keeper

import (
	"context"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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

	if msg.VoteResult == types.CHALLENGE_SUCCEED {
		// check slash
		if k.ExistsSlash(ctx, strings.ToLower(msg.SpOperatorAddress), msg.ObjectId) {
			return nil, types.ErrDuplicatedSlash
		}

		// do slash & reward
		objectSize := objectInfo.PayloadSize
		err = k.doSlashAndRewards(ctx, uint64(objectSize), msg, validators)
		if err != nil {
			return nil, err
		}

		slash := types.Slash{
			SpOperatorAddress: strings.ToLower(msg.SpOperatorAddress),
			ObjectId:          msg.ObjectId,
			Height:            uint64(ctx.BlockHeight()),
		}
		k.SaveSlash(ctx, slash)
	}

	// check whether it is a heartbeat, and will trigger rewards
	heartbeatInterval := k.HeartbeatInterval(ctx)
	if msg.ChallengeId%heartbeatInterval == 0 {
		// reward tx validator & submitter
		err = k.doHeartbeatAndRewards(ctx, msg)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgAttestResponse{}, nil
}

func (k msgServer) calculateSlashAmount(ctx sdk.Context, objectSize uint64) sdkmath.Int {
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

func (k msgServer) calculateSlashRewards(ctx sdk.Context, total sdkmath.Int, challenger string, validators int64) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
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
		submitterReward = total
	}
	return challengerReward, eachValidatorReward, submitterReward
}

func (k msgServer) doSlashAndRewards(ctx sdk.Context, objectSize uint64, msg *types.MsgAttest, validators []string) error {
	submitter := msg.Submitter
	challenger := msg.ChallengerAddress

	slashAmount := k.calculateSlashAmount(ctx, objectSize)
	challengerReward, eachValidatorReward, submitterReward := k.calculateSlashRewards(ctx, slashAmount,
		challenger, int64(len(validators)))

	denom := k.SpKeeper.DepositDenomForSP(ctx)
	rewards := make([]sptypes.RewardInfo, 0)
	if challenger != "" {
		rewards = append(rewards, sptypes.RewardInfo{
			Address: challenger,
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
		Address: submitter,
		Amount: sdk.Coin{
			Denom:  denom,
			Amount: submitterReward,
		}})
	spOperatorAddress, err := sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return err
	}
	err = k.SpKeeper.Slash(ctx, spOperatorAddress, rewards)
	if err != nil {
		return err
	}

	event := types.EventCompleteChallenge{
		ChallengeId:            msg.ChallengeId,
		Result:                 msg.VoteResult,
		SpOperatorAddress:      msg.SpOperatorAddress,
		SlashAmount:            slashAmount.String(),
		ChallengerAddress:      challenger,
		ChallengerRewardAmount: challengerReward.String(),
		SubmitterAddress:       submitter,
		SubmitterRewardAmount:  submitterReward.String(),
		ValidatorAddresses:     validators,
		ValidatorRewardAmount:  eachValidatorReward.String(),
	}
	return ctx.EventManager().EmitTypedEvents(&event)
}

func (k msgServer) calculateHeartbeatRewards(ctx sdk.Context, total sdkmath.Int) (sdkmath.Int, sdkmath.Int) {
	var submitterReward sdkmath.Int
	threshold := k.HeartbeatRewardThreshold(ctx)
	rated := k.HeartbeatRewardRate(ctx).Mul(sdk.NewDecFromInt(total)).TruncateInt()
	if rated.GT(threshold) {
		submitterReward = threshold
	} else {
		submitterReward = rated
	}

	return total.Sub(submitterReward), submitterReward
}

func (k msgServer) doHeartbeatAndRewards(ctx sdk.Context, msg *types.MsgAttest) error {
	submitterAddress, err := sdk.AccAddressFromHexUnsafe(msg.Submitter)
	if err != nil {
		return err
	}

	reward := k.paymentKeeper.QueryValidatorRewards(ctx)
	totalAmount := reward.Amount
	denom := reward.Denom

	if !totalAmount.IsZero() {
		validatorReward, submitterReward := k.calculateHeartbeatRewards(ctx, totalAmount)
		if validatorReward.IsPositive() && submitterReward.IsPositive() {
			toValidator := sdk.Coins{
				sdk.Coin{Denom: denom, Amount: validatorReward},
			}
			distModuleAcc := authtypes.NewModuleAddress(distributiontypes.ModuleName)
			err = k.paymentKeeper.TransferValidatorRewards(ctx, distModuleAcc, toValidator)
			if err != nil {
				return err
			}
			toSubmitter := sdk.Coins{
				sdk.Coin{Denom: denom, Amount: submitterReward},
			}
			err = k.paymentKeeper.TransferValidatorRewards(ctx, submitterAddress, toSubmitter)
			if err != nil {
				return err
			}
		}
	}
	return ctx.EventManager().EmitTypedEvents(&types.EventChallengeHeartbeat{
		ChallengeId: msg.ChallengeId,
	})
}
