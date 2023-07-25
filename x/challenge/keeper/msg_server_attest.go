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
	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// Attest handles user's request for attesting a challenge.
// The attestation can include a valid challenge or is only for heartbeat purpose.
// If the challenge is valid, the related storage provider will be slashed.
// For heartbeat attestation, the challenge is invalid and the storage provider will not be slashed.
func (k msgServer) Attest(goCtx context.Context, msg *types.MsgAttest) (*types.MsgAttestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	submitter := sdk.MustAccAddressFromHex(msg.Submitter)
	spOperator := sdk.MustAccAddressFromHex(msg.SpOperatorAddress)

	sp, found := k.SpKeeper.GetStorageProviderByOperatorAddr(ctx, spOperator)
	if !found {
		return nil, errors.Wrapf(types.ErrUnknownSp, "cannot find sp with operator address: %s", msg.SpOperatorAddress)
	}

	challenger := sdk.AccAddress{}
	if msg.ChallengerAddress != "" {
		challenger = sdk.MustAccAddressFromHex(msg.ChallengerAddress)
	}

	if !k.ExistsChallenge(ctx, msg.ChallengeId) {
		return nil, errors.Wrapf(types.ErrInvalidChallengeId, "challenge %d cannot be found, it could be expired", msg.ChallengeId)
	}

	historicalInfo, ok := k.stakingKeeper.GetHistoricalInfo(ctx, ctx.BlockHeight())
	if !ok {
		return nil, errors.Wrap(types.ErrInvalidVoteValidatorSet, "fail to get validators")
	}
	allValidators := historicalInfo.Valset
	inTurn, err := k.isInturnAttestation(ctx, submitter, allValidators)
	if err != nil {
		return nil, err
	}
	if !inTurn {
		return nil, types.ErrNotInturnChallenger
	}

	// check attest validators and signatures
	validators, err := k.verifySignature(ctx, msg, allValidators)
	if err != nil {
		return nil, err
	}

	//check object, and get object info
	objectInfo, found := k.StorageKeeper.GetObjectInfoById(ctx, msg.ObjectId)
	if !found { // be noted: even the object info is not in service now, we will continue slash the storage provider
		return nil, types.ErrUnknownBucketObject
	}

	//for migrating buckets, or swapping out, the process could be done when offline service does the verification
	bucketInfo, found := k.StorageKeeper.GetBucketInfo(ctx, objectInfo.BucketName)
	if !found {
		return nil, storagetypes.ErrNoSuchBucket.Wrapf("bucket not found when attest")
	}

	spInState := k.StorageKeeper.MustGetPrimarySPForBucket(ctx, bucketInfo)

	if spInState.Id != sp.Id {
		gvg, _ := k.StorageKeeper.GetObjectGVG(ctx, bucketInfo.Id, objectInfo.LocalVirtualGroupId)
		found = false
		for _, id := range gvg.SecondarySpIds {
			if id == sp.Id {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.Wrapf(types.ErrNotStoredOnSp, "sp %s does not store the object anymore", sp.OperatorAddress)
		}
	}

	if msg.VoteResult == types.CHALLENGE_SUCCEED {
		// check slash
		if k.ExistsSlash(ctx, sp.Id, msg.ObjectId) {
			return nil, types.ErrDuplicatedSlash
		}

		// check slash amount
		objectSize := objectInfo.PayloadSize
		toSlashAmount := k.calculateSlashAmount(ctx, objectSize)

		slashedAmount := k.GetSpSlashAmount(ctx, sp.Id)
		if !slashedAmount.IsZero() { // if it is the first time to slash, do not check the amount
			maxSlashAmount := k.GetParams(ctx).SpSlashMaxAmount
			if (slashedAmount.Add(toSlashAmount)).GT(maxSlashAmount) {
				return nil, types.ErrExceedMaxSlashAmount
			}
		}

		// do slash & reward
		err = k.doSlashAndRewards(ctx, msg.ChallengeId, msg.VoteResult, toSlashAmount, sp.Id, submitter, challenger, validators)
		if err != nil {
			return nil, err
		}

		slash := types.Slash{
			SpId:     sp.Id,
			ObjectId: msg.ObjectId,
			Height:   uint64(ctx.BlockHeight()),
		}
		k.SaveSlash(ctx, slash)
		k.SetSpSlashAmount(ctx, sp.Id, slashedAmount.Add(toSlashAmount))
	} else {
		// check whether it is a heartbeat attest
		heartbeatInterval := k.GetParams(ctx).HeartbeatInterval
		if msg.ChallengeId%heartbeatInterval != 0 {
			return nil, errors.Wrapf(types.ErrInvalidChallengeId, "heartbeat attestation should be submitted at interval %d", heartbeatInterval)
		}

		// reward validators & tx submitter
		err = k.doHeartbeatAndRewards(ctx, msg.ChallengeId, msg.VoteResult, sp.Id, submitter, challenger)
		if err != nil {
			return nil, err
		}
	}
	k.AppendAttestedChallenge(ctx, &types.AttestedChallenge{
		Id:     msg.ChallengeId,
		Result: msg.VoteResult,
	})

	return &types.MsgAttestResponse{}, nil
}

// calculateSlashAmount calculates the slash amount based on object size. There are also bounds of the amount.
func (k msgServer) calculateSlashAmount(ctx sdk.Context, objectSize uint64) sdkmath.Int {
	params := k.GetParams(ctx)
	sizeRate := params.SlashAmountSizeRate
	objectSizeInGB := sdk.NewDecFromBigInt(new(big.Int).SetUint64(objectSize)).QuoRoundUp(sdk.NewDec(1073741824))
	slashAmount := objectSizeInGB.MulMut(sizeRate).MulMut(sdk.NewDec(1e18)).TruncateInt()

	min := params.SlashAmountMin
	if slashAmount.LT(min) {
		return min
	}
	max := params.SlashAmountMax
	if slashAmount.GT(max) {
		return max
	}
	return slashAmount
}

// calculateSlashRewards calculates the rewards to challenger, submitter and validators when the total slash amount.
func (k msgServer) calculateSlashRewards(ctx sdk.Context, total sdkmath.Int, challenger sdk.AccAddress, validators int64) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	challengerReward := sdkmath.ZeroInt()
	var eachValidatorReward sdkmath.Int

	params := k.GetParams(ctx)
	threshold := params.RewardSubmitterThreshold
	submitterReward := params.RewardSubmitterRatio.Mul(sdk.NewDecFromInt(total)).TruncateInt()
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
		validatorRatio := params.RewardValidatorRatio
		eachValidatorReward = validatorRatio.MulInt(total).QuoInt64(validators).TruncateInt()
		for i := int64(0); i < validators; i++ {
			total = total.Sub(eachValidatorReward)
		}
		// the left is rewarded to challenger
		challengerReward = total
	}
	return challengerReward, eachValidatorReward, submitterReward
}

