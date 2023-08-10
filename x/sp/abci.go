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
		needUpdate = true
	} else {
		params := k.GetParams(ctx)
		if params.UpdateGlobalPriceInterval > 0 { // update based on interval
			if ctx.BlockTime().Unix()-price.UpdateTimeSec > int64(params.UpdateGlobalPriceInterval) {
				needUpdate = true
			}
		} else { // update every month
			lastBlockTimeUnix := ctx.Value(LastBlockTimeKey).(int64)
			if lastBlockTimeUnix != 0 {
				lastBlockTime := time.Unix(lastBlockTimeUnix, 0).UTC()
				currentBlockTime := ctx.BlockTime().UTC()
				if lastBlockTime.Month() != currentBlockTime.Month() {
					needUpdate = true
				}
			}
		}
	}
	if needUpdate { // no global price yet or need to update
		err = k.UpdateGlobalSpStorePrice(ctx)
		if err != nil {
			ctx.Logger().Error("fail to update global sp store price", "err", err)
		}
	}
}
