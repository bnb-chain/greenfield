package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetStreamRecord set a specific streamRecord in the store from its index
func (k Keeper) SetStreamRecord(ctx sdk.Context, streamRecord types.StreamRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))
	b := k.cdc.MustMarshal(&streamRecord)
	store.Set(types.StreamRecordKey(
		streamRecord.Account,
	), b)
}

// GetStreamRecord returns a streamRecord from its index
func (k Keeper) GetStreamRecord(
	ctx sdk.Context,
	account string,

) (val types.StreamRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))

	b := store.Get(types.StreamRecordKey(
		account,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveStreamRecord removes a streamRecord from the store
func (k Keeper) RemoveStreamRecord(
	ctx sdk.Context,
	account string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))
	store.Delete(types.StreamRecordKey(
		account,
	))
}

// GetAllStreamRecord returns all streamRecord
func (k Keeper) GetAllStreamRecord(ctx sdk.Context) (list []types.StreamRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.StreamRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) UpdateStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord, rate, staticBalance sdkmath.Int, autoTransfer bool) error {
	currentTimestamp := ctx.BlockTime().Unix()
	timestamp := streamRecord.CrudTimestamp
	params := k.GetParams(ctx)
	// update delta balance
	if currentTimestamp != timestamp {
		if !streamRecord.NetflowRate.IsZero() {
			flowDelta := streamRecord.NetflowRate.MulRaw(currentTimestamp - timestamp)
			streamRecord.StaticBalance = streamRecord.StaticBalance.Add(flowDelta)
		}
		streamRecord.CrudTimestamp = currentTimestamp
	}
	// update buffer balance
	if !rate.IsZero() {
		streamRecord.NetflowRate = streamRecord.NetflowRate.Add(rate)
		newBufferBalance := sdkmath.ZeroInt()
		if streamRecord.NetflowRate.IsNegative() {
			newBufferBalance = streamRecord.NetflowRate.Abs().Mul(sdkmath.NewIntFromUint64(params.ReserveTime))
		}
		if !newBufferBalance.Equal(streamRecord.BufferBalance) {
			streamRecord.StaticBalance = streamRecord.StaticBalance.Sub(newBufferBalance).Add(streamRecord.BufferBalance)
			streamRecord.BufferBalance = newBufferBalance
		}
	}
	// update static balance
	if !staticBalance.IsZero() {
		streamRecord.StaticBalance = streamRecord.StaticBalance.Add(staticBalance)
	}
	if streamRecord.StaticBalance.IsNegative() {
		if autoTransfer {
			account := sdk.MustAccAddressFromHex(streamRecord.Account)
			coins := sdk.NewCoins(sdk.NewCoin(types.Denom, streamRecord.StaticBalance.Abs()))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coins)
			if err != nil {
				ctx.Logger().Info("auto transfer failed", "account", streamRecord.Account, "err", err, "coins", coins)
			} else {
				streamRecord.StaticBalance = sdkmath.ZeroInt()
			}
		}
		// calculate settle time
		var settleTimestamp int64 = 0
		if streamRecord.NetflowRate.IsNegative() {
			payDuration := streamRecord.StaticBalance.Add(streamRecord.BufferBalance).Quo(streamRecord.NetflowRate.Abs())
			if payDuration.LTE(sdkmath.NewIntFromUint64(params.ForcedSettleTime)) {
				err := k.ForceSettle(ctx, streamRecord)
				if err != nil {
					return fmt.Errorf("check and force settle failed, err: %w", err)
				}
			} else {
				settleTimestamp = currentTimestamp - int64(params.ForcedSettleTime) + payDuration.Int64()
			}
		}
		k.UpdateAutoSettleQueue(ctx, streamRecord.Account, streamRecord.SettleTimestamp, settleTimestamp)
		streamRecord.SettleTimestamp = settleTimestamp
	}
	return nil
}

func (k Keeper) UpdateStreamRecordByAddr(ctx sdk.Context, addr string, rate, staticBalanceDelta sdkmath.Int, autoTransfer bool) error {
	streamRecord, found := k.GetStreamRecord(ctx, addr)
	if !found {
		streamRecord = types.NewStreamRecord(addr, ctx.BlockTime().Unix())
	}
	err := k.UpdateStreamRecord(ctx, &streamRecord, rate, staticBalanceDelta, autoTransfer)
	if err != nil {
		return err
	}
	k.SetStreamRecord(ctx, streamRecord)
	return nil
}

func (k Keeper) ForceSettle(ctx sdk.Context, streamRecord *types.StreamRecord) error {
	totalBalance := streamRecord.StaticBalance.Add(streamRecord.BufferBalance)
	err := k.UpdateStreamRecordByAddr(ctx, types.PaymentModuleGovAddress.String(), sdkmath.ZeroInt(), totalBalance, false)
	if err != nil {
		return fmt.Errorf("update governance stream record failed: %w", err)
	}
	// force settle
	streamRecord.StaticBalance = sdkmath.ZeroInt()
	streamRecord.BufferBalance = sdkmath.ZeroInt()
	streamRecord.NetflowRate = sdkmath.ZeroInt()
	k.FreezeFlowsByFromUser(ctx, streamRecord.Account)
	// todo: update receivers' stream record of the flows
	// emit event
	err = ctx.EventManager().EmitTypedEvents(&types.EventForceSettle{
		Addr:           streamRecord.Account,
		SettledBalance: totalBalance,
	})
	if err != nil {
		return err
	}
	return nil
}