// doSlashAndRewards will execute the slash, transfer the rewards and emit events.
func (k msgServer) doSlashAndRewards(ctx sdk.Context, challengeId uint64, voteResult types.VoteResult, slashAmount sdkmath.Int,
	spID uint32, submitter, challenger sdk.AccAddress, validators []string) error {

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

	err := k.SpKeeper.Slash(ctx, spID, rewards)
	if err != nil {
		return err
	}

	event := types.EventAttestChallenge{
		ChallengeId:            challengeId,
		Result:                 voteResult,
		SpId:                   spID,
		SlashAmount:            slashAmount.String(),
		ChallengerAddress:      challenger.String(),
		ChallengerRewardAmount: challengerReward.String(),
		SubmitterAddress:       submitter.String(),
		SubmitterRewardAmount:  submitterReward.String(),
		ValidatorRewardAmount:  eachValidatorReward.MulRaw(int64(len(validators))).String(),
	}
	return ctx.EventManager().EmitTypedEvents(&event)
}

// calculateHeartbeatRewards calculates the rewards to all validators and submitter.
func (k msgServer) calculateHeartbeatRewards(ctx sdk.Context, total sdkmath.Int) (sdkmath.Int, sdkmath.Int) {
	params := k.GetParams(ctx)
	threshold := params.RewardSubmitterThreshold
	submitterReward := params.RewardSubmitterRatio.Mul(sdk.NewDecFromInt(total)).TruncateInt()
	if submitterReward.GT(threshold) {
		submitterReward = threshold
	}

	return total.Sub(submitterReward), submitterReward
}

// doHeartbeatAndRewards will transfer the tax to distribution account and rewards to submitter.
func (k msgServer) doHeartbeatAndRewards(ctx sdk.Context, challengeId uint64, voteResult types.VoteResult,
	spID uint32, submitter, challenger sdk.AccAddress) error {
	totalAmount, err := k.paymentKeeper.QueryDynamicBalance(ctx, paymentmoduletypes.ValidatorTaxPoolAddress)
	if err != nil {
		return err
	}

	validatorReward, submitterReward := sdkmath.NewInt(0), sdkmath.NewInt(0)
	if !totalAmount.IsZero() {
		validatorReward, submitterReward = k.calculateHeartbeatRewards(ctx, totalAmount)
		if validatorReward.IsPositive() {
			distModuleAcc := authtypes.NewModuleAddress(distributiontypes.ModuleName)
			err = k.paymentKeeper.Withdraw(ctx, paymentmoduletypes.ValidatorTaxPoolAddress, distModuleAcc, validatorReward)
			if err != nil {
				return err
			}
		}
		if submitterReward.IsPositive() {
			err = k.paymentKeeper.Withdraw(ctx, paymentmoduletypes.ValidatorTaxPoolAddress, submitter, submitterReward)
			if err != nil {
				return err
			}
		}
	}
	return ctx.EventManager().EmitTypedEvents(&types.EventAttestChallenge{
		ChallengeId:            challengeId,
		Result:                 voteResult,
		SpId:                   spID,
		SlashAmount:            "",
		ChallengerAddress:      challenger.String(),
		ChallengerRewardAmount: "",
		SubmitterAddress:       submitter.String(),
		SubmitterRewardAmount:  submitterReward.String(),
		ValidatorRewardAmount:  validatorReward.String(),
	})
}
