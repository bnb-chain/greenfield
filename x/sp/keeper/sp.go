package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (k Keeper) GetStorageProvider(ctx sdk.Context, addr sdk.AccAddress) (sp types.StorageProvider, found bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetStorageProviderKey(addr))
	if value == nil {
		return sp, false
	}

	sp = types.MustUnmarshalStorageProvider(k.cdc, value)
	return sp, true
}

func (k Keeper) SetStorageProvider(ctx sdk.Context, sp types.StorageProvider) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalStorageProvider(k.cdc, &sp)
	store.Set(types.GetStorageProviderKey(sp.GetOperator()), bz)
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
