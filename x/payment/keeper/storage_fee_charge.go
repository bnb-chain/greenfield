package keeper

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// MergeStreamRecordChanges merge changes with same address
func (k Keeper) MergeStreamRecordChanges(changes []types.StreamRecordChange) []types.StreamRecordChange {
	if len(changes) <= 1 {
		return changes
	}
	changeMap := make(map[string]*types.StreamRecordChange)
	for _, change := range changes {
		currentChange, ok := changeMap[change.Addr.String()]
		if !ok {
			currentChange = types.NewDefaultStreamRecordChangeWithAddr(change.Addr)
		}
		currentChange.RateChange = currentChange.RateChange.Add(change.RateChange)
		currentChange.StaticBalanceChange = currentChange.StaticBalanceChange.Add(change.StaticBalanceChange)
		currentChange.LockBalanceChange = currentChange.LockBalanceChange.Add(change.LockBalanceChange)
		changeMap[change.Addr.String()] = currentChange
	}
	var result []types.StreamRecordChange
	for _, change := range changeMap {
		result = append(result, *change)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Addr.String() < result[j].Addr.String()
	})
	return result
}

// ApplyStreamRecordChanges assume StreamRecordChange is unique by Addr
func (k Keeper) ApplyStreamRecordChanges(ctx sdk.Context, streamRecordChanges []types.StreamRecordChange) error {
	streamRecordChanges = k.MergeStreamRecordChanges(streamRecordChanges)
	for i := 0; i < len(streamRecordChanges); i++ {
		_, err := k.UpdateStreamRecordByAddr(ctx, &streamRecordChanges[i])
		if err != nil {
			return fmt.Errorf("update stream record failed: %w", err)
		}
	}
	return nil
}

