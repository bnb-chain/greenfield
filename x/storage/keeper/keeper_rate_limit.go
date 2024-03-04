package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) SetBucketFlowRateLimit(ctx sdk.Context, operator sdk.AccAddress, bucketOwner sdk.AccAddress, paymentAccount sdk.AccAddress, bucketName string, rateLimit sdkmath.Int) error {
	// check the operator is the same as the payment account owner
	if !k.paymentKeeper.IsPaymentAccountOwner(ctx, paymentAccount, operator) {
		return paymenttypes.ErrNotPaymentAccountOwner
	}

	// get the bucket
	bucket, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}

	// check the bucket owner is the same as the bucket owner
	if bucket.Owner != bucketOwner.String() {
		return types.ErrInvalidBucketOwner
	}

	// if the bucket does not use the payment account, just set the flow rate limit
	if bucket.PaymentAddress != paymentAccount.String() {
		// set the flow rate limit to the store
		k.setBucketFlowRateLimit(ctx, paymentAccount, bucketName, &types.BucketFlowRateLimit{
			FlowRateLimit: rateLimit,
		})
		return nil
	}

	// below are the logic for the bucket using the payment account
	if rateLimit.IsZero() {
		err := k.setZeroBucketFlowRateLimit(ctx, bucket, paymentAccount, bucketName)
		if err != nil {
			return err
		}
	} else {
		err := k.setNonZeroBucketFlowRateLimit(ctx, bucket, paymentAccount, bucketName, rateLimit)
		if err != nil {
			return err
		}
	}

	// set the flow rate limit to the store
	k.setBucketFlowRateLimit(ctx, paymentAccount, bucketName, &types.BucketFlowRateLimit{
		FlowRateLimit: rateLimit,
	})

	return nil
}

func (k Keeper) unChargeBucketReadStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	bill.Flows = getNegFlows(bill.Flows)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []paymenttypes.UserFlows{bill})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %s %w", bucketInfo.BucketName, err)
	}
	return nil
}

func (k Keeper) chargeBucketReadStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []paymenttypes.UserFlows{bill})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %s %w", bucketInfo.BucketName, err)
	}
	return nil
}

func getTotalOutFlowRate(flows []paymenttypes.OutFlow) sdkmath.Int {
	totalFlowRate := sdkmath.ZeroInt()
	for _, flow := range flows {
		totalFlowRate = totalFlowRate.Add(flow.Rate)
	}
	return totalFlowRate
}

func (k Keeper) setZeroBucketFlowRateLimit(ctx sdk.Context, bucketInfo *types.BucketInfo, paymentAccount sdk.AccAddress, bucketName string) error {
	currentRateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)

	// do nothing if the bucket flow rate limit is already 0
	if found && currentRateLimit.FlowRateLimit.IsZero() {
		return nil
	}

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)
	// if the rate limit is not set and the payment account is not owned by the bucket owner,
	// the net flow rate of the bucket should be zero, but it's fine to call `unChargeBucketReadStoreFee`, it will do nothing
	return k.unChargeBucketReadStoreFee(ctx, bucketInfo, internalBucketInfo)
}

func (k Keeper) setNonZeroBucketFlowRateLimit(ctx sdk.Context, bucketInfo *types.BucketInfo, paymentAccount sdk.AccAddress, bucketName string, rateLimit sdkmath.Int) error {
	currentRateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)

	// do nothing if the bucket flow rate limit is already the same as the new rate limit
	if found && currentRateLimit.FlowRateLimit.Equal(rateLimit) {
		return nil
	}

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)

	// if the rate limit is set to zero before, we should resume the payment of the bucket
	if found && currentRateLimit.FlowRateLimit.IsZero() {
		internalBucketInfo.PriceTime = ctx.BlockTime().Unix()

		bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
		if err != nil {
			return fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
		}
		totalOutFlowRate := getTotalOutFlowRate(bill.Flows)
		if totalOutFlowRate.GT(rateLimit) {
			return fmt.Errorf("the total out flow rate(%s) of the bucket is greater than the new rate limit(%s)", totalOutFlowRate.String(), rateLimit.String())
		}

		return k.chargeBucketReadStoreFee(ctx, bucketInfo, internalBucketInfo)
	}

	// if the rate limit is not set t0 zero before, we should check the net flow rate of the bucket
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	totalOutFlowRate := getTotalOutFlowRate(bill.Flows)
	if totalOutFlowRate.GT(rateLimit) {
		return fmt.Errorf("the total out flow rate(%s) of the bucket is greater than the new rate limit(%s)", totalOutFlowRate.String(), rateLimit.String())
	}

	return nil
}

