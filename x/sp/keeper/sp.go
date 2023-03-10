package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (k Keeper) IsStorageProviderExistAndInService(ctx sdk.Context, addr sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetStorageProviderKey(addr))
	if value == nil {
		return types.ErrStorageProviderNotFound
	}

	sp := types.MustUnmarshalStorageProvider(k.cdc, value)
	if sp.Status != types.STATUS_IN_SERVICE {
		return types.ErrStorageProviderNotInService
	}
	return nil
}

func (k Keeper) GetStorageProvider(ctx sdk.Context, addr sdk.AccAddress) (sp types.StorageProvider, found bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetStorageProviderKey(addr))
	if value == nil {
		return sp, false
	}

	sp = types.MustUnmarshalStorageProvider(k.cdc, value)
	return sp, true
}

func (k Keeper) GetStorageProviderByFundingAddr(ctx sdk.Context, fundAddr sdk.AccAddress) (sp types.StorageProvider, found bool) {
	store := ctx.KVStore(k.storeKey)

	spAddr := store.Get(types.GetStorageProviderByFundingAddrKey(fundAddr))
	if spAddr == nil {
		return sp, false
	}

	return k.GetStorageProvider(ctx, spAddr)
}

func (k Keeper) GetStorageProviderBySealAddr(ctx sdk.Context, sealAddr sdk.AccAddress) (sp types.StorageProvider, found bool) {
	store := ctx.KVStore(k.storeKey)

	spAddr := store.Get(types.GetStorageProviderBySealAddrKey(sealAddr))
	if spAddr == nil {
		return sp, false
	}

	return k.GetStorageProvider(ctx, spAddr)
}

func (k Keeper) GetStorageProviderByApprovalAddr(ctx sdk.Context, approvalAddr sdk.AccAddress) (sp types.StorageProvider, found bool) {
	store := ctx.KVStore(k.storeKey)

	spAddr := store.Get(types.GetStorageProviderByApprovalAddrKey(approvalAddr))
	if spAddr == nil {
		return sp, false
	}

	return k.GetStorageProvider(ctx, spAddr)
}

func (k Keeper) SetStorageProvider(ctx sdk.Context, sp *types.StorageProvider) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalStorageProvider(k.cdc, sp)

	store.Set(types.GetStorageProviderKey(sp.GetOperator()), bz)
}

func (k Keeper) SetStorageProviderByFundingAddr(ctx sdk.Context, sp *types.StorageProvider) {
	fundAddr := sp.GetFundingAccAddress()
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetStorageProviderByFundingAddrKey(fundAddr), sp.GetOperator())
}

func (k Keeper) SetStorageProviderBySealAddr(ctx sdk.Context, sp *types.StorageProvider) {
	sealAddr := sp.GetSealAccAddress()
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetStorageProviderBySealAddrKey(sealAddr), sp.GetOperator())
}

func (k Keeper) SetStorageProviderByApprovalAddr(ctx sdk.Context, sp *types.StorageProvider) {
	approvalAddr := sp.GetApprovalAccAddress()
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetStorageProviderByApprovalAddrKey(approvalAddr), sp.GetOperator())
}

func (k Keeper) GetAllStorageProviders(ctx sdk.Context) (sps []types.StorageProvider) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.StorageProviderKey)

	for ; iter.Valid(); iter.Next() {
		sp := types.MustUnmarshalStorageProvider(k.cdc, iter.Value())
		sps = append(sps, sp)
	}
	return sps
}
