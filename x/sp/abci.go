package sp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight()%types.MaintenanceRecordsGCFrequencyInBlocks == 0 {
		k.ForceUpdateMaintenanceRecords(ctx)
	}

	blockTime := ctx.BlockTime().Unix()
	params := k.GetParams(ctx)
	price, err := k.GetGlobalSpStorePriceByTime(ctx, blockTime+1)
	if err != nil || blockTime-price.UpdateTimeSec > int64(params.UpdateGlobalPriceInterval) { // no global price yet or need to update
		err = k.UpdateGlobalSpStorePrice(ctx)
		if err != nil {
			ctx.Logger().Error("fail to update global sp store price", "err", err)
		}
		k.ClearSpUpdatePriceTimes(ctx)
	}
}
