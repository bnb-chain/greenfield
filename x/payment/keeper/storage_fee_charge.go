package keeper

import (
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//// given two price time, return the price diff between them
//func (k Keeper) GetReadPriceDiff(beforeTime, afterTime int64, beforeReadPacket, afterReadPacket types.ReadPacket) (sdkmath.Int, error) {
//	if beforeTime == afterTime {
//		return sdkmath.ZeroInt(), nil
//	}
//	beforeReadPrice, err := GetReadPrice(beforeReadPacket, beforeTime)
//	if err != nil {
//		return sdkmath.ZeroInt(), fmt.Errorf("get before read price failed: %w", err)
//	}
//	afterReadPrice, err := GetReadPrice(afterReadPacket, afterTime)
//	if err != nil {
//		return sdkmath.ZeroInt(), fmt.Errorf("get after read price failed: %w", err)
//	}
//}

func (k Keeper) MergeStreamRecordChanges(base *[]types.StreamRecordChange, newChanges []types.StreamRecordChange) {
	// merge changes with same address
	for _, newChange := range newChanges {
		found := false
		for i, baseChange := range *base {
			if baseChange.Addr == newChange.Addr {
				(*base)[i].Rate = baseChange.Rate.Add(newChange.Rate)
				(*base)[i].StaticBalance = baseChange.StaticBalance.Add(newChange.StaticBalance)
				found = true
				break
			}
		}
		if !found {
			*base = append(*base, newChange)
		}
	}
}

// assume StreamRecordChange is unique by Addr
func (k Keeper) ApplyStreamRecordChanges(ctx sdk.Context, streamRecordChanges []types.StreamRecordChange) error {
	//flowChangeMap := make(map[string]types.StreamRecordChange)
	//rateChangesSum := sdkmath.ZeroInt()
	//// merge changes with same address
	//for _, flowChange := range flowChanges {
	//	fc, found := flowChangeMap[flowChange.Addr]
	//	if !found {
	//		fc = types.StreamRecordChange{
	//			Addr:          flowChange.Addr,
	//			Rate:          sdkmath.ZeroInt(),
	//			StaticBalance: sdkmath.ZeroInt(),
	//		}
	//	}
	//	fc.Rate = fc.Rate.Add(flowChange.Rate)
	//	fc.StaticBalance = fc.StaticBalance.Add(flowChange.StaticBalance)
	//	rateChangesSum = rateChangesSum.Add(flowChange.Rate)
	//	flowChangeMap[flowChange.Addr] = fc
	//}
	//if !rateChangesSum.IsZero() {
	//	return fmt.Errorf("rate changes sum is not zero: %s", rateChangesSum.String())
	//}
	// charge fee
	for _, fc := range streamRecordChanges {
		_, isPaymentAccount := k.GetPaymentAccount(ctx, fc.Addr)
		err := k.UpdateStreamRecordByAddr(ctx, fc.Addr, fc.Rate, fc.StaticBalance, !isPaymentAccount)
		if err != nil {
			return fmt.Errorf("update stream record failed: %w", err)
		}
	}
	return nil
}

func (k Keeper) ApplyFlowChanges(ctx sdk.Context, flowChanges []types.Flow) error {
	streamRecordChangeMap := make(map[string]*types.StreamRecordChange)
	// merge changes with same address
	for _, flowChange := range flowChanges {
		fromFc, found := streamRecordChangeMap[flowChange.From]
		if !found {
			fc := types.NewDefaultStreamRecordChangeWithAddr(flowChange.From)
			fromFc = &fc
			streamRecordChangeMap[flowChange.From] = fromFc
		}
		fromFc.Rate = fromFc.Rate.Sub(flowChange.Rate)
		toFc, found := streamRecordChangeMap[flowChange.To]
		if !found {
			fc := types.NewDefaultStreamRecordChangeWithAddr(flowChange.To)
			toFc = &fc
			streamRecordChangeMap[flowChange.To] = toFc
		}
		toFc.Rate = toFc.Rate.Add(flowChange.Rate)
		// update flow
		err := k.UpdateFlow(ctx, flowChange)
		if err != nil {
			return fmt.Errorf("update flow failed: %w, flow: %+v", err, flowChange)
		}
	}
	streamRecordChanges := make([]types.StreamRecordChange, 0, len(streamRecordChangeMap))
	for _, fc := range streamRecordChangeMap {
		streamRecordChanges = append(streamRecordChanges, *fc)
	}
	// apply stream record changes
	err := k.ApplyStreamRecordChanges(ctx, streamRecordChanges)
	if err != nil {
		return fmt.Errorf("apply stream record changes failed: %w", err)
	}
	return nil
}

func (k Keeper) ChargeInitialReadFee(ctx sdk.Context, user, primarySP string, readPacket types.ReadPacket) error {
	currentTime := ctx.BlockTime().Unix()
	price, err := k.GetReadPrice(ctx, readPacket, currentTime)
	if err != nil {
		return fmt.Errorf("get read price failed: %w", err)
	}
	flowChanges := []types.Flow{
		{From: user, To: primarySP, Rate: price},
	}
	return k.ApplyFlowChanges(ctx, flowChanges)
}
