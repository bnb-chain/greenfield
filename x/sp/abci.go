package sp

import (
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.ForceMaintenanceRecords(ctx)
}
