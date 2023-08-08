package keeper

import (
	"cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/x/sp/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) UpdateToInMaintenance(ctx sdk.Context, sp *types.StorageProvider, requestDuration int64) error {
	params := k.GetParams(ctx)
	if requestDuration > params.GetMaintenanceDurationQuota() {
		return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "not enough quota, quota=%d is less than requested=%d", params.GetMaintenanceDurationQuota(), requestDuration)
	}
	store := ctx.KVStore(k.storeKey)
	newRecord := &types.MaintenanceRecord{
		Height:          ctx.BlockHeight(),
		RequestDuration: requestDuration,
		RequestAt:       ctx.BlockTime().Unix(),
	}
	key := types.GetStorageProviderMaintenanceRecordsKey(sdk.MustAccAddressFromHex(sp.OperatorAddress))
	bz := store.Get(key)
	sp.Status = types.STATUS_IN_MAINTENANCE
	if bz == nil {
		newStats := types.SpMaintenanceStats{Records: []*types.MaintenanceRecord{newRecord}}
		store.Set(key, k.cdc.MustMarshal(&newStats))
		return nil
	}
	var stats types.SpMaintenanceStats
	k.cdc.MustUnmarshal(bz, &stats)
	size := len(stats.Records)
	totalUsedTime := int64(0)
	for i := size - 1; i >= 0; i-- {
		record := stats.Records[i]
		if i == size-1 {
			threshHold := record.GetHeight() + params.GetNumOfLockupBlocksForMaintenance()
			if ctx.BlockHeight() < threshHold {
				return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "wait after block height %d", threshHold)
			}
		}
		totalUsedTime = totalUsedTime + record.GetActualDuration()
	}
	if totalUsedTime+requestDuration > params.GetMaintenanceDurationQuota() {
		return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "not enough quota, quota=%d is less than requested=%d", params.GetMaintenanceDurationQuota()-totalUsedTime, requestDuration)
	}
	stats.Records = append(stats.Records, newRecord)
	store.Set(key, k.cdc.MustMarshal(&stats))
	return nil
}

func (k Keeper) UpdateToInService(ctx sdk.Context, sp *types.StorageProvider) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetStorageProviderMaintenanceRecordsKey(sdk.MustAccAddressFromHex(sp.OperatorAddress))
	bz := store.Get(key)
	if bz != nil {
		var stats types.SpMaintenanceStats
		k.cdc.MustUnmarshal(bz, &stats)
		size := len(stats.Records)
		if size != 0 {
			lastRecord := stats.Records[size-1]
			lastRecord.ActualDuration = ctx.BlockTime().Unix() - lastRecord.RequestAt
			store.Set(key, k.cdc.MustMarshal(&stats))
		}
	}
	sp.Status = types.STATUS_IN_SERVICE
}

func (k Keeper) ForceUpdateMaintenanceRecords(ctx sdk.Context) {
	params := k.GetParams(ctx)
	store := ctx.KVStore(k.storeKey)
	curTime := ctx.BlockTime().Unix()
	iter := storetypes.KVStorePrefixIterator(store, types.StorageProviderKey)
	for ; iter.Valid(); iter.Next() {
		sp := types.MustUnmarshalStorageProvider(k.cdc, iter.Value())
		key := types.GetStorageProviderMaintenanceRecordsKey(sdk.MustAccAddressFromHex(sp.OperatorAddress))
		bz := store.Get(key)
		if bz != nil {
			var stats types.SpMaintenanceStats
			k.cdc.MustUnmarshal(bz, &stats)
			size := len(stats.Records)
			// force update any maintenance record that not been updated back to in_service after requested duration.
			changed := false
			if sp.Status != types.STATUS_IN_SERVICE {
				for i := size - 1; i >= 0; i-- {
					if stats.Records[i].GetActualDuration() == 0 && stats.Records[i].RequestAt+stats.Records[i].GetRequestDuration() < curTime {
						stats.Records[i].ActualDuration = stats.Records[i].RequestDuration
						store.Set(key, k.cdc.MustMarshal(&stats))
						sp.Status = types.STATUS_IN_SERVICE
						k.SetStorageProvider(ctx, sp)
						changed = true
					}
				}
			}
			// purge outdated records
			for i := size - 1; i >= 0; i-- {
				if stats.Records[i].GetHeight()+params.GetNumOfHistoricalBlocksForMaintenanceRecords() < ctx.BlockHeight() {
					stats.Records = stats.Records[i+1:]
					changed = true
					break
				}
			}
			if len(stats.Records) == 0 {
				store.Delete(key)
			} else if changed {
				store.Set(key, k.cdc.MustMarshal(&stats))
			}
		}
	}
}
