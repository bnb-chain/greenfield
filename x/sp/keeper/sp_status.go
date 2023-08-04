package keeper

import (
	"cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/x/sp/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) UpdateToInMaintenance(ctx sdk.Context, sp *types.StorageProvider, requestDuration int64) error {
	params := k.GetParams(ctx)
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
		if requestDuration > params.GetMaintenanceDurationQuota() {
			return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "not enough quota, quota=%d is less than requested=%d", params.GetMaintenanceDurationQuota(), requestDuration)
		}
		newStats := types.SpMaintenanceStats{Records: []*types.MaintenanceRecord{newRecord}}
		store.Set(key, k.cdc.MustMarshal(&newStats))
		return nil
	}
	var stats types.SpMaintenanceStats
	k.cdc.MustUnmarshal(bz, &stats)
	size := len(stats.Records)
	totalUsedTime := int64(0)
	// should not happen, stats with len 0 records will be deleted
	//if size == 0 {
	//	if requestDuration > params.GetMaintenanceDurationQuota() {
	//		return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "not enough quota, quota=%d is less than requested=%d", params.GetMaintenanceDurationQuota(), requestDuration)
	//	}
	//	stats.Records = append(stats.Records, newRecord)
	//	store.Set(key, k.cdc.MustMarshal(&stats))
	//	return nil
	//}

	for i := size - 1; i >= 0; i-- {
		record := stats.Records[i]
		if record == nil {
			continue // should not happen
		}
		threshHold := record.GetHeight() + params.GetNumOfLockupBlocksForMaintenance()
		if ctx.BlockHeight() < threshHold {
			return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "wait after block height %d", threshHold)
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
	// // update the latest record usedTime
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
			//// should not happen
			//if size == 0 {
			//	return
			//}
			for i := size - 1; i >= 0; i-- {
				if stats.Records[i] == nil {
					continue
				}
				// purge outdated records
				if stats.Records[i].GetHeight()+params.GetNumOfHistoricalBlocksForMaintenanceRecords() < ctx.BlockHeight() {
					stats.Records = append(stats.Records[:i], stats.Records[i+1:]...)
					if len(stats.Records) == 0 {
						store.Delete(key)
					} else {
						store.Set(key, k.cdc.MustMarshal(&stats))
					}
				} else if stats.Records[i].GetActualDuration() == 0 && stats.Records[i].RequestAt+stats.Records[i].GetRequestDuration() < curTime {
					// happen at most once
					stats.Records[i].ActualDuration = stats.Records[i].RequestDuration
					store.Set(key, k.cdc.MustMarshal(&stats))
					sp.Status = types.STATUS_IN_SERVICE
					k.SetStorageProvider(ctx, sp)
					k.SetStorageProviderByFundingAddr(ctx, sp)
					k.SetStorageProviderBySealAddr(ctx, sp)
					k.SetStorageProviderByApprovalAddr(ctx, sp)
					k.SetStorageProviderByGcAddr(ctx, sp)
					k.SetStorageProviderByTestAddr(ctx, sp)
					k.SetStorageProviderByBlsKey(ctx, sp)
				}
			}
		}
	}
}
