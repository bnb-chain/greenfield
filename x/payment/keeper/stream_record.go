package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) CheckStreamRecord(streamRecord *types.StreamRecord) {
	if streamRecord == nil {
		panic("streamRecord is nil")
	}
	if len(streamRecord.Account) != sdk.EthAddressLength*2+2 {
		panic(fmt.Sprintf("invalid streamRecord account %s", streamRecord.Account))
	}
	if streamRecord.Status != types.STREAM_ACCOUNT_STATUS_ACTIVE && streamRecord.Status != types.STREAM_ACCOUNT_STATUS_FROZEN {
		panic(fmt.Sprintf("invalid streamRecord status %d", streamRecord.Status))
	}
	if streamRecord.StaticBalance.IsNil() {
		panic(fmt.Sprintf("invalid streamRecord staticBalance %s", streamRecord.StaticBalance))
	}
	if streamRecord.NetflowRate.IsNil() {
		panic(fmt.Sprintf("invalid streamRecord netflowRate %s", streamRecord.NetflowRate))
	}
	if !streamRecord.FrozenNetflowRate.IsNil() && streamRecord.FrozenNetflowRate.IsPositive() {
		panic(fmt.Sprintf("invalid streamRecord frozenNetflowRate %s", streamRecord.NetflowRate))
	}
	if streamRecord.LockBalance.IsNil() || streamRecord.LockBalance.IsNegative() {
		panic(fmt.Sprintf("invalid streamRecord lockBalance %s", streamRecord.LockBalance))
	}
	if streamRecord.BufferBalance.IsNil() || streamRecord.BufferBalance.IsNegative() {
		panic(fmt.Sprintf("invalid streamRecord bufferBalance %s", streamRecord.BufferBalance))
	}
}

// SetStreamRecord set a specific streamRecord in the store from its index
func (k Keeper) SetStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord) {
	k.CheckStreamRecord(streamRecord)
	account := streamRecord.Account
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StreamRecordKeyPrefix)
	key := types.StreamRecordKey(sdk.MustAccAddressFromHex(account))
	streamRecord.Account = ""
	b := k.cdc.MustMarshal(streamRecord)
	store.Set(key, b)
	// set the field back, the streamRecord may be used after this function
	streamRecord.Account = account
	event := &types.EventStreamRecordUpdate{
		Account:           streamRecord.Account,
		StaticBalance:     streamRecord.StaticBalance,
		NetflowRate:       streamRecord.NetflowRate,
		FrozenNetflowRate: streamRecord.FrozenNetflowRate,
		CrudTimestamp:     streamRecord.CrudTimestamp,
		Status:            streamRecord.Status,
		LockBalance:       streamRecord.LockBalance,
		BufferBalance:     streamRecord.BufferBalance,
		SettleTimestamp:   streamRecord.SettleTimestamp,
	}
	_ = ctx.EventManager().EmitTypedEvents(event)
}

// GetStreamRecord returns a streamRecord from its index
func (k Keeper) GetStreamRecord(
	ctx sdk.Context,
	account sdk.AccAddress,
) (val *types.StreamRecord, found bool) {
	val = types.NewStreamRecord(account, ctx.BlockTime().Unix())
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StreamRecordKeyPrefix)

	b := store.Get(types.StreamRecordKey(
		account,
	))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, val)
	val.Account = account.String()
	return val, true
}

func (k Keeper) IsEmptyNetFlow(ctx sdk.Context, account sdk.AccAddress) bool {
	record, found := k.GetStreamRecord(ctx, account)
	if !found {
		return true // treat as empty, for internal use only
	}
	return record.NetflowRate.IsZero() && record.FrozenNetflowRate.IsZero()
}

// GetAllStreamRecord returns all streamRecord
func (k Keeper) GetAllStreamRecord(ctx sdk.Context) (list []types.StreamRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StreamRecordKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.StreamRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		val.Account = string(iterator.Key())
		list = append(list, val)
	}

	return
}

