package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (k Keeper) Slash(ctx sdk.Context, spAcc sdk.AccAddress, rewardInfos []types.RewardInfo) error {
	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		return types.ErrStorageProviderNotFound
	}

	totalAmount := sdkmath.NewInt(0)
	for _, rewardInfo := range rewardInfos {
		totalAmount.Add(rewardInfo.Amount.Amount)
	}

	if totalAmount.GT(sp.TotalDeposit) {
		return types.ErrInsufficientDepositAmount
	}

	for _, rewardInfo := range rewardInfos {
		rewardAcc, err := sdk.AccAddressFromHexUnsafe(rewardInfo.Address)
		if err != nil {
			return err
		}

		coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForSP(ctx), rewardInfo.GetAmount().Amount))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, rewardAcc, coins); err != nil {
			// TODO: need consider rollback
			return err
		}
	}

	// TODO: if the total deposit of SP is less than the MinDeposit, we will jail it.
	return nil

}
