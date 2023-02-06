package challenge

import (
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {}

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	// emit new challenge events

	// delete expired challenges
}
