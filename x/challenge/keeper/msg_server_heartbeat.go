package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	app "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) Heartbeat(goCtx context.Context, msg *types.MsgHeartbeat) (*types.MsgHeartbeatResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	submitterAddress, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	// check challenge id
	heartbeatInterval := k.HeartbeatInterval(ctx)
	maxChallengeId := k.GetOngoingChallengeId(ctx)
	heartbeatChallengeId := k.GetHeartbeatChallengeId(ctx)

	if msg.ChallengeId > maxChallengeId || msg.ChallengeId <= heartbeatChallengeId ||
		msg.ChallengeId%heartbeatInterval != 0 { // be noted, we allow to skip some heartbeats
		return nil, types.ErrInvalidChallengeId
	}

	// check heartbeat validators and signatures
	_, err = k.verifySignature(ctx, msg)
	if err != nil {
		return nil, err
	}

	// reward tx validator & submitter
	total := k.paymentKeeper.QueryValidatorRewards(ctx)
	if !total.IsZero() {
		validatorReward, submitterReward := k.calculateHeartbeatRewards(ctx, total)
		if validatorReward.IsPositive() && submitterReward.IsPositive() {
			toValidator := sdk.Coins{
				sdk.Coin{Denom: app.Denom, Amount: validatorReward},
			}
			err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, paymentmoduletypes.ModuleName, distributiontypes.ModuleName, toValidator)
			if err != nil {
				return nil, err
			}
			toSubmitter := sdk.Coins{
				sdk.Coin{Denom: app.Denom, Amount: submitterReward},
			}
			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, paymentmoduletypes.ModuleName, submitterAddress, toSubmitter)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventChallengeHeartbeat{
		ChallengeId: msg.ChallengeId,
	}); err != nil {
		return nil, err
	}

	k.SetHeartbeatChallengeId(ctx, msg.ChallengeId)

	return &types.MsgHeartbeatResponse{}, nil
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
