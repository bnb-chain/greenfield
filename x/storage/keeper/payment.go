package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) ChargeInitialReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) error {
	if bucketInfo.ReadQuota == 0 {
		return nil
	}
	bucketInfo.BillingInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %w", err)
	}
	return k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
}

func (k Keeper) UpdateBucketInfoAndCharge(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, newPaymentAddr string, newReadQuota uint64) error {
	if bucketInfo.PaymentAddress != newPaymentAddr && bucketInfo.ReadQuota != newReadQuota {
		return fmt.Errorf("payment address and read quota can not be changed at the same time")
	}
	err := k.ChargeViaBucketChange(ctx, bucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.PaymentAddress = newPaymentAddr
		bi.ReadQuota = newReadQuota
		return nil
	})
	return err
}

func (k Keeper) LockStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	amount, err := k.GetObjectLockFee(ctx, bucketInfo.PrimarySpAddress, objectInfo.CreateAt, objectInfo.PayloadSize)
	if err != nil {
		return fmt.Errorf("get object store fee rate failed: %w", err)
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(bucketInfo.PaymentAddress).WithLockBalanceChange(amount)
	streamRecord, err := k.paymentKeeper.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return fmt.Errorf("update stream record failed: %w", err)
	}
	if streamRecord.StaticBalance.IsNegative() {
		return fmt.Errorf("static balance is not enough, lacks %s", streamRecord.StaticBalance.Neg().String())
	}
	return nil
}

// UnlockStoreFee unlock store fee if the object is deleted in INIT state
func (k Keeper) UnlockStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	lockedBalance, err := k.GetObjectLockFee(ctx, bucketInfo.PrimarySpAddress, objectInfo.CreateAt, objectInfo.PayloadSize)
	if err != nil {
		return fmt.Errorf("get object store fee rate failed: %w", err)
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(bucketInfo.PaymentAddress).WithLockBalanceChange(lockedBalance.Neg())
	_, err = k.paymentKeeper.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return fmt.Errorf("update stream record failed: %w", err)
	}
	return nil
}

func (k Keeper) UnlockAndChargeStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	// unlock store fee
	err := k.UnlockStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return fmt.Errorf("unlock store fee failed: %w", err)
	}
	chargeSize := k.GetChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	return k.ChargeViaBucketChange(ctx, bucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.BillingInfo.TotalChargeSize += chargeSize
		for _, sp := range objectInfo.SecondarySpAddresses {
			bi.BillingInfo.SecondarySpObjectsSize = AddSecondarySpObjectsSize(bi.BillingInfo.SecondarySpObjectsSize, storagetypes.SecondarySpObjectsSize{
				SpAddress:       sp,
				TotalChargeSize: chargeSize,
			})
		}
		return nil
	})
}

func (k Keeper) ChargeViaBucketChange(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, changeFunc func(bucketInfo *storagetypes.BucketInfo) error) error {
	// get previous bill
	prevBill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %w", err)
	}
	// change bucket billing info
	if err = changeFunc(bucketInfo); err != nil {
		return errors.Wrapf(err, "change bucket billing info failed")
	}
	bucketInfo.BillingInfo.PriceTime = ctx.BlockTime().Unix()
	// calculate new bill
	newBill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return fmt.Errorf("get new bucket bill failed: %w", err)
	}
	// charge according to bill change
	err = k.ChargeAccordingToBillChange(ctx, prevBill, newBill)
	if err != nil {
		return fmt.Errorf("charge according to bill change failed: %w", err)
	}
	return nil
}

func (k Keeper) GetBucketBill(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) (userFlows types.UserFlows, err error) {
	userFlows.From = bucketInfo.PaymentAddress
	if bucketInfo.BillingInfo.TotalChargeSize == 0 && bucketInfo.ReadQuota == 0 {
		return userFlows, nil
	}
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: bucketInfo.PrimarySpAddress,
		PriceTime: bucketInfo.BillingInfo.PriceTime,
	})
	if err != nil {
		return userFlows, fmt.Errorf("get storage price failed: %w", err)
	}
	readFlowRate := price.ReadPrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.ReadQuota)).TruncateInt()
	primaryStoreFlowRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.BillingInfo.TotalChargeSize)).TruncateInt()
	primarySpRate := readFlowRate.Add(primaryStoreFlowRate)
	if primarySpRate.IsPositive() {
		userFlows.Flows = keeper.MergeOutFlows(&userFlows.Flows, []types.OutFlow{{
			ToAddress: bucketInfo.PrimarySpAddress,
			Rate:      primarySpRate,
		}})
	}
	for _, spObjectsSize := range bucketInfo.BillingInfo.SecondarySpObjectsSize {
		rate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(spObjectsSize.TotalChargeSize)).TruncateInt()
		if rate.IsZero() {
			continue
		}
		userFlows.Flows = keeper.MergeOutFlows(&userFlows.Flows, []types.OutFlow{{
			ToAddress: spObjectsSize.SpAddress,
			Rate:      rate,
		}})
	}
	return userFlows, nil
}

