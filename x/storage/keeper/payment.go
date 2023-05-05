package keeper

import (
	"fmt"
	"sort"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) GetFundingAddressBySpAddr(ctx sdk.Context, spAddr sdk.AccAddress) (string, error) {
	sp, found := k.spKeeper.GetStorageProvider(ctx, spAddr)
	if !found {
		return "", fmt.Errorf("storage provider %s not found", spAddr)
	}
	return sp.FundingAddress, nil
}

func (k Keeper) ChargeInitialReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) error {
	if bucketInfo.ChargedReadQuota == 0 {
		return nil
	}
	bucketInfo.BillingInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %w", err)
	}
	return k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
}

func (k Keeper) ChargeDeleteBucket(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) error {
	bill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return err
	}
	if len(bill.Flows) == 0 {
		return nil
	}
	// should only remain at most 2 flows: charged_read_quota fee and tax
	if len(bill.Flows) > 2 {
		panic(fmt.Sprintf("unexpected left flow number: %d", len(bill.Flows)))
	}
	bill.Flows = GetNegFlows(bill.Flows)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	return err
}

func (k Keeper) UpdateBucketInfoAndCharge(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, newPaymentAddr string, newReadQuota uint64) error {
	if bucketInfo.PaymentAddress != newPaymentAddr && bucketInfo.ChargedReadQuota != newReadQuota {
		return fmt.Errorf("payment address and read quota can not be changed at the same time")
	}
	err := k.ChargeViaBucketChange(ctx, bucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.PaymentAddress = newPaymentAddr
		bi.ChargedReadQuota = newReadQuota
		return nil
	})
	return err
}

func (k Keeper) LockStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	amount, err := k.GetObjectLockFee(ctx, bucketInfo.PrimarySpAddress, objectInfo.CreateAt, objectInfo.PayloadSize)
	if ctx.IsCheckTx() {
		_ = ctx.EventManager().EmitTypedEvents(&types.EventFeePreview{
			Account:        paymentAddr.String(),
			FeePreviewType: types.FEE_PREVIEW_TYPE_PRELOCKED_FEE,
			Amount:         amount,
		})
	}
	if err != nil {
		return fmt.Errorf("get object store fee rate failed: %w", err)
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(paymentAddr).WithLockBalanceChange(amount)
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
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	change := types.NewDefaultStreamRecordChangeWithAddr(paymentAddr).WithLockBalanceChange(lockedBalance.Neg())
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

	return k.ChargeStoreFee(ctx, bucketInfo, objectInfo)
}

func (k Keeper) ChargeStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	if err != nil {
		return fmt.Errorf("get charge size error: %w", err)
	}
	return k.ChargeViaBucketChange(ctx, bucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.BillingInfo.TotalChargeSize += chargeSize
		secondarySpObjectsSize := bi.BillingInfo.SecondarySpObjectsSize
		for _, sp := range objectInfo.SecondarySpAddresses {
			secondarySpObjectsSize = append(secondarySpObjectsSize, storagetypes.SecondarySpObjectsSize{
				SpAddress:       sp,
				TotalChargeSize: chargeSize,
			})
		}
		bi.BillingInfo.SecondarySpObjectsSize = MergeSecondarySpObjectsSize(secondarySpObjectsSize)
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
	userFlows.From = sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	if bucketInfo.BillingInfo.TotalChargeSize == 0 && bucketInfo.ChargedReadQuota == 0 {
		return userFlows, nil
	}
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: bucketInfo.PrimarySpAddress,
		PriceTime: bucketInfo.BillingInfo.PriceTime,
	})
	if err != nil {
		return userFlows, fmt.Errorf("get storage price failed: %w", err)
	}
	primarySpFundingAddr, err := k.GetFundingAddressBySpAddr(ctx, sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress))
	if err != nil {
		return userFlows, fmt.Errorf("get funding address by sp address failed: %w, sp: %s", err, bucketInfo.PrimarySpAddress)
	}
	totalUserOutRate := sdkmath.ZeroInt()
	readFlowRate := price.ReadPrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	primaryStoreFlowRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.BillingInfo.TotalChargeSize)).TruncateInt()
	primarySpRate := readFlowRate.Add(primaryStoreFlowRate)
	if primarySpRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: primarySpFundingAddr,
			Rate:      primarySpRate,
		})
		totalUserOutRate = totalUserOutRate.Add(primarySpRate)
	}
	for _, spObjectsSize := range bucketInfo.BillingInfo.SecondarySpObjectsSize {
		rate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(spObjectsSize.TotalChargeSize)).TruncateInt()
		if rate.IsZero() {
			continue
		}
		spFundingAddr, err := k.GetFundingAddressBySpAddr(ctx, sdk.MustAccAddressFromHex(spObjectsSize.SpAddress))
		if err != nil {
			return userFlows, fmt.Errorf("get funding address by sp address failed: %w, sp: %s", err, spObjectsSize.SpAddress)
		}
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: spFundingAddr,
			Rate:      rate,
		})
		totalUserOutRate = totalUserOutRate.Add(rate)
	}
	params := k.paymentKeeper.GetParams(ctx)
	validatorTaxRate := params.ValidatorTaxRate.MulInt(totalUserOutRate).TruncateInt()
	if validatorTaxRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: types.ValidatorTaxPoolAddress.String(),
			Rate:      validatorTaxRate,
		})
	}
	return userFlows, nil
}

