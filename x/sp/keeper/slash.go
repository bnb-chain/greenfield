package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (k Keeper) Slash(ctx sdk.Context, spID uint32, rewardInfos []types.RewardInfo) error {
	sp, found := k.GetStorageProvider(ctx, spID)
	if !found {
		return types.ErrStorageProviderNotFound
	}

	totalAmount := sdkmath.NewInt(0)
	for _, rewardInfo := range rewardInfos {
		totalAmount = totalAmount.Add(rewardInfo.Amount.Amount)
		if k.DepositDenomForSP(ctx) != rewardInfo.GetAmount().Denom {
			return types.ErrInvalidDenom.Wrapf("Expect: %s, actual: %s", k.DepositDenomForSP(ctx), rewardInfo.GetAmount().Denom)
		}
	}

	if totalAmount.GT(sp.TotalDeposit) {
		return types.ErrInsufficientDepositAmount
	}

	for _, rewardInfo := range rewardInfos {
		rewardAcc, err := sdk.AccAddressFromHexUnsafe(rewardInfo.Address)
		if err != nil {
			return err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, rewardAcc, sdk.NewCoins(rewardInfo.GetAmount()))
		if err != nil {
			return err
		}
	}

	sp.TotalDeposit = sp.TotalDeposit.Sub(totalAmount)
	k.SetStorageProvider(ctx, sp)

	return nil
}
