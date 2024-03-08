package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	vgtypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (k Keeper) ChargeBucketReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	if bucketInfo.ChargedReadQuota == 0 {
		return nil
	}
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketReadBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("charge bucket read fee failed, get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		ctx.Logger().Error("charge initial read fee failed", "bucket", bucketInfo.BucketName, "err", err.Error())
		return err
	}
	return nil
}

func (k Keeper) UnChargeBucketReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	if internalBucketInfo.TotalChargeSize > 0 {
		return fmt.Errorf("unexpected total store charge size: %s, %d", bucketInfo.BucketName, internalBucketInfo.TotalChargeSize)
	}

	bill, err := k.GetBucketReadBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("uncharge bucket read fee failed, get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	if len(bill.Flows) == 0 {
		return nil
	}
	bill.Flows = getNegFlows(bill.Flows)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		ctx.Logger().Error("uncharge bucket read fee failed", "bucket", bucketInfo.BucketName, "err", err.Error())
		return err
	}
	return nil
}

func (k Keeper) GetBucketReadBill(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) (userFlows types.UserFlows, err error) {
	userFlows.From = sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	if bucketInfo.ChargedReadQuota == 0 {
		return userFlows, nil
	}
	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return userFlows, fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}

	price, err := k.spKeeper.GetGlobalSpStorePriceByTime(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return userFlows, fmt.Errorf("get storage price failed: %d %w", internalBucketInfo.PriceTime, err)
	}

	// primary sp total rate
	primaryTotalFlowRate := price.ReadPrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()

	if primaryTotalFlowRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: gvgFamily.VirtualPaymentAddress,
			Rate:      primaryTotalFlowRate,
		})
	}

	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return userFlows, fmt.Errorf("failed to get validator tax rate: %d %w", internalBucketInfo.PriceTime, err)
	}
	validatorTaxRate := versionedParams.ValidatorTaxRate.MulInt(primaryTotalFlowRate).TruncateInt()
	if validatorTaxRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: types.ValidatorTaxPoolAddress.String(),
			Rate:      validatorTaxRate,
		})
	}

	return userFlows, nil
}

func (k Keeper) UpdateBucketInfoAndCharge(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, newPaymentAddr string, newReadQuota uint64) error {
	if bucketInfo.PaymentAddress != newPaymentAddr && bucketInfo.ChargedReadQuota != newReadQuota {
		return fmt.Errorf("payment address and read quota can not be changed at the same time")
	}
	err := k.ChargeViaBucketChange(ctx, bucketInfo, internalBucketInfo, func(bi *storagetypes.BucketInfo, ibi *storagetypes.InternalBucketInfo) error {
		bi.PaymentAddress = newPaymentAddr
		bi.ChargedReadQuota = newReadQuota
		return nil
	})
	return err
}

func (k Keeper) LockObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	return k.lockObjectStoreFee(ctx, bucketInfo, objectInfo.GetLatestUpdatedTime(), objectInfo.PayloadSize, objectInfo.ObjectName)
}

func (k Keeper) LockShadowObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ShadowObjectInfo, objectName string) error {
	return k.lockObjectStoreFee(ctx, bucketInfo, objectInfo.UpdatedAt, objectInfo.PayloadSize, objectName)
}

func (k Keeper) lockObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, timestamp int64, payloadSize uint64, objectName string) error {
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	amount, err := k.GetObjectLockFee(ctx, timestamp, payloadSize)
	if err != nil {
		return fmt.Errorf("get object store fee rate failed: %s %s %w", bucketInfo.BucketName, objectName, err)
	}
	if ctx.IsCheckTx() {
		_ = ctx.EventManager().EmitTypedEvents(&types.EventFeePreview{
			Account:        paymentAddr.String(),
			FeePreviewType: types.FEE_PREVIEW_TYPE_PRELOCKED_FEE,
			Amount:         amount,
		})
	}
	fmt.Println("bucket", bucketInfo.BucketName, "objectName", objectName, "payloadSize", payloadSize, "timestamp", timestamp, "amount", amount)
	change := types.NewDefaultStreamRecordChangeWithAddr(paymentAddr).WithLockBalanceChange(amount)
	streamRecord, err := k.paymentKeeper.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return fmt.Errorf("update stream record failed: %s %s %w", bucketInfo.BucketName, objectName, err)
	}
	if streamRecord.StaticBalance.IsNegative() {
		return fmt.Errorf("static balance is not enough for %s %s, lacks %s", bucketInfo.BucketName, objectName, streamRecord.StaticBalance.Neg().String())
	}
	return nil
}

