package keeper

import (
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetFlow set a specific flow in the store from its index
func (k Keeper) SetFlow(ctx sdk.Context, flow types.Flow) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FlowKeyPrefix))
	b := k.cdc.MustMarshal(&flow)
	store.Set(types.FlowKey(
		flow.From,
		flow.To,
	), b)
}

// GetFlow returns a flow from its index
func (k Keeper) GetFlow(
	ctx sdk.Context,
	from string,
	to string,

) (val types.Flow, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FlowKeyPrefix))

	b := store.Get(types.FlowKey(
		from,
		to,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveFlow removes a flow from the store
func (k Keeper) RemoveFlow(
	ctx sdk.Context,
	from string,
	to string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FlowKeyPrefix))
	store.Delete(types.FlowKey(
		from,
		to,
	))
}

// GetAllFlow returns all flow
func (k Keeper) GetAllFlow(ctx sdk.Context) (list []types.Flow) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FlowKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Flow
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) GetAllFlowByFromUser(ctx sdk.Context, from string) (list []types.Flow) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FlowKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte(from))

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Flow
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// merge the incoming flow with the existing flow
func (k Keeper) UpdateFlow(ctx sdk.Context, flow types.Flow) error {
	existingFlow, found := k.GetFlow(ctx, flow.From, flow.To)
	if found {
		existingFlow.Rate = flow.Rate.Add(existingFlow.Rate)
	} else {
		existingFlow = flow
	}
	if existingFlow.Rate.IsNegative() {
		return fmt.Errorf("flow rate cannot be negative")
	}
	k.SetFlow(ctx, existingFlow)
	return nil
}

func (k Keeper) FreezeFlowsByFromUser(ctx sdk.Context, from string) []types.Flow {
	flows := k.GetAllFlowByFromUser(ctx, from)
	for _, flow := range flows {
		flow.Frozen = true
		k.SetFlow(ctx, flow)
	}
	return flows
}

func (k Keeper) UnfreezeFlowsByFromUser(ctx sdk.Context, from string) []types.Flow {
	flows := k.GetAllFlowByFromUser(ctx, from)
	for _, flow := range flows {
		flow.Frozen = false
		k.SetFlow(ctx, flow)
	}
	return flows
}
