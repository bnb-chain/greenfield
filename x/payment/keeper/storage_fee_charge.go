package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) MergeStreamRecordChanges(base *[]types.StreamRecordChange, newChanges []types.StreamRecordChange) {
	// merge changes with same address
	for _, newChange := range newChanges {
		found := false
		for i, baseChange := range *base {
			if baseChange.Addr == newChange.Addr {
				(*base)[i].RateChange = baseChange.RateChange.Add(newChange.RateChange)
				(*base)[i].StaticBalanceChange = baseChange.StaticBalanceChange.Add(newChange.StaticBalanceChange)
				(*base)[i].LockBalanceChange = baseChange.LockBalanceChange.Add(newChange.LockBalanceChange)
				found = true
				break
			}
		}
		if !found {
			*base = append(*base, newChange)
		}
	}
}

// ApplyStreamRecordChanges assume StreamRecordChange is unique by Addr
func (k Keeper) ApplyStreamRecordChanges(ctx sdk.Context, streamRecordChanges []types.StreamRecordChange) error {
	for i := 0; i < len(streamRecordChanges); i++ {
		_, err := k.UpdateStreamRecordByAddr(ctx, &streamRecordChanges[i])
		if err != nil {
			return fmt.Errorf("update stream record failed: %w", err)
		}
	}
	return nil
}

func (k Keeper) ApplyFlowChanges(ctx sdk.Context, from string, flowChanges []types.OutFlow) (err error) {
	currentTime := ctx.BlockTime().Unix()
	streamRecord, found := k.GetStreamRecord(ctx, from)
	if !found {
		streamRecord = types.NewStreamRecord(from, currentTime)
	}
	var streamRecordChanges []types.StreamRecordChange
	// calculate rate changes in flowChanges
	for _, flowChange := range flowChanges {
		k.MergeStreamRecordChanges(&streamRecordChanges, []types.StreamRecordChange{
			*types.NewDefaultStreamRecordChangeWithAddr(from).WithRateChange(flowChange.Rate.Neg()),
			*types.NewDefaultStreamRecordChangeWithAddr(flowChange.ToAddress).WithRateChange(flowChange.Rate),
		})
	}
	// update flows
	MergeOutFlows(&streamRecord.OutFlows, flowChanges)
	k.SetStreamRecord(ctx, streamRecord)
	err = k.ApplyStreamRecordChanges(ctx, streamRecordChanges)
	if err != nil {
		return fmt.Errorf("apply stream record changes failed: %w", err)
	}
	return nil
}

func MergeOutFlows(flow *[]types.OutFlow, changes []types.OutFlow) []types.OutFlow {
	for _, change := range changes {
		found := false
		for i, f := range *flow {
			if f.ToAddress == change.ToAddress {
				found = true
				(*flow)[i].Rate = (*flow)[i].Rate.Add(change.Rate)
				break
			}
		}
		if !found {
			*flow = append(*flow, change)
		}
	}
	return *flow
}