// UnlockObjectStoreFee unlock store fee if the object is deleted in INIT state
func (k Keeper) UnlockObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	lockedBalance, err := k.GetObjectLockFee(ctx, objectInfo.GetLatestUpdatedTime(), objectInfo.PayloadSize)
	if err != nil {
		return fmt.Errorf("get object store fee rate failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	change := types.NewDefaultStreamRecordChangeWithAddr(paymentAddr).WithLockBalanceChange(lockedBalance.Neg())
	_, err = k.paymentKeeper.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return fmt.Errorf("update stream record failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}
	return nil
}

// UnlockShadowObjectStoreFee unlock store fee if the object is deleted in INIT state
func (k Keeper) UnlockShadowObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ShadowObjectInfo) error {
	lockedBalance, err := k.GetObjectLockFee(ctx, objectInfo.GetUpdatedAt(), objectInfo.PayloadSize)
	if err != nil {
		return fmt.Errorf("get shadow object store fee rate failed, objectID: %s %w", objectInfo.Id.String(), err)
	}
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	change := types.NewDefaultStreamRecordChangeWithAddr(paymentAddr).WithLockBalanceChange(lockedBalance.Neg())
	_, err = k.paymentKeeper.UpdateStreamRecordByAddr(ctx, change)
	if err != nil {
		return fmt.Errorf("update stream record failed, objectID: %s %w", objectInfo.Id.String(), err)
	}
	return nil
}

func (k Keeper) UnlockAndChargeObjectStoreFee(ctx sdk.Context, primarySpId uint32, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	// unlock store fee
	err := k.UnlockObjectStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return fmt.Errorf("unlock store fee failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}

	return k.ChargeObjectStoreFee(ctx, primarySpId, bucketInfo, internalBucketInfo, objectInfo)
}

func (k Keeper) IsPriceChanged(ctx sdk.Context, primarySpId uint32, priceTime int64) (bool, *sptypes.GlobalSpStorePrice, sdk.Dec, *sptypes.GlobalSpStorePrice, sdk.Dec, error) {
	prePrice, err := k.spKeeper.GetGlobalSpStorePriceByTime(ctx, priceTime)
	if err != nil {
		return false, nil, sdk.ZeroDec(), nil, sdk.ZeroDec(), err
	}

	currentPrice, err := k.spKeeper.GetGlobalSpStorePriceByTime(ctx, ctx.BlockTime().Unix())
	if err != nil {
		return false, nil, sdk.ZeroDec(), nil, sdk.ZeroDec(), err
	}

	preParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, priceTime)
	if err != nil {
		return false, nil, sdk.ZeroDec(), nil, sdk.ZeroDec(), err
	}

	currentParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, ctx.BlockTime().Unix())
	if err != nil {
		return false, nil, sdk.ZeroDec(), nil, sdk.ZeroDec(), err
	}

	return !(prePrice.ReadPrice.Equal(currentPrice.ReadPrice) &&
			prePrice.PrimaryStorePrice.Equal(currentPrice.PrimaryStorePrice) &&
			prePrice.SecondaryStorePrice.Equal(currentPrice.SecondaryStorePrice) &&
			preParams.ValidatorTaxRate.Equal(currentParams.ValidatorTaxRate)),
		&prePrice, preParams.ValidatorTaxRate, &currentPrice, currentParams.ValidatorTaxRate, nil
}

