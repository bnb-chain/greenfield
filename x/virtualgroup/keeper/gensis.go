package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	gvgStakingPool := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)

	if gvgStakingPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	// if account has zero balance it probably means it's not set, so we set it
	depositBalance := k.bankKeeper.GetAllBalances(ctx, gvgStakingPool.GetAddress())
	if depositBalance.IsZero() {
		k.accountKeeper.SetModuleAccount(ctx, gvgStakingPool)
	}
	depositAmount := sdk.ZeroInt()
	depositCoins := sdk.NewCoins(sdk.NewCoin(genState.Params.DepositDenom, depositAmount))

	if !depositBalance.IsEqual(depositCoins) {
		panic(fmt.Sprintf("sp deposit pool balance is different from sp deposit coins: %s <-> %s", depositBalance.String(), depositCoins.String()))
	}
}
