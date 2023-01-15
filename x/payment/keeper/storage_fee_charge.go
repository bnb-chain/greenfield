package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// assume StreamRecordChange is unique by Addr
func (k Keeper) ApplyStreamRecordChanges(ctx sdk.Context, streamRecordChanges []types.StreamRecordChange) error {
	for _, fc := range streamRecordChanges {
		_, err := k.UpdateStreamRecordByAddr(ctx, &fc)
		if err != nil {
			return fmt.Errorf("update stream record failed: %w", err)
		}
	}
	return nil
}

func (k Keeper) ApplyUSDFlowChanges(ctx sdk.Context, from string, flowChanges []types.OutFlowInUSD) (err error) {
	currentTime := ctx.BlockTime().Unix()
	currentBNBPrice, err := k.GetBNBPriceByTime(ctx, currentTime)
	if err != nil {
		return fmt.Errorf("get current bnb price failed: %w", err)
	}
	streamRecord, found := k.GetStreamRecord(ctx, from)
	if !found {
		streamRecord = types.NewStreamRecord(from, currentTime)
	}
	prevTime := streamRecord.CrudTimestamp
	priceChanged := false
	var prevBNBPrice types.BNBPrice
	if prevTime != currentTime {
		prevBNBPrice, err = k.GetBNBPriceByTime(ctx, prevTime)
		if err != nil {
			return fmt.Errorf("get bnb price by time failed: %w", err)
		}
		priceChanged = !prevBNBPrice.Equal(currentBNBPrice)
	}
	var streamRecordChanges []types.StreamRecordChange
	// calculate rate changes in flowChanges
	for _, flowChange := range flowChanges {
		rateChangeInBNB := USD2BNB(flowChange.Rate, currentBNBPrice)
		k.MergeStreamRecordChanges(&streamRecordChanges, []types.StreamRecordChange{
			types.NewDefaultStreamRecordChangeWithAddr(from).WithRateChange(rateChangeInBNB.Neg()),
			types.NewDefaultStreamRecordChangeWithAddr(flowChange.SpAddress).WithRateChange(rateChangeInBNB),
		})
	}
	// calculate rate changes if price changes
	if priceChanged {
		for _, flow := range streamRecord.OutFlowsInUSD {
			prevRateInBNB := USD2BNB(flow.Rate, prevBNBPrice)
			currentRateInBNB := USD2BNB(flow.Rate, currentBNBPrice)
			rateChangeInBNB := currentRateInBNB.Sub(prevRateInBNB)
			k.MergeStreamRecordChanges(&streamRecordChanges, []types.StreamRecordChange{
				types.NewDefaultStreamRecordChangeWithAddr(from).WithRateChange(rateChangeInBNB.Neg()),
				types.NewDefaultStreamRecordChangeWithAddr(flow.SpAddress).WithRateChange(rateChangeInBNB),
			})
		}
	}
	// update flows
	MergeOutFlows(&streamRecord.OutFlowsInUSD, flowChanges)
	k.SetStreamRecord(ctx, streamRecord)
	err = k.ApplyStreamRecordChanges(ctx, streamRecordChanges)
	if err != nil {
		return fmt.Errorf("apply stream record changes failed: %w", err)
	}
	return nil
}

func USD2BNB(usd sdkmath.Int, bnbPrice types.BNBPrice) (bnb sdkmath.Int) {
	return usd.Mul(bnbPrice.Precision).Quo(bnbPrice.Num)
}