func (k Keeper) ChargeAccordingToBillChange(ctx sdk.Context, prev, current types.UserFlows) error {
	prev.Flows = GetNegFlows(prev.Flows)
	if prev.From == current.From {
		flowChanges := keeper.MergeOutFlows(&prev.Flows, current.Flows)
		err := k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{{From: prev.From, Flows: flowChanges}})
		if err != nil {
			return fmt.Errorf("apply flow changes failed: %w", err)
		}
	} else {
		err := k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{prev, current})
		if err != nil {
			return fmt.Errorf("apply user flows list failed: %w", err)
		}
	}
	return nil
}

func (k Keeper) ChargeDeleteObject(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize := k.GetChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	return k.ChargeViaBucketChange(ctx, bucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.BillingInfo.TotalChargeSize -= chargeSize
		var err error
		for _, sp := range objectInfo.SecondarySpAddresses {
			bucketInfo.BillingInfo.SecondarySpObjectsSize, err = SubSecondarySpObjectsSize(bucketInfo.BillingInfo.SecondarySpObjectsSize, storagetypes.SecondarySpObjectsSize{
				SpAddress:       sp,
				TotalChargeSize: chargeSize,
			})
			if err != nil {
				return errors.Wrapf(err, "sub secondary sp objects size")
			}
		}
		return nil
	})
}

func GetNegFlows(flows []types.OutFlow) (negFlows []types.OutFlow) {
	negFlows = make([]types.OutFlow, len(flows))
	for i, flow := range flows {
		negFlows[i] = types.OutFlow{ToAddress: flow.ToAddress, Rate: flow.Rate.Neg()}
	}
	return negFlows
}

func AddSecondarySpObjectsSize(prev []storagetypes.SecondarySpObjectsSize, new storagetypes.SecondarySpObjectsSize) []storagetypes.SecondarySpObjectsSize {
	found := false
	for i, spObjectsSize := range prev {
		if spObjectsSize.SpAddress == new.SpAddress {
			prev[i].TotalChargeSize += new.TotalChargeSize
			found = true
			break
		}
	}
	if !found {
		prev = append(prev, new)
	}
	return prev
}

func SubSecondarySpObjectsSize(prev []storagetypes.SecondarySpObjectsSize, toBeSub storagetypes.SecondarySpObjectsSize) ([]storagetypes.SecondarySpObjectsSize, error) {
	found := false
	for i, spObjectsSize := range prev {
		if spObjectsSize.SpAddress == toBeSub.SpAddress {
			if spObjectsSize.TotalChargeSize < toBeSub.TotalChargeSize {
				return nil, fmt.Errorf("secondary sp %s total charge size %d is less than to be sub %d", toBeSub.SpAddress, spObjectsSize.TotalChargeSize, toBeSub.TotalChargeSize)
			}
			prev[i].TotalChargeSize -= toBeSub.TotalChargeSize
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("secondary sp %s not found", toBeSub.SpAddress)
	}
	return prev, nil
}

func (k Keeper) GetObjectLockFee(ctx sdk.Context, primarySpAddress string, priceTime int64, payloadSize uint64) (amount sdkmath.Int, err error) {
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: primarySpAddress,
		PriceTime: priceTime,
	})
	if err != nil {
		return amount, fmt.Errorf("get store price failed: %w", err)
	}
	chargeSize := k.GetChargeSize(ctx, payloadSize, priceTime)
	rate := price.PrimaryStorePrice.Add(price.SecondaryStorePrice.MulInt64(storagetypes.SecondarySPNum)).MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	reserveTime := k.paymentKeeper.GetParams(ctx).ReserveTime
	amount = rate.Mul(sdkmath.NewIntFromUint64(reserveTime))
	return amount, nil
}

// todo(Fynn): refactor when we have a way to record the min charge size parameter history
func (k Keeper) GetChargeSize(ctx sdk.Context, payloadSize uint64, _time int64) uint64 {
	minChargeSize := k.GetParams(ctx).MinChargeSize
	if payloadSize < minChargeSize {
		return minChargeSize
	} else {
		return payloadSize
	}
}
