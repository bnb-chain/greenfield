package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) ChargeInitialReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) error {
	currentTime := ctx.BlockTime().Unix()
	price, err := k.paymentKeeper.GetReadPrice(ctx, bucketInfo.PrimarySpAddress, bucketInfo.ReadQuota, currentTime)
	if err != nil {
		return fmt.Errorf("get read price failed: %w", err)
	}
	flowChanges := []types.OutFlow{
		{ToAddress: bucketInfo.PrimarySpAddress, Rate: price},
	}
	return k.paymentKeeper.ApplyFlowChanges(ctx, bucketInfo.PaymentAddress, flowChanges)
}

func (k Keeper) ChargeUpdateReadQuota(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, newReadPacket uint64) error {
	prevPrice, err := k.paymentKeeper.GetReadPrice(ctx, bucketInfo.PaymentAddress, bucketInfo.ReadQuota, bucketInfo.PaymentPriceTime)
	if err != nil {
		return fmt.Errorf("get prev read price failed: %w", err)
	}
	newPrice, err := k.paymentKeeper.GetReadPrice(ctx, bucketInfo.PaymentAddress, newReadPacket, ctx.BlockTime().Unix())
	if err != nil {
		return fmt.Errorf("get new read price failed: %w", err)
	}
	flowChanges := []types.OutFlow{
		{ToAddress: bucketInfo.PrimarySpAddress, Rate: newPrice.Sub(prevPrice)},
	}
	err = k.paymentKeeper.ApplyFlowChanges(ctx, bucketInfo.PaymentAddress, flowChanges)
	if err != nil {
		return fmt.Errorf("apply usd flow changes failed: %w", err)
	}
	return nil
}

func (k Keeper) LockStoreFeeByRate(ctx sdk.Context, user string, rate sdkmath.Int) (sdkmath.Int, error) {
	reserveTime := k.paymentKeeper.GetParams(ctx).ReserveTime
	lockAmount := rate.Mul(sdkmath.NewIntFromUint64(reserveTime))
	change := types.NewDefaultStreamRecordChangeWithAddr(user).WithLockBalanceChange(lockAmount)
	streamRecord, err := k.paymentKeeper.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return lockAmount, fmt.Errorf("update stream record failed: %w", err)
	}
	if streamRecord.StaticBalance.IsNegative() {
		return lockAmount, fmt.Errorf("static balance is not enough, lacks %s", streamRecord.StaticBalance.Neg().String())
	}
	return lockAmount, nil
}

func (k Keeper) LockStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	feePrice := k.paymentKeeper.GetStorePrice(ctx, bucketInfo, objectInfo)
	lockedBalance, err := k.LockStoreFeeByRate(ctx, bucketInfo.PaymentAddress, feePrice.UserPayRate)
	if err != nil {
		return fmt.Errorf("lock store fee by rate failed: %w", err)
	}
	objectInfo.LockedBalance = &lockedBalance
	return nil
}

// UnlockStoreFee unlock store fee if the object is deleted in INIT state
func (k Keeper) UnlockStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	lockedBalance := objectInfo.LockedBalance
	change := types.NewDefaultStreamRecordChangeWithAddr(bucketInfo.PaymentAddress).WithLockBalanceChange(lockedBalance.Neg())
	_, err := k.paymentKeeper.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return fmt.Errorf("update stream record failed: %w", err)
	}
	return nil
}

func (k Keeper) UnlockAndChargeStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	// todo: what if store payment account is changed before unlock?
	feePrice := k.paymentKeeper.GetStorePrice(ctx, bucketInfo, objectInfo)
	err := k.UnlockStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return fmt.Errorf("unlock store fee failed: %w", err)
	}
	err = k.paymentKeeper.ApplyFlowChanges(ctx, bucketInfo.PaymentAddress, feePrice.Flows)
	if err != nil {
		return fmt.Errorf("apply usd flow changes failed: %w", err)
	}
	return nil
}

func (k Keeper) ChargeUpdatePaymentAccount(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, paymentAddress *string) error {
	if paymentAddress != nil {
		// update old read payment account
		prevReadPrice, err := k.paymentKeeper.GetReadPrice(ctx, bucketInfo.PaymentAddress, bucketInfo.ReadQuota, bucketInfo.PaymentPriceTime)
		if err != nil {
			return fmt.Errorf("get prev read price failed: %w", err)
		}
		err = k.paymentKeeper.ApplyFlowChanges(ctx, bucketInfo.PaymentAddress, []types.OutFlow{
			{ToAddress: bucketInfo.PrimarySpAddress, Rate: prevReadPrice.Neg()},
		})
		if err != nil {
			return fmt.Errorf("apply prev read payment account usd flow changes failed: %w", err)
		}
		// update new read payment account
		currentReadPrice, err := k.paymentKeeper.GetReadPrice(ctx, bucketInfo.PaymentAddress, bucketInfo.ReadQuota, ctx.BlockTime().Unix())
		if err != nil {
			return fmt.Errorf("get current read price failed: %w", err)
		}
		err = k.paymentKeeper.ApplyFlowChanges(ctx, *paymentAddress, []types.OutFlow{
			{ToAddress: bucketInfo.PrimarySpAddress, Rate: currentReadPrice},
		})
		if err != nil {
			return fmt.Errorf("apply current read payment account usd flow changes failed: %w", err)
		}
		// update bucket meta
		bucketInfo.PaymentAddress = *paymentAddress
		bucketInfo.PaymentPriceTime = ctx.BlockTime().Unix()

		//// update old store
		//flows := bucketInfo.PaymentOutFlows
		//negFlows := GetNegFlows(flows)
		//err = k.paymentKeeper.ApplyFlowChanges(ctx, bucketInfo.PaymentAddress, negFlows)
		//if err != nil {
		//	return fmt.Errorf("apply prev store payment account usd flow changes failed: %w", err)
		//}
		//err = k.ApplyFlowChanges(ctx, *paymentAddress, flows)
		//if err != nil {
		//	return fmt.Errorf("apply current store payment account usd flow changes failed: %w", err)
		//}
		//bucketInfo.PaymentAddress = *paymentAddress
	}
	return nil
}

func (k Keeper) ChargeDeleteObject(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	feePrice := k.paymentKeeper.GetStorePrice(ctx, bucketInfo, objectInfo)
	negFlows := GetNegFlows(feePrice.Flows)
	err := k.paymentKeeper.ApplyFlowChanges(ctx, bucketInfo.PaymentAddress, negFlows)
	if err != nil {
		return fmt.Errorf("apply usd flow changes failed: %w", err)
	}
	return nil
}

func GetNegFlows(flows []types.OutFlow) (negFlows []types.OutFlow) {
	negFlows = make([]types.OutFlow, len(flows))
	for i, flow := range flows {
		negFlows[i] = types.OutFlow{ToAddress: flow.ToAddress, Rate: flow.Rate.Neg()}
	}
	return negFlows
}