func MergeOutFlows(flow *[]types.OutFlowInUSD, changes []types.OutFlowInUSD) []types.OutFlowInUSD {
	for _, change := range changes {
		found := false
		for i, f := range *flow {
			if f.SpAddress == change.SpAddress {
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

func (k Keeper) ChargeInitialReadFee(ctx sdk.Context, bucketMeta *types.MockBucketMeta) error {
	currentTime := ctx.BlockTime().Unix()
	price, err := k.GetReadPrice(ctx, bucketMeta.ReadPacket, currentTime)
	if err != nil {
		return fmt.Errorf("get read price failed: %w", err)
	}
	flowChanges := []types.OutFlowInUSD{
		{SpAddress: bucketMeta.SpAddress, Rate: price},
	}
	return k.ApplyUSDFlowChanges(ctx, bucketMeta.ReadPaymentAccount, flowChanges)
}

func (k Keeper) ChargeUpdateReadPacket(ctx sdk.Context, bucketMeta *types.MockBucketMeta, newReadPacket types.ReadPacket) error {
	prevPrice, err := k.GetReadPrice(ctx, bucketMeta.ReadPacket, bucketMeta.PriceTime)
	if err != nil {
		return fmt.Errorf("get prev read price failed: %w", err)
	}
	newPrice, err := k.GetReadPrice(ctx, newReadPacket, ctx.BlockTime().Unix())
	if err != nil {
		return fmt.Errorf("get new read price failed: %w", err)
	}
	flowChanges := []types.OutFlowInUSD{
		{SpAddress: bucketMeta.SpAddress, Rate: newPrice.Sub(prevPrice)},
	}
	err = k.ApplyUSDFlowChanges(ctx, bucketMeta.ReadPaymentAccount, flowChanges)
	if err != nil {
		return fmt.Errorf("apply usd flow changes failed: %w", err)
	}
	return nil
}

func (k Keeper) LockStoreFeeByRate(ctx sdk.Context, user string, rate sdkmath.Int) (sdkmath.Int, error) {
	var lockAmountInBNB sdkmath.Int
	reserveTime := k.GetParams(ctx).ReserveTime
	bnbPrice, err := k.GetCurrentBNBPrice(ctx)
	if err != nil {
		return lockAmountInBNB, fmt.Errorf("get current bnb price failed: %w", err)
	}
	lockAmountInBNB = rate.Mul(sdkmath.NewIntFromUint64(reserveTime)).Mul(bnbPrice.Precision).Quo(bnbPrice.Num)
	change := types.NewDefaultStreamRecordChangeWithAddr(user).WithLockBalanceChange(lockAmountInBNB)
	streamRecord, err := k.UpdateStreamRecordByAddr(ctx, &change)
	if err != nil {
		return lockAmountInBNB, fmt.Errorf("update stream record failed: %w", err)
	}
	if streamRecord.StaticBalance.IsNegative() {
		return lockAmountInBNB, fmt.Errorf("static balance is not enough, lacks %s", streamRecord.StaticBalance.Neg().String())
	}
	return lockAmountInBNB, nil
}

func (k Keeper) LockStoreFee(ctx sdk.Context, bucketMeta *types.MockBucketMeta, objectInfo *types.MockObjectInfo) error {
	feePrice := k.GetStorePrice(ctx, bucketMeta, objectInfo)
	lockedBalance, err := k.LockStoreFeeByRate(ctx, bucketMeta.StorePaymentAccount, feePrice.UserPayRate)
	if err != nil {
		return fmt.Errorf("lock store fee by rate failed: %w", err)
	}
	objectInfo.LockedBalance = &lockedBalance
	return nil
}

// UnlockStoreFee unlock store fee if the object is deleted in INIT state
func (k Keeper) UnlockStoreFee(ctx sdk.Context, bucketMeta *types.MockBucketMeta, objectInfo *types.MockObjectInfo) error {
	lockedBalance := objectInfo.LockedBalance
	change := types.NewDefaultStreamRecordChangeWithAddr(bucketMeta.StorePaymentAccount).WithLockBalanceChange(lockedBalance.Neg())
	_, err := k.UpdateStreamRecordByAddr(ctx, &change)
	if err != nil {
		return fmt.Errorf("update stream record failed: %w", err)
	}
	return nil
}

func (k Keeper) UnlockAndChargeStoreFee(ctx sdk.Context, bucketMeta *types.MockBucketMeta, objectInfo *types.MockObjectInfo) error {
	// todo: what if store payment account is changed before unlock?
	feePrice := k.GetStorePrice(ctx, bucketMeta, objectInfo)
	err := k.UnlockStoreFee(ctx, bucketMeta, objectInfo)
	if err != nil {
		return fmt.Errorf("unlock store fee failed: %w", err)
	}
	err = k.ApplyUSDFlowChanges(ctx, bucketMeta.StorePaymentAccount, feePrice.Flows)
	if err != nil {
		return fmt.Errorf("apply usd flow changes failed: %w", err)
	}
	MergeOutFlows(&bucketMeta.OutFlowsInUSD, feePrice.Flows)
	return nil
}

func (k Keeper) ChargeUpdatePaymentAccount(ctx sdk.Context, bucketMeta *types.MockBucketMeta, readPaymentAccount, storePaymentAccount *string) error {
	if readPaymentAccount != nil {
		// update old read payment account
		prevReadPrice, err := k.GetReadPrice(ctx, bucketMeta.ReadPacket, bucketMeta.PriceTime)
		if err != nil {
			return fmt.Errorf("get prev read price failed: %w", err)
		}
		err = k.ApplyUSDFlowChanges(ctx, bucketMeta.ReadPaymentAccount, []types.OutFlowInUSD{
			{SpAddress: bucketMeta.SpAddress, Rate: prevReadPrice.Neg()},
		})
		if err != nil {
			return fmt.Errorf("apply prev read payment account usd flow changes failed: %w", err)
		}
		// update new read payment account
		currentReadPrice, err := k.GetReadPrice(ctx, bucketMeta.ReadPacket, ctx.BlockTime().Unix())
		if err != nil {
			return fmt.Errorf("get current read price failed: %w", err)
		}
		err = k.ApplyUSDFlowChanges(ctx, *readPaymentAccount, []types.OutFlowInUSD{
			{SpAddress: bucketMeta.SpAddress, Rate: currentReadPrice},
		})
		if err != nil {
			return fmt.Errorf("apply current read payment account usd flow changes failed: %w", err)
		}
		// update bucket meta
		bucketMeta.ReadPaymentAccount = *readPaymentAccount
		bucketMeta.PriceTime = ctx.BlockTime().Unix()
	}
	if storePaymentAccount != nil {
		flows := bucketMeta.OutFlowsInUSD
		negFlows := make([]types.OutFlowInUSD, len(flows))
		for i, flow := range flows {
			negFlows[i] = types.OutFlowInUSD{SpAddress: flow.SpAddress, Rate: flow.Rate.Neg()}
		}
		err := k.ApplyUSDFlowChanges(ctx, bucketMeta.StorePaymentAccount, negFlows)
		if err != nil {
			return fmt.Errorf("apply prev store payment account usd flow changes failed: %w", err)
		}
		err = k.ApplyUSDFlowChanges(ctx, *storePaymentAccount, flows)
		if err != nil {
			return fmt.Errorf("apply current store payment account usd flow changes failed: %w", err)
		}
		bucketMeta.StorePaymentAccount = *storePaymentAccount
	}
	return nil
}