// ApplyUserFlowsList
func (k Keeper) ApplyUserFlowsList(ctx sdk.Context, userFlowsList []types.UserFlows) (err error) {
	userFlowsList = k.MergeUserFlows(userFlowsList)
	currentTime := ctx.BlockTime().Unix()

	for _, userFlows := range userFlowsList {
		from := userFlows.From
		streamRecord, found := k.GetStreamRecord(ctx, from)
		if !found {
			streamRecord = types.NewStreamRecord(from, currentTime)
		}
		if streamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE {
			err = k.applyActiveUserFlows(ctx, userFlows, from, streamRecord)
			if err != nil {
				return err
			}
		} else { // frozen status, should be called in end block for stop serving (uncharge fee)
			err = k.applyFrozenUserFlows(ctx, userFlows, from, streamRecord)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k Keeper) applyActiveUserFlows(ctx sdk.Context, userFlows types.UserFlows, from sdk.AccAddress, streamRecord *types.StreamRecord) error {
	var rateChanges []types.StreamRecordChange
	totalRate := sdk.ZeroInt()
	for _, flowChange := range userFlows.Flows {
		rateChanges = append(rateChanges, *types.NewDefaultStreamRecordChangeWithAddr(sdk.MustAccAddressFromHex(flowChange.ToAddress)).WithRateChange(flowChange.Rate))
		totalRate = totalRate.Add(flowChange.Rate)
	}
	streamRecordChange := types.NewDefaultStreamRecordChangeWithAddr(from).WithRateChange(totalRate.Neg())
	// storage fee preview
	if ctx.IsCheckTx() {
		reserveTime := k.GetParams(ctx).VersionedParams.ReserveTime
		changeRate := totalRate.Neg()
		event := &types.EventFeePreview{
			Account: from.String(),
			Amount:  changeRate.Mul(sdkmath.NewIntFromUint64(reserveTime)).Abs(),
		}
		if changeRate.IsPositive() {
			event.FeePreviewType = types.FEE_PREVIEW_TYPE_UNLOCKED_FEE
		} else {
			event.FeePreviewType = types.FEE_PREVIEW_TYPE_PRELOCKED_FEE
		}
		_ = ctx.EventManager().EmitTypedEvents(event)
	}
	err := k.UpdateStreamRecord(ctx, streamRecord, streamRecordChange)
	if err != nil {
		return fmt.Errorf("apply stream record changes for user failed: %w", err)
	}

	// update flows
	deltaFlowCount := k.MergeActiveOutFlows(ctx, from, userFlows.Flows) // deltaFlowCount can be negative
	streamRecord.OutFlowCount = uint64(int64(streamRecord.OutFlowCount) + int64(deltaFlowCount))

	k.SetStreamRecord(ctx, streamRecord)
	err = k.ApplyStreamRecordChanges(ctx, rateChanges)
	if err != nil {
		return fmt.Errorf("apply stream record changes failed: %w", err)
	}
	return nil
}

func (k Keeper) applyFrozenUserFlows(ctx sdk.Context, userFlows types.UserFlows, from sdk.AccAddress, streamRecord *types.StreamRecord) error {
	forced, _ := ctx.Value(types.ForceUpdateStreamRecordKey).(bool)
	if !forced {
		return fmt.Errorf("stream record %s is frozen", streamRecord.Account)
	}

	// the stream record could be totally frozen, or in the process of resuming
	var activeOutFlows, frozenOutFlows []types.OutFlow
	var activeRateChanges []types.StreamRecordChange
	//var frozenRateChanges []types.StreamRecordChange
	totalActiveRate, totalFrozenRate := sdk.ZeroInt(), sdk.ZeroInt()
	for _, flowChange := range userFlows.Flows {
		outFlow := k.GetOutFlow(ctx, sdk.MustAccAddressFromHex(streamRecord.Account), types.OUT_FLOW_STATUS_ACTIVE, sdk.MustAccAddressFromHex(flowChange.ToAddress))
		if outFlow != nil {
			activeOutFlows = append(activeOutFlows, flowChange)
			activeRateChanges = append(activeRateChanges, *types.NewDefaultStreamRecordChangeWithAddr(sdk.MustAccAddressFromHex(flowChange.ToAddress)).WithRateChange(flowChange.Rate))
			totalActiveRate = totalActiveRate.Add(flowChange.Rate)
		} else {
			frozenOutFlows = append(frozenOutFlows, flowChange)
			//frozenRateChanges = append(frozenRateChanges, *types.NewDefaultStreamRecordChangeWithAddr(sdk.MustAccAddressFromHex(flowChange.ToAddress)).WithFrozenRateChange(flowChange.Rate))
			totalFrozenRate = totalFrozenRate.Add(flowChange.Rate)
		}
	}
	streamRecordChange := types.NewDefaultStreamRecordChangeWithAddr(from).
		WithRateChange(totalActiveRate.Neg()).WithFrozenRateChange(totalFrozenRate.Neg())
	err := k.UpdateStreamRecord(ctx, streamRecord, streamRecordChange)
	if err != nil {
		return fmt.Errorf("apply stream record changes for user failed: %w", err)
	}

	// update flows
	deltaActiveFlowCount := k.MergeActiveOutFlows(ctx, from, activeOutFlows) // can be negative
	deltaFrozenFlowCount := k.MergeFrozenOutFlows(ctx, from, frozenOutFlows) // can be negative
	streamRecord.OutFlowCount = uint64(int64(streamRecord.OutFlowCount) + int64(deltaActiveFlowCount) + int64(deltaFrozenFlowCount))

	k.SetStreamRecord(ctx, streamRecord)
	//only apply activeRateChanges, for frozen rate changes, the out flow to gvg & gvg family had been deducted when settling
	err = k.ApplyStreamRecordChanges(ctx, activeRateChanges)
	if err != nil {
		return fmt.Errorf("apply stream record changes failed: %w", err)
	}
	return nil
}

// MergeUserFlows merge flows with same From address
func (k Keeper) MergeUserFlows(userFlowsList []types.UserFlows) []types.UserFlows {
	if len(userFlowsList) <= 1 {
		return userFlowsList
	}
	userFlowsMap := make(map[string][]types.OutFlow)
	for _, userFlows := range userFlowsList {
		flows := append(userFlowsMap[userFlows.From.String()], userFlows.Flows...)
		userFlowsMap[userFlows.From.String()] = flows
	}
	var newUserFlowsList []types.UserFlows
	for from, userFlows := range userFlowsMap {
		newUserFlowsList = append(newUserFlowsList, types.UserFlows{
			From:  sdk.MustAccAddressFromHex(from),
			Flows: k.MergeOutFlows(userFlows),
		})
	}
	sort.Slice(newUserFlowsList, func(i, j int) bool {
		return newUserFlowsList[i].From.String() < newUserFlowsList[j].From.String()
	})
	return newUserFlowsList
}

// MergeOutFlows merge flows with same address
func (k Keeper) MergeOutFlows(flows []types.OutFlow) []types.OutFlow {
	if len(flows) <= 1 {
		return flows
	}
	flowMap := make(map[string]sdkmath.Int)
	for _, flow := range flows {
		rate, found := flowMap[flow.ToAddress]
		if found {
			flowMap[flow.ToAddress] = rate.Add(flow.Rate)
		} else {
			flowMap[flow.ToAddress] = flow.Rate
		}
	}
	var newFlows []types.OutFlow
	for addr, rate := range flowMap {
		if rate.IsZero() {
			continue
		}
		newFlows = append(newFlows, types.OutFlow{
			ToAddress: addr,
			Rate:      rate,
		})
	}
	sort.Slice(newFlows, func(i, j int) bool {
		return newFlows[i].ToAddress < newFlows[j].ToAddress
	})
	return newFlows
}
