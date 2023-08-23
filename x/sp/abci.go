package sp

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

// LastBlockTimeKey is the key to record last block's time, which will be set by app
const LastBlockTimeKey = "last_block_time"

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight()%types.MaintenanceRecordsGCFrequencyInBlocks == 0 {
		k.ForceUpdateMaintenanceRecords(ctx)
	}

	needUpdate := false
	price, err := k.GetGlobalSpStorePriceByTime(ctx, ctx.BlockTime().Unix()+1)
	if err != nil { // no global price yet
		ctx.Logger().Error("xxxx fail to GetGlobalSpStorePriceByTime", "err", err)
		needUpdate = true
	} else {
		params := k.GetParams(ctx)
		ctx.Logger().Error("xxxx GetParams", "UpdateGlobalPriceInterval", params.UpdateGlobalPriceInterval)
		if params.UpdateGlobalPriceInterval > 0 { // update based on interval
			if ctx.BlockTime().Unix()-price.UpdateTimeSec > int64(params.UpdateGlobalPriceInterval) {
				ctx.Logger().Error("xxxx needUpdate 1", "ctx.BlockTime().Unix()", ctx.BlockTime().Unix(), "price.UpdateTimeSec", price.UpdateTimeSec, "params.UpdateGlobalPriceInterval", params.UpdateGlobalPriceInterval)
				needUpdate = true
			} else {
				ctx.Logger().Error("xxxx no needUpdate 1")
			}

		} else { // update every month
			ctx.Logger().Error("xxxx every month")
			lastBlockTimeUnix := ctx.Value(LastBlockTimeKey).(int64)
			if lastBlockTimeUnix != 0 {
				lastBlockTime := time.Unix(lastBlockTimeUnix, 0).UTC()
				currentBlockTime := ctx.BlockTime().UTC()
				if lastBlockTime.Month() != currentBlockTime.Month() {
					ctx.Logger().Error("xxxx 2 needUpdate")
					needUpdate = true
				}
				ctx.Logger().Error("xxxx 2", "lastBlockTime.Month()", lastBlockTime.Month(), "currentBlockTime.Month()", currentBlockTime.Month())
			}
			ctx.Logger().Error("xxxx 2", "ctx.BlockTime().Unix()", ctx.BlockTime().Unix(), "lastBlockTimeUnix", lastBlockTimeUnix, "lastBlockTime", time.Unix(lastBlockTimeUnix, 0).UTC())
		}
	}
	if needUpdate { // no global price yet or need to update
		err = k.UpdateGlobalSpStorePrice(ctx)
		if err != nil {
			ctx.Logger().Error("fail to update global sp store price", "err", err)
		}
	}
	if ctx.BlockHeight() >= 153349 {
		panic("exit 153349")
	}
}