// getBucketFlowRateLimit returns the flow rate limit of the bucket from the store
func (k Keeper) getBucketFlowRateLimit(ctx sdk.Context, paymentAccount sdk.AccAddress, bucketName string) (*types.BucketFlowRateLimit, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetBucketFlowRateLimitKey(paymentAccount, bucketName))
	if bz == nil {
		return nil, false
	}

	var rateLimit types.BucketFlowRateLimit
	k.cdc.MustUnmarshal(bz, &rateLimit)
	return &rateLimit, true
}

// setBucketFlowRateLimit sets the flow rate limit of the bucket to the store
func (k Keeper) setBucketFlowRateLimit(ctx sdk.Context, paymentAccount sdk.AccAddress, bucketName string, rateLimit *types.BucketFlowRateLimit) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(rateLimit)
	store.Set(types.GetBucketFlowRateLimitKey(paymentAccount, bucketName), bz)
}

// isBucketFlowRateLimitSetToZero checks if the flow rate limit of the bucket is set to zero by the payment account owner
func (k Keeper) isBucketFlowRateLimitSetToZero(ctx sdk.Context, paymentAccount sdk.AccAddress, bucketName string) bool {
	rateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)
	if !found {
		return false
	}
	return rateLimit.FlowRateLimit.IsZero()
}

// isBucketFlowRateUnderLimit checks if the flow rate of the bucket is under the flow rate limit
func (k Keeper) isBucketFlowRateUnderLimit(ctx sdk.Context, paymentAccount, bucketOwner sdk.AccAddress, bucketName string, userFlows paymenttypes.UserFlows) error {
	totalFlowRate := getTotalOutFlowRate(userFlows.Flows)

	isPaymentAccountOwner := k.paymentKeeper.IsPaymentAccountOwner(ctx, paymentAccount, bucketOwner)

	rateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)
	// if the rate limit is not set
	if !found {
		// if the bucket owner is owner of the payment account and the rate limit is not set, the flow rate is unlimited
		if isPaymentAccountOwner {
			return nil
		}

		// if the total net flow rate is zero, it should be allowed
		if totalFlowRate.IsZero() {
			return nil
		}

		return fmt.Errorf("the flow rate limit is not set for the bucket %s", bucketName)
	}

	if totalFlowRate.LTE(rateLimit.FlowRateLimit) {
		return nil
	}
	return fmt.Errorf("the total flow rate of the bucket(%s) %s is greater than the flow rate limit(%s)", totalFlowRate.String(), bucketName, rateLimit.String())
}

func (k Keeper) isBucketFlowRateUnderLimitWithRate(ctx sdk.Context, paymentAccount, bucketOwner sdk.AccAddress, bucketName string, rate sdkmath.Int) error {
	isPaymentAccountOwner := k.paymentKeeper.IsPaymentAccountOwner(ctx, paymentAccount, bucketOwner)

	rateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)
	// if the rate limit is not set
	if !found {
		// if the bucket owner is owner of the payment account and the rate limit is not set, the flow rate is unlimited
		if isPaymentAccountOwner {
			return nil
		}

		// if the total net flow rate is zero, it should be allowed
		if rate.IsZero() {
			return nil
		}

		return fmt.Errorf("the flow rate limit is not set for the bucket %s", bucketName)
	}

	if rate.LTE(rateLimit.FlowRateLimit) {
		return nil
	}
	return fmt.Errorf("the total flow rate of the bucket %s is greater than the flow rate limit", bucketName)
}
