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
	var streamRecordChanges []types.StreamRecordChange
	for _, userFlows := range userFlowsList {
		from := userFlows.From
		streamRecord, found := k.GetStreamRecord(ctx, from)
		if !found {
			streamRecord = types.NewStreamRecord(from, currentTime)
		}
		// calculate rate changes in flowChanges
		totalRate := sdk.ZeroInt()
		for _, flowChange := range userFlows.Flows {
			streamRecordChanges = append(streamRecordChanges, *types.NewDefaultStreamRecordChangeWithAddr(sdk.MustAccAddressFromHex(flowChange.ToAddress)).WithRateChange(flowChange.Rate))
			totalRate = totalRate.Add(flowChange.Rate)
		}
		// update flows
		streamRecord.OutFlows = k.MergeOutFlows(append(streamRecord.OutFlows, userFlows.Flows...))
		streamRecordChange := types.NewDefaultStreamRecordChangeWithAddr(from).WithRateChange(totalRate.Neg())
		err = k.UpdateStreamRecord(ctx, streamRecord, streamRecordChange, false)
		if err != nil {
			return fmt.Errorf("apply stream record changes for user failed: %w", err)
		}
		k.SetStreamRecord(ctx, streamRecord)
	}
	err = k.ApplyStreamRecordChanges(ctx, streamRecordChanges)
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