func (k Keeper) ChargeAccordingToBillChange(ctx sdk.Context, prev, current types.UserFlows) error {
	prev.Flows = GetNegFlows(prev.Flows)
	err := k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{prev, current})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %w", err)
	}
	return nil
}

func (k Keeper) ChargeDeleteObject(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	if err != nil {
		return fmt.Errorf("get charge size error: %w", err)
	}
	return k.ChargeViaBucketChange(ctx, bucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.BillingInfo.TotalChargeSize -= chargeSize
		var toBeSub []storagetypes.SecondarySpObjectsSize
		for _, sp := range objectInfo.SecondarySpAddresses {
			toBeSub = append(toBeSub, storagetypes.SecondarySpObjectsSize{
				SpAddress:       sp,
				TotalChargeSize: chargeSize,
			})
		}
		bi.BillingInfo.SecondarySpObjectsSize = SubSecondarySpObjectsSize(bi.BillingInfo.SecondarySpObjectsSize, toBeSub)
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

func MergeSecondarySpObjectsSize(list []storagetypes.SecondarySpObjectsSize) []storagetypes.SecondarySpObjectsSize {
	if len(list) <= 1 {
		return list
	}
	helperMap := make(map[string]uint64)
	for _, spObjectsSize := range list {
		helperMap[spObjectsSize.SpAddress] += spObjectsSize.TotalChargeSize
	}
	res := make([]storagetypes.SecondarySpObjectsSize, 0, len(helperMap))
	for sp, size := range helperMap {
		if size == 0 {
			continue
		}
		res = append(res, storagetypes.SecondarySpObjectsSize{
			SpAddress:       sp,
			TotalChargeSize: size,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].SpAddress < res[j].SpAddress
	})
	return res
}

func SubSecondarySpObjectsSize(prev []storagetypes.SecondarySpObjectsSize, toBeSub []storagetypes.SecondarySpObjectsSize) []storagetypes.SecondarySpObjectsSize {
	if len(toBeSub) == 0 {
		return prev
	}
	helperMap := make(map[string]uint64)
	// merge prev
	for _, spObjectsSize := range prev {
		helperMap[spObjectsSize.SpAddress] += spObjectsSize.TotalChargeSize
	}
	// sub toBeSub
	for _, spObjectsSize := range toBeSub {
		helperMap[spObjectsSize.SpAddress] -= spObjectsSize.TotalChargeSize
	}
	// merge the result
	res := make([]storagetypes.SecondarySpObjectsSize, 0, len(helperMap))
	for sp, size := range helperMap {
		if size == 0 {
			continue
		}
		res = append(res, storagetypes.SecondarySpObjectsSize{
			SpAddress:       sp,
			TotalChargeSize: size,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].SpAddress < res[j].SpAddress
	})
	return res
}

func (k Keeper) GetObjectLockFee(ctx sdk.Context, primarySpAddress string, priceTime int64, payloadSize uint64) (amount sdkmath.Int, err error) {
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: primarySpAddress,
		PriceTime: priceTime,
	})
	if err != nil {
		return amount, fmt.Errorf("get store price failed: %w", err)
	}
	chargeSize, err := k.GetChargeSize(ctx, payloadSize, priceTime)
	if err != nil {
		return amount, fmt.Errorf("get charge size error: %w", err)
	}
	rate := price.PrimaryStorePrice.Add(price.SecondaryStorePrice.MulInt64(storagetypes.SecondarySPNum)).MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	reserveTime := k.paymentKeeper.GetParams(ctx).ReserveTime
	amount = rate.Mul(sdkmath.NewIntFromUint64(reserveTime))
	return amount, nil
}

func (k Keeper) GetChargeSize(ctx sdk.Context, payloadSize uint64, ts int64) (size uint64, err error) {
	params, err := k.GetParamsWithTs(ctx, ts)
	if err != nil {
		return size, fmt.Errorf("get charge size failed, ts:%d, error: %w", ts, err)
	}
	minChargeSize := params.MinChargeSize
	if payloadSize < minChargeSize {
		return minChargeSize, nil
	} else {
		return payloadSize, nil
	}
}
