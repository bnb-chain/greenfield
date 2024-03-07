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

	// check the bucket owner is the same as the bucket owner of the bucket
	if found && bucket.Owner != bucketOwner.String() {
		return fmt.Errorf("the bucket owner is not the same as the bucket owner of the bucket")
	}

	// if the bucket does not use the payment account, just set the flow rate limit
	if !found || (found && bucket.PaymentAddress != paymentAccount.String()) {
		// set the flow rate limit to the store
		k.setBucketFlowRateLimit(ctx, paymentAccount, bucketName, &types.BucketFlowRateLimit{
			FlowRateLimit: rateLimit,
			BucketOwner:   bucketOwner.String(),
		})
		return nil
	}

	// delete the flow rate limit for different bucket owner, since there may be a wrong rate limit or
	// the previous rate limit set for the bucket which owned by the different bucket owner, for example,
	// the previous bucket owner delete the bucket and the current bucket owner create a new bucket with the same name
	k.deleteBucketFlowRateLimitForDifferentBucketOwner(ctx, paymentAccount, bucketOwner, bucketName)

	// set the flow rate limit for the bucket for the current bucket owner
	err := k.setFlowRateLimit(ctx, bucket, paymentAccount, bucketName, rateLimit)
	if err != nil {
		return err
	}

	// set the flow rate limit to the store
	k.setBucketFlowRateLimit(ctx, paymentAccount, bucketName, &types.BucketFlowRateLimit{
		FlowRateLimit: rateLimit,
		BucketOwner:   bucketOwner.String(),
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
	k.SetInternalBucketInfo(ctx, bucketInfo.Id, internalBucketInfo)
	return nil
}

func getTotalOutFlowRate(flows []paymenttypes.OutFlow) sdkmath.Int {
	totalFlowRate := sdkmath.ZeroInt()
	for _, flow := range flows {
		totalFlowRate = totalFlowRate.Add(flow.Rate)
	}
	return totalFlowRate
}

// unChargeAndLimitBucket uncharges the bucket and limits the bucket
func (k Keeper) unChargeAndLimitBucket(ctx sdk.Context, bucketInfo *types.BucketInfo, paymentAccount sdk.AccAddress, bucketName string) error {
	k.setBucketFlowRateLimitStatus(ctx, bucketName, &types.BucketFlowRateLimitStatus{
		IsBucketLimited: true,
		PaymentAddress:  paymentAccount.String(),
	})

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)
	// if the rate limit is not set and the payment account is not owned by the bucket owner,
	// the net flow rate of the bucket should be zero, but it's fine to call `unChargeBucketReadStoreFee`, it will do nothing
	return k.unChargeBucketReadStoreFee(ctx, bucketInfo, internalBucketInfo)
}

// setFlowRateLimit sets the flow rate limit of the bucket
func (k Keeper) setFlowRateLimit(ctx sdk.Context, bucketInfo *types.BucketInfo, paymentAccount sdk.AccAddress, bucketName string, rateLimit sdkmath.Int) error {
	currentRateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)

	// do nothing if the bucket flow rate limit is already the same as the new rate limit
	if found && currentRateLimit.FlowRateLimit.Equal(rateLimit) {
		return nil
	}

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)
	isRateLimited := k.isBucketRateLimited(ctx, bucketName)

	if found && isRateLimited {
		internalBucketInfo.PriceTime = ctx.BlockTime().Unix()

		currentBill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
		if err != nil {
			return fmt.Errorf("get bucket currentBill failed: %s %s", bucketInfo.BucketName, err.Error())
		}
		totalOutFlowRate := getTotalOutFlowRate(currentBill.Flows)

		// if the total net flow rate is less than or equal to the new rate limit,
		// resume the charge of the bucket and delete the flow rate limit status
		if totalOutFlowRate.LTE(rateLimit) {
			err := k.chargeBucketReadStoreFee(ctx, bucketInfo, internalBucketInfo)
			if err != nil {
				return fmt.Errorf("charge bucket failed: %s %s", bucketInfo.BucketName, err.Error())
			}
			k.deleteBucketFlowRateLimitStatus(ctx, bucketName)
			return nil
		}

		return nil
	}

	// if there is no rate limit before or the bucket is not rate limited, we just need to compare the
	// total out flow rate with the new rate limit, if the total out flow rate is greater than the new rate limit,
	// we should uncharge and limit the bucket
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	totalOutFlowRate := getTotalOutFlowRate(bill.Flows)
	if totalOutFlowRate.GT(rateLimit) {
		return k.unChargeAndLimitBucket(ctx, bucketInfo, paymentAccount, bucketName)
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

// deleteBucketFlowRateLimitForDifferentBucketOwner deletes the flow rate limit of the bucket for differnt bucket owner from the store
func (k Keeper) deleteBucketFlowRateLimitForDifferentBucketOwner(ctx sdk.Context, paymentAccount, bucketOwner sdk.AccAddress, bucketName string) {
	rateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)
	if !found {
		return
	}

	if rateLimit.BucketOwner != bucketOwner.String() {
		store := ctx.KVStore(k.storeKey)
		store.Delete(types.GetBucketFlowRateLimitKey(paymentAccount, bucketName))
	}
}