func (k Keeper) ChargeObjectStoreFee(ctx sdk.Context, primarySpId uint32, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetObjectChargeSize(ctx, objectInfo.PayloadSize, objectInfo.GetLatestUpdatedTime())
	if err != nil {
		return fmt.Errorf("get charge size failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}

	priceChanged, _, _, _, _, err := k.IsPriceChanged(ctx, primarySpId, internalBucketInfo.PriceTime)
	if err != nil {
		return fmt.Errorf("check whether price changed failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}

	if !priceChanged {
		_, err := k.ChargeViaObjectChange(ctx, bucketInfo, internalBucketInfo, objectInfo, chargeSize, false)
		if err != nil {
			return fmt.Errorf("apply object store bill failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
		}
		return nil
	}

	return k.ChargeViaBucketChange(ctx, bucketInfo, internalBucketInfo, func(bi *storagetypes.BucketInfo, ibi *storagetypes.InternalBucketInfo) error {
		ibi.TotalChargeSize += chargeSize
		for _, lvg := range ibi.LocalVirtualGroups {
			if lvg.Id == objectInfo.LocalVirtualGroupId {
				lvg.TotalChargeSize = lvg.TotalChargeSize + chargeSize
				break
			}
		}
		return nil
	})
}

func (k Keeper) UnChargeObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetObjectChargeSize(ctx, objectInfo.PayloadSize, objectInfo.GetLatestUpdatedTime())
	if err != nil {
		return fmt.Errorf("get charge size failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}

	userFlows, err := k.ChargeViaObjectChange(ctx, bucketInfo, internalBucketInfo, objectInfo, chargeSize, true)
	if err != nil {
		return fmt.Errorf("apply object store bill failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}

	blockTime := ctx.BlockTime().Unix()
	versionParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return fmt.Errorf("failed to get versioned params: %w", err)
	}
	timeToPay := objectInfo.GetLatestUpdatedTime() + int64(versionParams.ReserveTime) - blockTime
	if timeToPay > 0 { // store less than reserve time
		err = k.ChargeObjectStoreFeeForEarlyDeletion(ctx, userFlows, bucketInfo, objectInfo, timeToPay)
		forced, _ := ctx.Value(types.ForceUpdateStreamRecordKey).(bool) // force update in end block
		if !forced && err != nil {
			return fmt.Errorf("pay for early deletion failed: %s %s %w", bucketInfo.BucketName, objectInfo.ObjectName, err)
		}
	}
	return nil
}

func (k Keeper) ChargeObjectStoreFeeForEarlyDeletion(ctx sdk.Context, userFlows []types.OutFlow, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo, timeToPay int64) error {
	totalStaticBalanceChange := sdkmath.NewInt(0)
	for _, flow := range userFlows {
		staticBalanceChange := flow.Rate.Abs().MulRaw(timeToPay)
		_, err := k.paymentKeeper.UpdateStreamRecordByAddr(ctx,
			types.NewDefaultStreamRecordChangeWithAddr(sdk.MustAccAddressFromHex(flow.ToAddress)).WithStaticBalanceChange(staticBalanceChange))
		if err != nil {
			return fmt.Errorf("pay address %s failed: %s %s %s", sdk.MustAccAddressFromHex(flow.ToAddress), bucketInfo.BucketName, objectInfo.ObjectName, err)
		}
		totalStaticBalanceChange = totalStaticBalanceChange.Add(staticBalanceChange)
	}
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	_, err := k.paymentKeeper.UpdateStreamRecordByAddr(ctx, types.NewDefaultStreamRecordChangeWithAddr(paymentAddr).
		WithStaticBalanceChange(totalStaticBalanceChange.Neg()))
	if err != nil {
		return fmt.Errorf("subtracting from payment account failed: %s %s %s", bucketInfo.BucketName, objectInfo.ObjectName, err)
	}
	return nil
}

func (k Keeper) ChargeViaBucketChange(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo,
	changeFunc func(bi *storagetypes.BucketInfo, ibi *storagetypes.InternalBucketInfo) error) error {

	// get previous bill
	prevBill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("charge via bucket change failed, get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	// change bucket internal info
	if err = changeFunc(bucketInfo, internalBucketInfo); err != nil {
		return errors.Wrapf(err, "change bucket internal info failed: %s", bucketInfo.BucketName)
	}
	// calculate new bill
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	newBill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get new bucket bill failed: %s %w", bucketInfo.BucketName, err)
	}

	// charge according to bill change
	err = k.ApplyBillChanges(ctx, prevBill, newBill)
	if err != nil {
		ctx.Logger().Error("charge via bucket change failed", "bucket", bucketInfo.BucketName, "err", err.Error())
		return err
	}
	return nil
}

func (k Keeper) ChargeViaObjectChange(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo, chargeSize uint64, delete bool) ([]types.OutFlow, error) {
	userFlows := types.UserFlows{
		From:  sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress),
		Flows: make([]types.OutFlow, 0),
	}
	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return nil, fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}

	price, err := k.spKeeper.GetGlobalSpStorePriceByTime(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return nil, fmt.Errorf("get storage price failed: %d %w", internalBucketInfo.PriceTime, err)
	}

	var lvg *storagetypes.LocalVirtualGroup
	for _, l := range internalBucketInfo.LocalVirtualGroups {
		if l.Id == objectInfo.LocalVirtualGroupId {
			lvg = l
			break
		}
	}

	gvg, found := k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)
	if !found {
		return nil, fmt.Errorf("get GVG failed: %d, %s", lvg.GlobalVirtualGroupId, lvg.String())
	}

	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator tax rate: %w, time: %d", err, internalBucketInfo.PriceTime)
	}

	preOutFlows := k.calculateLVGStoreBill(ctx, price, versionedParams, gvgFamily, gvg, lvg)
	var newOutFlows []types.OutFlow
	if !delete { // seal object
		internalBucketInfo.TotalChargeSize = internalBucketInfo.TotalChargeSize + chargeSize
		lvg.TotalChargeSize = lvg.TotalChargeSize + chargeSize
		newOutFlows = k.calculateLVGStoreBill(ctx, price, versionedParams, gvgFamily, gvg, lvg)
	} else { // delete object
		internalBucketInfo.TotalChargeSize = internalBucketInfo.TotalChargeSize - chargeSize
		lvg.TotalChargeSize = lvg.TotalChargeSize - chargeSize
		newOutFlows = k.calculateLVGStoreBill(ctx, price, versionedParams, gvgFamily, gvg, lvg)
	}

	userFlows.Flows = append(userFlows.Flows, getNegFlows(preOutFlows)...)
	userFlows.Flows = append(userFlows.Flows, newOutFlows...)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	if err != nil {
		ctx.Logger().Error("charge object store fee failed", "bucket", bucketInfo.BucketName,
			"object", objectInfo.ObjectName, "err", err.Error())
		return nil, err
	}
	// merge outflows for early deletion usage
	return k.paymentKeeper.MergeOutFlows(userFlows.Flows), nil
}

func (k Keeper) calculateLVGStoreBill(ctx sdk.Context, price sptypes.GlobalSpStorePrice, params types.VersionedParams,
	gvgFamily *vgtypes.GlobalVirtualGroupFamily, gvg *vgtypes.GlobalVirtualGroup, lvg *storagetypes.LocalVirtualGroup) []types.OutFlow {
	outFlows := make([]types.OutFlow, 0)

	// primary sp
	primaryStoreFlowRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(lvg.TotalChargeSize)).TruncateInt()
	if primaryStoreFlowRate.IsPositive() {
		outFlows = append(outFlows, types.OutFlow{
			ToAddress: gvgFamily.VirtualPaymentAddress,
			Rate:      primaryStoreFlowRate,
		})
	}

	//secondary sp
	secondaryStoreFlowRate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(lvg.TotalChargeSize)).TruncateInt()
	secondaryStoreFlowRate = secondaryStoreFlowRate.MulRaw(int64(len(gvg.SecondarySpIds)))
	if secondaryStoreFlowRate.IsPositive() {
		outFlows = append(outFlows, types.OutFlow{
			ToAddress: gvg.VirtualPaymentAddress,
			Rate:      secondaryStoreFlowRate,
		})
	}

	validatorTaxStoreFlowRate := params.ValidatorTaxRate.MulInt(primaryStoreFlowRate.Add(secondaryStoreFlowRate)).TruncateInt()
	if validatorTaxStoreFlowRate.IsPositive() {
		outFlows = append(outFlows, types.OutFlow{
			ToAddress: types.ValidatorTaxPoolAddress.String(),
			Rate:      validatorTaxStoreFlowRate,
		})
	}

	return outFlows
}

func (k Keeper) GetBucketReadStoreBill(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) (userFlows types.UserFlows, err error) {
	userFlows.From = sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)

	if internalBucketInfo.TotalChargeSize == 0 && bucketInfo.ChargedReadQuota == 0 {
		return userFlows, nil
	}

	// calculate read fee & store fee separately, for precision
	// calculate read fee
	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return userFlows, fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}

	price, err := k.spKeeper.GetGlobalSpStorePriceByTime(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return userFlows, fmt.Errorf("get storage price failed: %d %w", internalBucketInfo.PriceTime, err)
	}

	primaryReadFlowRate := price.ReadPrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	if primaryReadFlowRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: gvgFamily.VirtualPaymentAddress,
			Rate:      primaryReadFlowRate,
		})
	}

	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return userFlows, fmt.Errorf("failed to get validator tax rate: %d %w", internalBucketInfo.PriceTime, err)
	}
	validatorTaxReadFlowRate := versionedParams.ValidatorTaxRate.MulInt(primaryReadFlowRate).TruncateInt()
	if validatorTaxReadFlowRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: types.ValidatorTaxPoolAddress.String(),
			Rate:      validatorTaxReadFlowRate,
		})
	}

	// calculate store fee
	// be noted, here we split the fee calculation for each lvg, to make sure each lvg's calculation is precise
	for _, lvg := range internalBucketInfo.LocalVirtualGroups {
		//secondary sp
		gvg, found := k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)
		if !found {
			return userFlows, fmt.Errorf("get GVG failed: %d, %s", lvg.GlobalVirtualGroupId, lvg.String())
		}
		outFlows := k.calculateLVGStoreBill(ctx, price, versionedParams, gvgFamily, gvg, lvg)
		userFlows.Flows = append(userFlows.Flows, outFlows...)
	}

	return userFlows, nil
}

