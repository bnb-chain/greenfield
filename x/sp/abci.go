package sp

import (
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight()%types.MaintenanceRecordsGCFrequencyInBlocks == 0 {
		k.ForceUpdateMaintenanceRecords(ctx)
	}
}
