package keeper

import (
	"cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) UpdateToInMaintenance(ctx sdk.Context, sp *types.StorageProvider, requestDuration int64) error {
	params := k.GetParams(ctx)
	store := ctx.KVStore(k.storeKey)
	recordsPrefixStore := prefix.NewStore(store, types.GetStorageProviderMaintenanceRecordsPrefix(sdk.MustAccAddressFromHex(sp.OperatorAddress)))
	iterator := recordsPrefixStore.ReverseIterator(nil, nil)
	defer iterator.Close()
	newRecord := &types.MaintenanceRecord{
		Height:          ctx.BlockHeight(),
		RequestDuration: requestDuration,
		RequestAt:       ctx.BlockTime().Unix(),
	}
	sp.Status = types.STATUS_IN_MAINTENANCE
	if !iterator.Valid() {
		if requestDuration > params.GetMaintenanceDurationQuota() {
			return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "not enough quota, quota=%d is less than requested=%d", params.GetMaintenanceDurationQuota(), requestDuration)
		}
		store.Set(types.GetStorageProviderMaintenanceRecordKey(sdk.MustAccAddressFromHex(sp.OperatorAddress), ctx.BlockHeight()), k.cdc.MustMarshal(newRecord))
		return nil
	}

	totalUsedTime := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		record := &types.MaintenanceRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), record)
		threshHold := record.GetHeight() + params.GetNumOfLockupBlocksForMaintenance()
		if ctx.BlockHeight() < threshHold {
			return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "wait after block height %d", threshHold)
		}
		totalUsedTime = totalUsedTime + record.GetActualDuration()
	}
	if totalUsedTime+requestDuration > params.GetMaintenanceDurationQuota() {
		return errors.Wrapf(types.ErrStorageProviderStatusUpdateNotAllow, "not enough quota, quota=%d is less than requested=%d", params.GetMaintenanceDurationQuota()-totalUsedTime, requestDuration)
	}
	store.Set(types.GetStorageProviderMaintenanceRecordKey(sdk.MustAccAddressFromHex(sp.OperatorAddress), ctx.BlockHeight()), k.cdc.MustMarshal(newRecord))
	return nil
}

func (k Keeper) UpdateToInService(ctx sdk.Context, sp *types.StorageProvider) {
	store := ctx.KVStore(k.storeKey)
	spAddr := types.GetStorageProviderMaintenanceRecordsPrefix(sdk.MustAccAddressFromHex(sp.OperatorAddress))
	recordsPrefixStore := prefix.NewStore(store, spAddr)
	iterator := recordsPrefixStore.ReverseIterator(nil, nil)

	// update the latest record usedTime
	if iterator.Valid() {
		record := &types.MaintenanceRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), record)
		record.ActualDuration = ctx.BlockTime().Unix() - record.RequestAt
		recordsPrefixStore.Set(iterator.Key(), k.cdc.MustMarshal(record))
	}
	sp.Status = types.STATUS_IN_SERVICE
}

func (k Keeper) ForceMaintenanceRecords(ctx sdk.Context) {
	params := k.GetParams(ctx)
	store := ctx.KVStore(k.storeKey)
	curTime := ctx.BlockTime().Unix()
	iter := storetypes.KVStorePrefixIterator(store, types.StorageProviderKey)
	for ; iter.Valid(); iter.Next() {
		sp := types.MustUnmarshalStorageProvider(k.cdc, iter.Value())
		prefixStore := prefix.NewStore(store, types.GetStorageProviderMaintenanceRecordsPrefix(sdk.MustAccAddressFromHex(sp.OperatorAddress)))

		iterator := prefixStore.Iterator(nil, nil)
		for ; iterator.Valid(); iterator.Next() {
			record := &types.MaintenanceRecord{}
			k.cdc.MustUnmarshal(iterator.Value(), record)
			// purge old records
			if record.GetHeight()+params.GetNumOfHistoricalBlocksForMaintenanceRecords() < ctx.BlockHeight() {
				prefixStore.Delete(iterator.Key())
			} else if record.GetActualDuration() == 0 && record.RequestAt+record.GetRequestDuration() < curTime {
				record.ActualDuration = record.RequestDuration
				prefixStore.Set(iterator.Key(), k.cdc.MustMarshal(record))
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
