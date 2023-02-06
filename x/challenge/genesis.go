package challenge

import (
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the ongoingChallenge
	for _, elem := range genState.OngoingChallenges {
		k.SetOngoingChallenge(ctx, elem)
	}
	// Set all the recentSlash
	for _, elem := range genState.RecentSlashes {
		k.SetRecentSlash(ctx, elem)
	}

	// Set recentSlash count
	k.SetRecentSlashCount(ctx, genState.RecentSlashCount)
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.OngoingChallenges = k.GetAllOngoingChallenge(ctx)
	genesis.RecentSlashes = k.GetAllRecentSlash(ctx)
	genesis.RecentSlashCount = k.GetRecentSlashCount(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