// UpdateFrozenStreamRecord updates frozen streamRecord in `force delete` scenarios
// it only handles the lock balance change and ignore the other changes(since the streams are already changed and the
// accumulated OutFlows are changed outside this function)
func (k Keeper) UpdateFrozenStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord, change *types.StreamRecordChange) error {
	currentTimestamp := ctx.BlockTime().Unix()
	streamRecord.CrudTimestamp = currentTimestamp
	// update lock balance
	if !change.LockBalanceChange.IsZero() {
		streamRecord.LockBalance = streamRecord.LockBalance.Add(change.LockBalanceChange)
		streamRecord.StaticBalance = streamRecord.StaticBalance.Sub(change.LockBalanceChange)
		if streamRecord.LockBalance.IsNegative() {
			return fmt.Errorf("lock balance can not become negative, current: %s", streamRecord.LockBalance)
		}
	}
	return nil
}

func (k Keeper) UpdateStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord, change *types.StreamRecordChange) error {
	forced, _ := ctx.Value(types.ForceUpdateStreamRecordKey).(bool) // force update in end block
	if streamRecord.Status != types.STREAM_ACCOUNT_STATUS_ACTIVE {
		if forced { //stream record is frozen
			return k.UpdateFrozenStreamRecord(ctx, streamRecord, change)
		}
		return fmt.Errorf("stream account %s is frozen", streamRecord.Account)
	}

	isPay := change.StaticBalanceChange.IsNegative() || change.RateChange.IsNegative()
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
			newBufferBalance = streamRecord.NetflowRate.Abs().Mul(sdkmath.NewIntFromUint64(params.VersionedParams.ReserveTime))
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
			coins := sdk.NewCoins(sdk.NewCoin(params.FeeDenom, streamRecord.StaticBalance.Abs()))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coins)
			if err != nil {
				ctx.Logger().Info("auto transfer failed", "account", streamRecord.Account, "err", err, "coins", coins)
			} else {
				streamRecord.StaticBalance = sdkmath.ZeroInt()
			}
		}
	}
	// if the change is a pay (which decreases the static balance or netflow rate), the left static balance should be enough
	if !forced && isPay && streamRecord.StaticBalance.IsNegative() {
		return fmt.Errorf("stream account %s balance not enough, lack of %s BNB", streamRecord.Account, streamRecord.StaticBalance.Abs())
	}
	//calculate settle time
	var settleTimestamp int64 = 0
	if streamRecord.NetflowRate.IsNegative() {
		payDuration := streamRecord.StaticBalance.Add(streamRecord.BufferBalance).Quo(streamRecord.NetflowRate.Abs())
		if payDuration.LTE(sdkmath.NewIntFromUint64(params.ForcedSettleTime)) {
			if !forced {
				return fmt.Errorf("stream account %s balance not enough, lack of %s BNB", streamRecord.Account, streamRecord.StaticBalance.Abs())
			}
		}
		settleTimestamp = currentTimestamp - int64(params.ForcedSettleTime) + payDuration.Int64()
	}
	k.UpdateAutoSettleRecord(ctx, sdk.MustAccAddressFromHex(streamRecord.Account), streamRecord.SettleTimestamp, settleTimestamp)
	streamRecord.SettleTimestamp = settleTimestamp
	return nil
}

func (k Keeper) SettleStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord) error {
	currentTimestamp := ctx.BlockTime().Unix()
	crudTimestamp := streamRecord.CrudTimestamp
	params := k.GetParams(ctx)

	if currentTimestamp != crudTimestamp {
		if !streamRecord.NetflowRate.IsZero() {
			flowDelta := streamRecord.NetflowRate.MulRaw(currentTimestamp - crudTimestamp)
			streamRecord.StaticBalance = streamRecord.StaticBalance.Add(flowDelta)
		}
		streamRecord.CrudTimestamp = currentTimestamp
	}

	if streamRecord.StaticBalance.IsNegative() {
		account := sdk.MustAccAddressFromHex(streamRecord.Account)
		hasBankAccount := k.accountKeeper.HasAccount(ctx, account)
		if hasBankAccount {
			coins := sdk.NewCoins(sdk.NewCoin(params.FeeDenom, streamRecord.StaticBalance.Abs()))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coins)
			if err != nil {
				ctx.Logger().Info("auto transfer failed when settling", "account", streamRecord.Account, "err", err, "coins", coins)
			} else {
				streamRecord.StaticBalance = sdkmath.ZeroInt()
			}
		}
	}

	if streamRecord.NetflowRate.IsNegative() {
		payDuration := streamRecord.StaticBalance.Add(streamRecord.BufferBalance).Quo(streamRecord.NetflowRate.Abs())
		if payDuration.LTE(sdkmath.NewIntFromUint64(params.ForcedSettleTime)) {
			err := k.ForceSettle(ctx, streamRecord)
			if err != nil {
				ctx.Logger().Error("fail to force settle stream record", "err", err, "record", streamRecord.Account)
				return err
			}
		} else {
			settleTimestamp := currentTimestamp - int64(params.ForcedSettleTime) + payDuration.Int64()
			k.UpdateAutoSettleRecord(ctx, sdk.MustAccAddressFromHex(streamRecord.Account), streamRecord.SettleTimestamp, settleTimestamp)
			streamRecord.SettleTimestamp = settleTimestamp
		}
	}

	return nil
}

