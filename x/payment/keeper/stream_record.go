package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetStreamRecord set a specific streamRecord in the store from its index
func (k Keeper) SetStreamRecord(ctx sdk.Context, streamRecord types.StreamRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StreamRecordKeyPrefix)
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
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StreamRecordKeyPrefix)

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
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StreamRecordKeyPrefix)
	store.Delete(types.StreamRecordKey(
		account,
	))
}

// GetAllStreamRecord returns all streamRecord
func (k Keeper) GetAllStreamRecord(ctx sdk.Context) (list []types.StreamRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StreamRecordKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.StreamRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) UpdateStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord, change *types.StreamRecordChange) error {
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
	// update lock balance
	if !change.LockBalanceChange.IsZero() {
		streamRecord.LockBalance = streamRecord.LockBalance.Add(change.LockBalanceChange)
		streamRecord.StaticBalance = streamRecord.StaticBalance.Sub(change.LockBalanceChange)
		if streamRecord.LockBalance.IsNegative() {
			return fmt.Errorf("lock balance can not become negative, current: %s", streamRecord.LockBalance)
		}
	}
	// update buffer balance
	if !change.RateChange.IsZero() {
		streamRecord.NetflowRate = streamRecord.NetflowRate.Add(change.RateChange)
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
	if !change.StaticBalanceChange.IsZero() {
		streamRecord.StaticBalance = streamRecord.StaticBalance.Add(change.StaticBalanceChange)
	}
	if streamRecord.StaticBalance.IsNegative() {
		account := sdk.MustAccAddressFromHex(streamRecord.Account)
		hasBankAccount := k.accountKeeper.HasAccount(ctx, account)
		if hasBankAccount {
			coins := sdk.NewCoins(sdk.NewCoin(types.Denom, streamRecord.StaticBalance.Abs()))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coins)
			if err != nil {
				ctx.Logger().Info("auto transfer failed", "account", streamRecord.Account, "err", err, "coins", coins)
			} else {
				streamRecord.StaticBalance = sdkmath.ZeroInt()
			}
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
	k.UpdateAutoSettleRecord(ctx, streamRecord.Account, streamRecord.SettleTimestamp, settleTimestamp)
	streamRecord.SettleTimestamp = settleTimestamp
	return nil
}

func (k Keeper) UpdateStreamRecordByAddr(ctx sdk.Context, change *types.StreamRecordChange) (ret *types.StreamRecord, err error) {
	streamRecord, found := k.GetStreamRecord(ctx, change.Addr)
	if !found {
		streamRecord = types.NewStreamRecord(change.Addr, ctx.BlockTime().Unix())
	}
	err = k.UpdateStreamRecord(ctx, &streamRecord, change)
	if err != nil {
		return
	}
	k.SetStreamRecord(ctx, streamRecord)
	return &streamRecord, nil
}

func (k Keeper) ForceSettle(ctx sdk.Context, streamRecord *types.StreamRecord) error {
	totalBalance := streamRecord.StaticBalance.Add(streamRecord.BufferBalance)
	change := types.NewDefaultStreamRecordChangeWithAddr(types.GovernanceAddress.String()).WithStaticBalanceChange(totalBalance)
	_, err := k.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return fmt.Errorf("update governance stream record failed: %w", err)
	}
	// force settle
	streamRecord.StaticBalance = sdkmath.ZeroInt()
	streamRecord.BufferBalance = sdkmath.ZeroInt()
	streamRecord.NetflowRate = sdkmath.ZeroInt()
	streamRecord.Status = types.StreamPaymentAccountStatusFrozen
	flows := k.FreezeFlowsByFromUser(ctx, streamRecord.Account)
	// todo: use a cache for SP stream record update to optimize
	// the implementation itself may cause chain force settle, but in reality, it will not happen.
	// only the SP can be the flow receiver, so in settlement, the rate of SP will reduce, but never get below zero and
	// trigger another force settle.
	for _, flow := range flows {
		change = types.NewDefaultStreamRecordChangeWithAddr(flow.To).WithRateChange(flow.Rate.Neg())
		_, err := k.UpdateStreamRecordByAddr(ctx, change)
		if err != nil {
			return fmt.Errorf("update receiver stream record failed: %w", err)
		}
	}
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

func (k Keeper) AutoSettle(ctx sdk.Context) {
	currentTimestamp := ctx.BlockTime().Unix()
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoSettleRecordKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	var num uint64 = 0
	maxNum := k.GetParams(ctx).MaxAutoForceSettleNum
	for ; iterator.Valid(); iterator.Next() {
		if num >= maxNum {
			return
		}
		var val types.AutoSettleRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.Timestamp > currentTimestamp {
			return
		}
		streamRecord, found := k.GetStreamRecord(ctx, val.Addr)
		if !found {
			ctx.Logger().Error("stream record not found", "addr", val.Addr)
			panic("stream record not found")
		}
		change := types.NewDefaultStreamRecordChangeWithAddr(val.Addr)
		err := k.UpdateStreamRecord(ctx, &streamRecord, change)
		if err != nil {
			ctx.Logger().Error("force settle failed", "addr", val.Addr, "err", err)
			panic("force settle failed")
		}
		num += 1
	}

}
