package keeper

import (
	sdkmath "cosmossdk.io/math"
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
				(*base)[i].RateChange = baseChange.RateChange.Add(newChange.RateChange)
				(*base)[i].StaticBalanceChange = baseChange.StaticBalanceChange.Add(newChange.StaticBalanceChange)
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
	//			RateChange:          sdkmath.ZeroInt(),
	//			StaticBalanceChange: sdkmath.ZeroInt(),
	//		}
	//	}
	//	fc.RateChange = fc.RateChange.Add(flowChange.RateChange)
	//	fc.StaticBalanceChange = fc.StaticBalanceChange.Add(flowChange.StaticBalanceChange)
	//	rateChangesSum = rateChangesSum.Add(flowChange.RateChange)
	//	flowChangeMap[flowChange.Addr] = fc
	//}
	//if !rateChangesSum.IsZero() {
	//	return fmt.Errorf("rate changes sum is not zero: %s", rateChangesSum.String())
	//}
	// charge fee
	for _, fc := range streamRecordChanges {
		// todo: check is payment account not accurate
		_, isPaymentAccount := k.GetPaymentAccount(ctx, fc.Addr)
		change := types.NewDefaultStreamRecordChangeWithAddr(fc.Addr).WithRateChange(fc.RateChange).WithStaticBalanceChange(fc.StaticBalanceChange).WithAutoTransfer(!isPaymentAccount)
		_, err := k.UpdateStreamRecordByAddr(ctx, &change)
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
		fromFc.RateChange = fromFc.RateChange.Sub(flowChange.Rate)
		toFc, found := streamRecordChangeMap[flowChange.To]
		if !found {
			fc := types.NewDefaultStreamRecordChangeWithAddr(flowChange.To)
			toFc = &fc
			streamRecordChangeMap[flowChange.To] = toFc
		}
		toFc.RateChange = toFc.RateChange.Add(flowChange.Rate)
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

func (k Keeper) LockStoreFeeByRate(ctx sdk.Context, user string, rate sdkmath.Int) error {
	reserveTime := k.GetParams(ctx).ReserveTime
	bnbPriceNum, bnbPricePrecision, err := k.GetCurrentBNBPrice(ctx)
	if err != nil {
		return fmt.Errorf("get current bnb price failed: %w", err)
	}
	lockAmountInBNB := rate.Mul(sdkmath.NewIntFromUint64(reserveTime)).Mul(bnbPricePrecision).Quo(bnbPriceNum)
	change := types.NewDefaultStreamRecordChangeWithAddr(user).WithLockBalanceChange(lockAmountInBNB.Neg()).WithAutoTransfer(true)
	streamRecord, err := k.UpdateStreamRecordByAddr(ctx, &change)
	if err != nil {
		return fmt.Errorf("update stream record failed: %w", err)
	}
	if streamRecord.StaticBalance.LT(streamRecord.LockBalance) {
		return fmt.Errorf("static balance is not enough, lacks %s", streamRecord.StaticBalance.String())
	}
	return nil
}

func (k Keeper) LockStoreFee(ctx sdk.Context, bucketMeta *types.MockBucketMeta, objectInfo *types.MockObjectInfo) error {
	feePrice := k.GetStorePrice(ctx, bucketMeta, objectInfo)
	return k.LockStoreFeeByRate(ctx, bucketMeta.StorePaymentAccount, feePrice.UserPayRate)
}