func (k Keeper) UpdateStreamRecordByAddr(ctx sdk.Context, change *types.StreamRecordChange) (ret *types.StreamRecord, err error) {
	streamRecord, _ := k.GetStreamRecord(ctx, change.Addr)
	err = k.UpdateStreamRecord(ctx, streamRecord, change)
	if err != nil {
		return
	}
	k.SetStreamRecord(ctx, streamRecord)
	return streamRecord, nil
}

func (k Keeper) ForceSettle(ctx sdk.Context, streamRecord *types.StreamRecord) error {
	totalBalance := streamRecord.StaticBalance.Add(streamRecord.BufferBalance)
	if totalBalance.IsPositive() {
		change := types.NewDefaultStreamRecordChangeWithAddr(types.GovernanceAddress).WithStaticBalanceChange(totalBalance)
		_, err := k.UpdateStreamRecordByAddr(ctx, change)
		if err != nil {
			return fmt.Errorf("update governance stream record failed: %w", err)
		}
	}
	// force settle
	streamRecord.StaticBalance = sdkmath.ZeroInt()
	streamRecord.BufferBalance = sdkmath.ZeroInt()
	streamRecord.Status = types.STREAM_ACCOUNT_STATUS_FROZEN
	// emit event
	_ = ctx.EventManager().EmitTypedEvents(&types.EventForceSettle{
		Addr:           streamRecord.Account,
		SettledBalance: totalBalance,
	})
	return nil
}

func (k Keeper) AutoSettle(ctx sdk.Context) {
	currentTimestamp := ctx.BlockTime().Unix()
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoSettleRecordKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	count := uint64(0)
	max := k.GetParams(ctx).MaxAutoSettleFlowCount
	for ; iterator.Valid(); iterator.Next() {
		if count >= max {
			return
		}
		record := types.ParseAutoSettleRecordKey(iterator.Key())
		addr := sdk.MustAccAddressFromHex(record.Addr)
		if record.Timestamp > currentTimestamp {
			return
		}
		streamRecord, found := k.GetStreamRecord(ctx, addr)
		if !found { // should not happen
			ctx.Logger().Error("auto settle, stream record not found", "address", record.Addr)
			continue
		}

		if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE {
			count++ // add one for a stream record
			err := k.SettleStreamRecord(ctx, streamRecord)
			if err != nil {
				ctx.Logger().Error("auto settle, settle stream record failed", "err", err.Error())
				continue
			}
			k.SetStreamRecord(ctx, streamRecord)
			if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE {
				continue
			}
			if count >= max {
				return
			}
		}

		if k.ExistsAutoResumeRecord(ctx, record.Timestamp, addr) { // this check should be cheap usually
			continue //skip the one if the stream account is in resuming
		}

		activeFlowKey := types.OutFlowKey(addr, types.OUT_FLOW_STATUS_ACTIVE, nil)
		flowStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OutFlowKeyPrefix)
		flowIterator := flowStore.Iterator(activeFlowKey, nil)
		defer flowIterator.Close()

		finished := false
		totalRate := sdk.ZeroInt()
		toUpdate := make([]types.OutFlow, 0)
		for ; flowIterator.Valid(); flowIterator.Next() {
			if count >= max {
				break
			}
			addrInKey, outFlow := types.ParseOutFlowKey(flowIterator.Key())
			if !addrInKey.Equals(addr) {
				finished = true
				break
			}
			if outFlow.Status == types.OUT_FLOW_STATUS_FROZEN {
				finished = true
				break
			}
			outFlow.Rate = types.ParseOutFlowValue(flowIterator.Value())
			ctx.Logger().Debug("auto settling record", "height", ctx.BlockHeight(),
				"address", addr.String(),
				"to address", outFlow.ToAddress,
				"rate", outFlow.Rate.String())

			toAddr := sdk.MustAccAddressFromHex(outFlow.ToAddress)
			flowChange := types.NewDefaultStreamRecordChangeWithAddr(toAddr).WithRateChange(outFlow.Rate.Neg())
			_, err := k.UpdateStreamRecordByAddr(ctx, flowChange)
			if err != nil {
				ctx.Logger().Error("auto settle, update stream record failed", "address", outFlow.ToAddress, "rate", outFlow.Rate.Neg())
				panic("should not happen")
			}

			flowStore.Delete(flowIterator.Key())

			outFlow.Status = types.OUT_FLOW_STATUS_FROZEN
			toUpdate = append(toUpdate, outFlow)

			totalRate = totalRate.Add(outFlow.Rate)
			count++
		}

		for _, outFlow := range toUpdate {
			k.SetOutFlow(ctx, addr, &outFlow)
		}

		streamRecord.NetflowRate = streamRecord.NetflowRate.Add(totalRate)
		streamRecord.FrozenNetflowRate = streamRecord.FrozenNetflowRate.Add(totalRate.Neg())

		if !flowIterator.Valid() || finished {
			if !streamRecord.NetflowRate.IsZero() {
				ctx.Logger().Error("should not happen, stream netflow rate is not zero", "address", streamRecord.Account)
				panic("should not happen")
			}
			k.RemoveAutoSettleRecord(ctx, record.Timestamp, addr)
		}

		k.SetStreamRecord(ctx, streamRecord)
	}
}

