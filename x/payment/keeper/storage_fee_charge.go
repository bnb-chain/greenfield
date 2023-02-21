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
	for _, fc := range streamRecordChanges {
		_, err := k.UpdateStreamRecordByAddr(ctx, &fc)
		if err != nil {
			return fmt.Errorf("update stream record failed: %w", err)
		}
	}
	return nil
}

func (k Keeper) ApplyFlowChanges(ctx sdk.Context, from string, flowChanges []types.OutFlow) (err error) {
	//currentTime := ctx.BlockTime().Unix()
	//streamRecord, found := k.GetStreamRecord(ctx, from)
	//if !found {
	//	streamRecord = types.NewStreamRecord(from, currentTime)
	//}
	//prevTime := streamRecord.CrudTimestamp
	//priceChanged := false
	//var prevBNBPrice types.BNBPrice
	//if prevTime != currentTime {
	//	prevBNBPrice, err = k.GetBNBPriceByTime(ctx, prevTime)
	//	if err != nil {
	//		return fmt.Errorf("get bnb price by time failed: %w", err)
	//	}
	//	priceChanged = !prevBNBPrice.Equal(currentBNBPrice)
	//}
	//var streamRecordChanges []types.StreamRecordChange
	//// calculate rate changes in flowChanges
	//for _, flowChange := range flowChanges {
	//	rateChangeInBNB := USD2BNB(flowChange.Rate, currentBNBPrice)
	//	k.MergeStreamRecordChanges(&streamRecordChanges, []types.StreamRecordChange{
	//		*types.NewDefaultStreamRecordChangeWithAddr(from).WithRateChange(rateChangeInBNB.Neg()),
	//		*types.NewDefaultStreamRecordChangeWithAddr(flowChange.SpAddress).WithRateChange(rateChangeInBNB),
	//	})
	//}
	//// calculate rate changes if price changes
	//if priceChanged {
	//	for _, flow := range streamRecord.OutFlowsInUSD {
	//		prevRateInBNB := USD2BNB(flow.Rate, prevBNBPrice)
	//		currentRateInBNB := USD2BNB(flow.Rate, currentBNBPrice)
	//		rateChangeInBNB := currentRateInBNB.Sub(prevRateInBNB)
	//		k.MergeStreamRecordChanges(&streamRecordChanges, []types.StreamRecordChange{
	//			*types.NewDefaultStreamRecordChangeWithAddr(from).WithRateChange(rateChangeInBNB.Neg()),
	//			*types.NewDefaultStreamRecordChangeWithAddr(flow.SpAddress).WithRateChange(rateChangeInBNB),
	//		})
	//	}
	//}
	//// update flows
	//MergeOutFlows(&streamRecord.OutFlowsInUSD, flowChanges)
	//k.SetStreamRecord(ctx, streamRecord)
	//err = k.ApplyStreamRecordChanges(ctx, streamRecordChanges)
	//if err != nil {
	//	return fmt.Errorf("apply stream record changes failed: %w", err)
	//}
	return nil
}
