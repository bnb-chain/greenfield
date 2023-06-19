package keeper

import (
	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetOutFlow set a specific OutFlow in the store from its index
func (k Keeper) SetOutFlow(ctx sdk.Context, acc sdk.AccAddress, outFlow *types.OutFlow) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OutFlowKeyPrefix)
	bz, err := outFlow.Rate.Marshal()
	if err != nil {
		panic("should not happen")
	}
	store.Set(types.OutFlowKey(
		acc,
		outFlow.Status,
		sdk.MustAccAddressFromHex(outFlow.ToAddress),
	), bz)
}

// SetOutFlow set a specific OutFlow in the store from its index
func (k Keeper) GetOutFlows(ctx sdk.Context, acc sdk.AccAddress) []types.OutFlow {
	key := types.OutFlowKey(acc, types.OUT_FLOW_STATUS_ACTIVE, nil)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OutFlowKeyPrefix)
	iterator := store.Iterator(key, nil)
	defer iterator.Close()

	outFlows := make([]types.OutFlow, 0)
	for ; iterator.Valid(); iterator.Next() {
		_, outFlow := types.ParseOutFlowKey(iterator.Key())
		outFlow.Rate = types.ParseOutFlowValue(iterator.Value())
		outFlows = append(outFlows, outFlow)
	}
	return outFlows
}

// DeleteOutFlow set a specific OutFlow from the store
func (k Keeper) DeleteOutFlow(ctx sdk.Context, key []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OutFlowKeyPrefix)
	store.Delete(key)
}

// MergeActiveOutFlows merge active OutFlows and save in the store
func (k Keeper) MergeActiveOutFlows(ctx sdk.Context, acc sdk.AccAddress, outFlows []types.OutFlow) int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OutFlowKeyPrefix)
	deltaFlowCount := 0
	for _, outFlow := range outFlows {
		outFlow.Status = types.OUT_FLOW_STATUS_ACTIVE
		key := types.OutFlowKey(acc, outFlow.Status, sdk.MustAccAddressFromHex(outFlow.ToAddress))
		value := store.Get(key)
		if value == nil {
			k.SetOutFlow(ctx, acc, &outFlow)
			deltaFlowCount++
			continue
		}
		outFlow.Rate = types.ParseOutFlowValue(value).Add(outFlow.Rate)
		if outFlow.Rate.IsZero() {
			k.DeleteOutFlow(ctx, key)
			deltaFlowCount--
		} else {
			k.SetOutFlow(ctx, acc, &outFlow)
		}
	}
	return deltaFlowCount
}