func (k Keeper) TryResumeStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord, depositBalance sdkmath.Int) error {
	if streamRecord.Status != types.STREAM_ACCOUNT_STATUS_FROZEN {
		return fmt.Errorf("stream account %s status is not frozen", streamRecord.Account)
	}

	if !streamRecord.NetflowRate.IsZero() { // the account is resuming or settling
		return fmt.Errorf("stream account %s status is resuming or settling, please wait", streamRecord.Account)
	}

	params := k.GetParams(ctx)
	reserveTime := params.VersionedParams.ReserveTime
	forcedSettleTime := params.ForcedSettleTime

	totalRate := streamRecord.NetflowRate.Add(streamRecord.FrozenNetflowRate)
	streamRecord.StaticBalance = streamRecord.StaticBalance.Add(depositBalance)
	expectedBalanceToResume := totalRate.Neg().Mul(sdkmath.NewIntFromUint64(reserveTime))
	if streamRecord.StaticBalance.LT(expectedBalanceToResume) {
		// deposit balance is not enough to resume, only add static balance
		k.SetStreamRecord(ctx, streamRecord)
		return nil
	}

	now := ctx.BlockTime().Unix()
	streamRecord.SettleTimestamp = now + streamRecord.StaticBalance.Quo(totalRate).Int64() - int64(forcedSettleTime)
	streamRecord.BufferBalance = expectedBalanceToResume
	streamRecord.StaticBalance = streamRecord.StaticBalance.Sub(expectedBalanceToResume)
	streamRecord.CrudTimestamp = now

	ctx.Logger().Debug("try to resume stream account", "streamRecord.OutFlowCount", streamRecord.OutFlowCount, "params.MaxAutoResumeFlowCount", params.MaxAutoResumeFlowCount)
	if streamRecord.OutFlowCount <= params.MaxAutoResumeFlowCount { //only rough judgement, resume directly
		streamRecord.Status = types.STREAM_ACCOUNT_STATUS_ACTIVE
		streamRecord.NetflowRate = totalRate
		streamRecord.FrozenNetflowRate = sdkmath.ZeroInt()

		addr := sdk.MustAccAddressFromHex(streamRecord.Account)
		frozenFlowKey := types.OutFlowKey(addr, types.OUT_FLOW_STATUS_FROZEN, nil)
		flowStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OutFlowKeyPrefix)
		flowIterator := flowStore.Iterator(frozenFlowKey, nil)
		defer flowIterator.Close()

		toUpdate := make([]types.OutFlow, 0)
		for ; flowIterator.Valid(); flowIterator.Next() {
			addrInKey, outFlow := types.ParseOutFlowKey(flowIterator.Key())
			if !addrInKey.Equals(addr) {
				break
			}

			outFlow.Rate = types.ParseOutFlowValue(flowIterator.Value())

			toAddr := sdk.MustAccAddressFromHex(outFlow.ToAddress)
			change := types.NewDefaultStreamRecordChangeWithAddr(toAddr).WithRateChange(outFlow.Rate)
			_, err := k.UpdateStreamRecordByAddr(ctx, change)
			if err != nil {
				return fmt.Errorf("try resume, update receiver stream record failed: %w", err)
			}

			flowStore.Delete(flowIterator.Key())

			outFlow.Status = types.OUT_FLOW_STATUS_ACTIVE
			toUpdate = append(toUpdate, outFlow)
		}
		for _, outFlow := range toUpdate {
			k.SetOutFlow(ctx, addr, &outFlow)
		}

		k.SetStreamRecord(ctx, streamRecord)
		k.UpdateAutoSettleRecord(ctx, sdk.MustAccAddressFromHex(streamRecord.Account), 0, streamRecord.SettleTimestamp)
		return nil
	} else { //enqueue for resume in end block
		k.SetStreamRecord(ctx, streamRecord)
		k.SetAutoResumeRecord(ctx, &types.AutoResumeRecord{
			Timestamp: now,
			Addr:      streamRecord.Account,
		})
		return nil
	}
}