func (k Keeper) UnChargeBucketReadStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	bill.Flows = getNegFlows(bill.Flows)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %s %w", bucketInfo.BucketName, err)
	}
	return nil
}

func (k Keeper) ChargeBucketReadStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed: %s %s", bucketInfo.BucketName, err.Error())
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %s %w", bucketInfo.BucketName, err)
	}
	return nil
}

func (k Keeper) ApplyBillChanges(ctx sdk.Context, prevFlows, currentFlows types.UserFlows) error {
	prevFlows.Flows = getNegFlows(prevFlows.Flows)
	err := k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{prevFlows, currentFlows})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %w", err)
	}
	return nil
}

func getNegFlows(flows []types.OutFlow) (negFlows []types.OutFlow) {
	negFlows = make([]types.OutFlow, len(flows))
	for i, flow := range flows {
		negFlows[i] = types.OutFlow{ToAddress: flow.ToAddress, Rate: flow.Rate.Neg()}
	}
	return negFlows
}

func (k Keeper) GetObjectLockFee(ctx sdk.Context, priceTime int64, payloadSize uint64) (amount sdkmath.Int, err error) {
	price, err := k.spKeeper.GetGlobalSpStorePriceByTime(ctx, priceTime)
	if err != nil {
		return amount, fmt.Errorf("get store price failed: %d %w", priceTime, err)
	}
	chargeSize, err := k.GetObjectChargeSize(ctx, payloadSize, priceTime)
	if err != nil {
		return amount, fmt.Errorf("get charge size failed: %d %w", priceTime, err)
	}

	primaryRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()

	secondarySPNum := int64(k.GetExpectSecondarySPNumForECObject(ctx, priceTime))
	secondaryRate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	secondaryRate = secondaryRate.MulRaw(int64(secondarySPNum))

	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, priceTime)
	if err != nil {
		return amount, fmt.Errorf("get versioned reserve time error: %w", err)
	}
	validatorTaxRate := versionedParams.ValidatorTaxRate.MulInt(primaryRate.Add(secondaryRate)).TruncateInt()

	rate := primaryRate.Add(secondaryRate).Add(validatorTaxRate) // should also lock for validator tax pool
	amount = rate.Mul(sdkmath.NewIntFromUint64(versionedParams.ReserveTime))
	return amount, nil
}

func (k Keeper) GetObjectChargeSize(ctx sdk.Context, payloadSize uint64, ts int64) (size uint64, err error) {
	params, err := k.GetVersionedParamsWithTs(ctx, ts)
	if err != nil {
		return size, fmt.Errorf("get charge size failed: %d %w", ts, err)
	}
	minChargeSize := params.MinChargeSize
	if payloadSize < minChargeSize {
		return minChargeSize, nil
	} else {
		return payloadSize, nil
	}
}
