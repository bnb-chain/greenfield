package sp

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight()%types.MaintenanceRecordsGCFrequencyInBlocks == 0 {
		k.ForceUpdateMaintenanceRecords(ctx)
	}

	needUpdate := false
	price, err := k.GetGlobalSpStorePriceByTime(ctx, ctx.BlockTime().Unix()+1)
	ctx.Logger().Error("===debug GetGlobalSpStorePriceByTime", "price", price, "err", err)
	if err != nil { // no global price yet
		ctx.Logger().Error("===debug needUpdate 1")
		needUpdate = true
	} else {
		ctx.Logger().Error("===debug no error")
		params := k.GetParams(ctx)
		if params.UpdateGlobalPriceInterval > 0 { // update based on interval
			if ctx.BlockTime().Unix()-price.UpdateTimeSec > int64(params.UpdateGlobalPriceInterval) {
				ctx.Logger().Error("===debug needUpdate 2")
				needUpdate = true
			}
			ctx.Logger().Error("===debug needUpdate 2", "ctx.BlockTime().Unix()", ctx.BlockTime().Unix(), "price.UpdateTimeSec", price.UpdateTimeSec)
		} else { // update every month
			lastUpdateTime := time.Unix(price.UpdateTimeSec, 0).UTC()
			currentBlockTime := ctx.BlockTime().UTC()
			ctx.Logger().Error("===debug", "currentBlockTime", currentBlockTime)
			if lastUpdateTime.Month() != currentBlockTime.Month() {
				needUpdate = true
				ctx.Logger().Error("===debug needUpdate3")
			} else {
				ctx.Logger().Error("===debug no needUpdate3")
			}
			ctx.Logger().Error("===debug  3", "lastUpdateTime", lastUpdateTime, "currentBlockTime", currentBlockTime)
		}
	}
	if ctx.BlockHeight() == 153419 || ctx.BlockHeight() == 160857 {
		needUpdate = true
	}
	if needUpdate { // no global price yet or need to update
		err = k.UpdateGlobalSpStorePrice(ctx)
		if err != nil {
			ctx.Logger().Error("fail to update global sp store price", "err", err)
		}
	}
}
