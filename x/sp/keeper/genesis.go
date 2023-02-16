package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	depositAmount := sdk.ZeroInt()
	for _, sp := range genState.StorageProviders {
		k.SetStorageProvider(ctx, sp)

		switch sp.GetStatus() {
		case types.STATUS_IN_SERVICE,
			types.STATUS_IN_JAILED:
			depositAmount = depositAmount.Add(sp.GetTotalDeposit())
		default:
			panic("invalid initialization storage provider status in genesis block")
		}
	}

	depositCoins := sdk.NewCoins(sdk.NewCoin(genState.Params.DepositDenom, depositAmount))

	spDepositPool := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if spDepositPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	depositBalance := k.bankKeeper.GetAllBalances(ctx, spDepositPool.GetAddress())
	if depositBalance.IsZero() {
		k.accountKeeper.SetModuleAccount(ctx, spDepositPool)
	}

	if !depositBalance.IsEqual(depositCoins) {
		panic(fmt.Sprintf("sp deposit pool balance is different from sp deposit coins: %s <-> %s", depositBalance.String(), depositCoins.String()))
	}
}