func (k Keeper) AutoResume(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoResumeRecordKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	var count uint64 = 0
	max := k.GetParams(ctx).MaxAutoResumeFlowCount
	for ; iterator.Valid(); iterator.Next() {
		record := types.ParseAutoResumeRecordKey(iterator.Key())
		addr := sdk.MustAccAddressFromHex(record.Addr)

		streamRecord, found := k.GetStreamRecord(ctx, addr)
		if !found { // should not happen
			ctx.Logger().Error("auto resume, stream record not found", "address", record.Addr)
			continue
		}

		totalRate := sdk.ZeroInt()
		frozenFlowKey := types.OutFlowKey(addr, types.OUT_FLOW_STATUS_FROZEN, nil)
		flowStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OutFlowKeyPrefix)
		flowIterator := flowStore.Iterator(frozenFlowKey, nil)
		defer flowIterator.Close()

		finished := false
		toUpdate := make([]types.OutFlow, 0)
		for ; flowIterator.Valid(); flowIterator.Next() {
			if count >= max {
				break
			}
			addrInKey, outFlow := types.ParseOutFlowKey(flowIterator.Key())
			if !addrInKey.Equals(addr) {
				finished = true
				break
			}

			outFlow.Rate = types.ParseOutFlowValue(flowIterator.Value())
			ctx.Logger().Debug("auto resuming record", "height", ctx.BlockHeight(),
				"address", addr.String(),
				"to address", outFlow.ToAddress,
				"rate", outFlow.Rate.String())

			toAddr := sdk.MustAccAddressFromHex(outFlow.ToAddress)
			flowChange := types.NewDefaultStreamRecordChangeWithAddr(toAddr).WithRateChange(outFlow.Rate)
			_, err := k.UpdateStreamRecordByAddr(ctx, flowChange)
			if err != nil {
				ctx.Logger().Error("auto resume, update receiver stream record failed", "address", outFlow.ToAddress, "err", err.Error())
				panic("should not happen")
			}

			flowStore.Delete(flowIterator.Key())

			outFlow.Status = types.OUT_FLOW_STATUS_ACTIVE
			toUpdate = append(toUpdate, outFlow)

			totalRate = totalRate.Add(outFlow.Rate)
			count++
		}

		for _, outFlow := range toUpdate {
			k.SetOutFlow(ctx, addr, &outFlow)
		}

		streamRecord.NetflowRate = streamRecord.NetflowRate.Add(totalRate.Neg())
		streamRecord.FrozenNetflowRate = streamRecord.FrozenNetflowRate.Add(totalRate)
		if !flowIterator.Valid() || finished {
			if !streamRecord.FrozenNetflowRate.IsZero() {
				ctx.Logger().Error("should not happen, stream frozen netflow rate is not zero", "address", streamRecord.Account)
			}
			streamRecord.Status = types.STREAM_ACCOUNT_STATUS_ACTIVE
			change := types.NewDefaultStreamRecordChangeWithAddr(addr)
			err := k.UpdateStreamRecord(ctx, streamRecord, change)
			if err != nil {
				ctx.Logger().Error("auto resume, update  stream record failed", "err", err.Error())
				panic("should not happen")
			}
			k.RemoveAutoResumeRecord(ctx, record.Timestamp, addr)
		}
		k.SetStreamRecord(ctx, streamRecord)
	}
}