// setBucketFlowRateLimitStatus sets the flow rate limit status of the bucket to the store
func (k Keeper) setBucketFlowRateLimitStatus(ctx sdk.Context, bucketName string, status *types.BucketFlowRateLimitStatus) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(status)
	store.Set(types.GetBucketFlowRateLimitStatusKey(bucketName), bz)
}

// getBucketFlowRateLimitStatus returns the flow rate limit status of the bucket from the store
func (k Keeper) getBucketFlowRateLimitStatus(ctx sdk.Context, bucketName string) (*types.BucketFlowRateLimitStatus, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetBucketFlowRateLimitStatusKey(bucketName))
	if bz == nil {
		return nil, false
	}

	var status types.BucketFlowRateLimitStatus
	k.cdc.MustUnmarshal(bz, &status)
	return &status, true
}

// deleteBucketFlowRateLimitStatus deletes the flow rate limit status of the bucket from the store
func (k Keeper) deleteBucketFlowRateLimitStatus(ctx sdk.Context, bucketName string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetBucketFlowRateLimitStatusKey(bucketName))
}

// isBucketRateLimited checks if the bucket is rate limited
func (k Keeper) isBucketRateLimited(ctx sdk.Context, bucketName string) bool {
	status, found := k.getBucketFlowRateLimitStatus(ctx, bucketName)
	if !found {
		return false
	}
	return status.IsBucketLimited
}

// isBucketFlowRateUnderLimit checks if the flow rate of the bucket is under the flow rate limit
func (k Keeper) isBucketFlowRateUnderLimit(ctx sdk.Context, paymentAccount, bucketOwner sdk.AccAddress, bucketName string, userFlows paymenttypes.UserFlows) error {
	totalFlowRate := getTotalOutFlowRate(userFlows.Flows)

	return k.isBucketFlowRateUnderLimitWithRate(ctx, paymentAccount, bucketOwner, bucketName, totalFlowRate)
}

// isBucketFlowRateUnderLimitWithRate checks if the flow rate of the bucket is under the flow rate limit
func (k Keeper) isBucketFlowRateUnderLimitWithRate(ctx sdk.Context, paymentAccount, bucketOwner sdk.AccAddress, bucketName string, rate sdkmath.Int) error {
	// if the total net flow rate is zero, it should be allowed
	if rate.IsZero() {
		return nil
	}

	isPaymentAccountOwner := k.paymentKeeper.IsPaymentAccountOwner(ctx, paymentAccount, bucketOwner)

	rateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAccount, bucketName)
	// if the rate limit is not set
	if !found {
		// if the bucket owner is owner of the payment account and the rate limit is not set, the flow rate is unlimited
		if isPaymentAccountOwner {
			return nil
		}

		return fmt.Errorf("the flow rate limit is not set for the bucket %s", bucketName)
	}

	// check the flow rate limit is granted to the bucket owner
	if rateLimit.BucketOwner != bucketOwner.String() {
		return fmt.Errorf("the flow rate limit is not granted to the bucket owner %s", bucketOwner.String())
	}

	// check the flow rate is under the limit
	if rate.GT(rateLimit.FlowRateLimit) {
		return fmt.Errorf("the total flow rate of the bucket %s is greater than the flow rate limit", bucketName)
	}
	return nil
}

// GetBucketExtraInfo returns the extra info of the bucket
func (k Keeper) GetBucketExtraInfo(ctx sdk.Context, bucketName string) (*types.BucketExtraInfo, error) {
	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	paymentAcc, err := sdk.AccAddressFromHexUnsafe(bucketInfo.PaymentAddress)
	if err != nil {
		return nil, err
	}
	rateLimit, found := k.getBucketFlowRateLimit(ctx, paymentAcc, bucketName)

	extraInfo := &types.BucketExtraInfo{}

	if !found {
		extraInfo.FlowRateLimit = sdk.NewInt(-1)
	} else {
		extraInfo.FlowRateLimit = rateLimit.FlowRateLimit
	}

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)

	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return nil, fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	totalOutFlowRate := getTotalOutFlowRate(bill.Flows)
	extraInfo.CurrentFlowRate = totalOutFlowRate

	return extraInfo, nil
}
