package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) ChargeBucketReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	if bucketInfo.ChargedReadQuota == 0 {
		return nil
	}
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketReadBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("charge bucket read fee failed, get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		ctx.Logger().Error("charge initial read fee failed", "err", err.Error())
		return err
	}
	return nil
}

func (k Keeper) UnChargeBucketReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	if internalBucketInfo.TotalChargeSize > 0 {
		return fmt.Errorf("unexpected total store charge size: %d", internalBucketInfo.TotalChargeSize)
	}

	bill, err := k.GetBucketReadBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("uncharge bucket read fee failed, get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	if len(bill.Flows) == 0 {
		return nil
	}
	bill.Flows = getNegFlows(bill.Flows)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		ctx.Logger().Error("uncharge bucket read fee failed", "err", err.Error())
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
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: bucketInfo.PrimarySpId,
		PriceTime: internalBucketInfo.PriceTime,
	})
	if err != nil {
		return userFlows, fmt.Errorf("get storage price failed: %w", err)
	}

	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.PrimarySpId, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return userFlows, fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
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
		return userFlows, fmt.Errorf("failed to get validator tax rate: %w, time: %d", err, internalBucketInfo.PriceTime)
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
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	amount, err := k.GetObjectLockFee(ctx, bucketInfo.PrimarySpId, objectInfo.CreateAt, objectInfo.PayloadSize)
	if err != nil {
		return fmt.Errorf("get object store fee rate failed: %w", err)
	}
	if ctx.IsCheckTx() {
		_ = ctx.EventManager().EmitTypedEvents(&types.EventFeePreview{
			Account:        paymentAddr.String(),
			FeePreviewType: types.FEE_PREVIEW_TYPE_PRELOCKED_FEE,
			Amount:         amount,
		})
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

// UnlockObjectStoreFee unlock store fee if the object is deleted in INIT state
func (k Keeper) UnlockObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	lockedBalance, err := k.GetObjectLockFee(ctx, bucketInfo.PrimarySpId, objectInfo.CreateAt, objectInfo.PayloadSize)
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

func (k Keeper) UnlockAndChargeObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	// unlock store fee
	err := k.UnlockObjectStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return fmt.Errorf("unlock store fee failed: %w", err)
	}

	return k.ChargeObjectStoreFee(ctx, bucketInfo, internalBucketInfo, objectInfo)
}

func (k Keeper) IsPriceChanged(ctx sdk.Context, primarySpId uint32, priceTime int64) (bool, error) {
	prePrice, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: primarySpId,
		PriceTime: priceTime,
	})
	if err != nil {
		return false, fmt.Errorf("get storage price failed: %w", err)
	}
	currentPrice, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: primarySpId,
		PriceTime: ctx.BlockTime().Unix(),
	})
	if err != nil {
		return false, fmt.Errorf("get storage price failed: %w", err)
	}

	return !(prePrice.ReadPrice.Equal(currentPrice.ReadPrice) &&
		prePrice.PrimaryStorePrice.Equal(currentPrice.PrimaryStorePrice) &&
		prePrice.SecondaryStorePrice.Equal(currentPrice.SecondaryStorePrice)), nil
}

func (k Keeper) ChargeObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetObjectChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	if err != nil {
		return fmt.Errorf("get charge size error: %w", err)
	}

	priceChanged, err := k.IsPriceChanged(ctx, bucketInfo.PrimarySpId, internalBucketInfo.PriceTime)
	if err != nil {
		return fmt.Errorf("check whether price changed error: %w", err)
	}

	if !priceChanged {
		err := k.ChargeViaObjectChange(ctx, bucketInfo, internalBucketInfo, objectInfo, chargeSize, false)
		if err != nil {
			return fmt.Errorf("apply object store bill error: %w", err)
		}
		return nil
	}

	return k.ChargeViaBucketChange(ctx, bucketInfo, internalBucketInfo, func(bi *storagetypes.BucketInfo, ibi *storagetypes.InternalBucketInfo) error {
		ibi.TotalChargeSize += chargeSize
		for _, lvg := range ibi.LocalVirtualGroups {
			if lvg.Id == objectInfo.LocalVirtualGroupId {
				lvg.TotalChargeSize += lvg.TotalChargeSize + chargeSize
				break
			}
		}
		return nil
	})
}

func (k Keeper) UnChargeObjectStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetObjectChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	if err != nil {
		return fmt.Errorf("get charge size error: %w", err)
	}

	priceChanged, err := k.IsPriceChanged(ctx, bucketInfo.PrimarySpId, internalBucketInfo.PriceTime)
	if err != nil {
		return fmt.Errorf("check whether price changed error: %w", err)
	}

	oldInternalBucketInfo := &storagetypes.InternalBucketInfo{
		PriceTime:          internalBucketInfo.PriceTime,
		TotalChargeSize:    internalBucketInfo.TotalChargeSize,
		LocalVirtualGroups: internalBucketInfo.LocalVirtualGroups,
	}

	if !priceChanged {
		err = k.ChargeViaObjectChange(ctx, bucketInfo, internalBucketInfo, objectInfo, chargeSize, true)
		if err != nil {
			return fmt.Errorf("apply object store bill error: %w", err)
		}
	} else {
		err = k.ChargeViaBucketChange(ctx, bucketInfo, internalBucketInfo, func(bi *storagetypes.BucketInfo, ibi *storagetypes.InternalBucketInfo) error {
			ibi.TotalChargeSize -= chargeSize
			for _, lvg := range ibi.LocalVirtualGroups {
				if lvg.Id == objectInfo.LocalVirtualGroupId {
					lvg.TotalChargeSize -= chargeSize
					break
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("apply object store bill error: %w", err)
		}
	}

	blockTime := ctx.BlockTime().Unix()
	versionParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, oldInternalBucketInfo.PriceTime)
	if err != nil {
		return fmt.Errorf("failed to get versioned params: %w", err)
	}
	timeToPay := objectInfo.CreateAt + int64(versionParams.ReserveTime) - blockTime
	if timeToPay > 0 { // store less than reserve time
		err = k.ChargeObjectStoreFeeForEarlyDeletion(ctx, bucketInfo, oldInternalBucketInfo, objectInfo, chargeSize, timeToPay)
		forced, _ := ctx.Value(types.ForceUpdateStreamRecordKey).(bool) // force update in end block
		if !forced && err != nil {
			return fmt.Errorf("fail to pay for early deletion, error: %w", err)
		}
	}
	return nil
}

func (k Keeper) ChargeObjectStoreFeeForEarlyDeletion(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo,
	chargeSize uint64, timeToPay int64) error {
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: bucketInfo.PrimarySpId,
		PriceTime: internalBucketInfo.PriceTime,
	})
	if err != nil {
		return fmt.Errorf("get storage price failed: %w", err)
	}

	// primary sp total rate
	primaryTotalFlowRate := sdk.ZeroInt()

	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.PrimarySpId, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}

	primaryRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	if primaryRate.IsPositive() {
		primaryTotalFlowRate = primaryRate
		err = k.paymentKeeper.Withdraw(ctx, paymentAddr, sdk.MustAccAddressFromHex(gvgFamily.VirtualPaymentAddress),
			primaryTotalFlowRate.MulRaw(timeToPay))
		if err != nil {
			return fmt.Errorf("fail to pay GVG family: %s", err)
		}
	}

	// secondary sp total rate
	secondaryTotalFlowRate := sdk.ZeroInt()

	var lvg *storagetypes.LocalVirtualGroup
	for _, l := range internalBucketInfo.LocalVirtualGroups {
		if l.Id == objectInfo.LocalVirtualGroupId {
			lvg = l
			break
		}
	}

	gvg, found := k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)
	if !found {
		return fmt.Errorf("get GVG failed: %d, %s", lvg.GlobalVirtualGroupId, lvg.String())
	}

	secondaryRate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	secondaryRate = secondaryRate.MulRaw(int64(len(gvg.SecondarySpIds)))
	if secondaryRate.IsPositive() {
		secondaryTotalFlowRate = secondaryTotalFlowRate.Add(secondaryRate)
		err = k.paymentKeeper.Withdraw(ctx, paymentAddr, sdk.MustAccAddressFromHex(gvg.VirtualPaymentAddress),
			secondaryTotalFlowRate.MulRaw(timeToPay))
		if err != nil {
			return fmt.Errorf("fail to pay GVG: %s", err)
		}
	}

	// validator tax rate
	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return fmt.Errorf("failed to get validator tax rate: %w, time: %d", err, internalBucketInfo.PriceTime)
	}
	validatorTaxRate := versionedParams.ValidatorTaxRate.MulInt(primaryTotalFlowRate.Add(secondaryTotalFlowRate)).TruncateInt()
	if validatorTaxRate.IsPositive() {
		err = k.paymentKeeper.Withdraw(ctx, paymentAddr, types.ValidatorTaxPoolAddress,
			validatorTaxRate.MulRaw(timeToPay))
		if err != nil {
			return fmt.Errorf("fail to pay validator: %s", err)
		}
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
		return errors.Wrapf(err, "change bucket internal info failed")
	}

	// calculate new bill
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	newBill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get new bucket bill failed: %w", err)
	}

	// charge according to bill change
	err = k.ApplyBillChanges(ctx, prevBill, newBill)
	if err != nil {
		ctx.Logger().Error("charge via bucket change failed", "err", err.Error())
		return err
	}
	return nil
}

func (k Keeper) ChargeViaObjectChange(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo, chargeSize uint64, delete bool) error {
	userFlows := types.UserFlows{
		From:  sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress),
		Flows: make([]types.OutFlow, 0),
	}

	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: bucketInfo.PrimarySpId,
		PriceTime: internalBucketInfo.PriceTime,
	})
	if err != nil {
		return fmt.Errorf("get storage price failed: %w", err)
	}

	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.PrimarySpId, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}

	// primary sp total rate
	primaryTotalFlowRate := sdk.ZeroInt()

	// secondary sp total rate
	secondaryTotalFlowRate := sdk.ZeroInt()

	var lvg *storagetypes.LocalVirtualGroup
	for _, l := range internalBucketInfo.LocalVirtualGroups {
		if l.Id == objectInfo.LocalVirtualGroupId {
			lvg = l
			break
		}
	}

	// primary sp
	primaryRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	if primaryRate.IsPositive() {
		primaryTotalFlowRate = primaryTotalFlowRate.Add(primaryRate)
	}

	//secondary sp
	gvg, found := k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)
	if !found {
		return fmt.Errorf("get GVG failed: %d, %s", lvg.GlobalVirtualGroupId, lvg.String())
	}

	secondaryRate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	secondaryRate = secondaryRate.MulRaw(int64(len(gvg.SecondarySpIds)))
	if secondaryRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: gvg.VirtualPaymentAddress,
			Rate:      secondaryRate,
		})
		secondaryTotalFlowRate = secondaryTotalFlowRate.Add(secondaryRate)
	}

	if primaryTotalFlowRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: gvgFamily.VirtualPaymentAddress,
			Rate:      primaryTotalFlowRate,
		})
	}

	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return fmt.Errorf("failed to get validator tax rate: %w, time: %d", err, internalBucketInfo.PriceTime)
	}
	validatorTaxRate := versionedParams.ValidatorTaxRate.MulInt(primaryTotalFlowRate.Add(secondaryTotalFlowRate)).TruncateInt()
	if validatorTaxRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: types.ValidatorTaxPoolAddress.String(),
			Rate:      validatorTaxRate,
		})
	}

	if !delete {
		internalBucketInfo.TotalChargeSize = internalBucketInfo.TotalChargeSize + chargeSize
		lvg.TotalChargeSize = lvg.TotalChargeSize + chargeSize
	} else {
		internalBucketInfo.TotalChargeSize = internalBucketInfo.TotalChargeSize - chargeSize
		lvg.TotalChargeSize = lvg.TotalChargeSize - chargeSize

		userFlows.Flows = getNegFlows(userFlows.Flows)
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	if err != nil {
		ctx.Logger().Error("charge object store fee failed", "err", err.Error())
		return err
	}

	return nil
}

func (k Keeper) GetBucketReadStoreBill(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) (userFlows types.UserFlows, err error) {
	userFlows.From = sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	if internalBucketInfo.TotalChargeSize == 0 && bucketInfo.ChargedReadQuota == 0 {
		return userFlows, nil
	}
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: bucketInfo.PrimarySpId,
		PriceTime: internalBucketInfo.PriceTime,
	})
	if err != nil {
		return userFlows, fmt.Errorf("get storage price failed: %w", err)
	}

	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.PrimarySpId, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return userFlows, fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}

	// primary sp total rate
	primaryTotalFlowRate := price.ReadPrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()

	// secondary sp total rate
	secondaryTotalFlowRate := sdk.ZeroInt()

	for _, lvg := range internalBucketInfo.LocalVirtualGroups {
		// primary sp
		primaryRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(lvg.TotalChargeSize)).TruncateInt()
		if primaryRate.IsPositive() {
			primaryTotalFlowRate = primaryTotalFlowRate.Add(primaryRate)
		}

		//secondary sp
		gvg, found := k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)
		if !found {
			return userFlows, fmt.Errorf("get GVG failed: %d, %s", lvg.GlobalVirtualGroupId, lvg.String())
		}

		secondaryRate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(lvg.TotalChargeSize)).TruncateInt()
		secondaryRate = secondaryRate.MulRaw(int64(len(gvg.SecondarySpIds)))
		if secondaryRate.IsPositive() {
			userFlows.Flows = append(userFlows.Flows, types.OutFlow{
				ToAddress: gvg.VirtualPaymentAddress,
				Rate:      secondaryRate,
			})
			secondaryTotalFlowRate = secondaryTotalFlowRate.Add(secondaryRate)
		}
	}

	if primaryTotalFlowRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: gvgFamily.VirtualPaymentAddress,
			Rate:      primaryTotalFlowRate,
		})
	}

	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, internalBucketInfo.PriceTime)
	if err != nil {
		return userFlows, fmt.Errorf("failed to get validator tax rate: %w, time: %d", err, internalBucketInfo.PriceTime)
	}
	validatorTaxRate := versionedParams.ValidatorTaxRate.MulInt(primaryTotalFlowRate.Add(secondaryTotalFlowRate)).TruncateInt()
	if validatorTaxRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: types.ValidatorTaxPoolAddress.String(),
			Rate:      validatorTaxRate,
		})
	}

	return userFlows, nil
}

func (k Keeper) UnChargeBucketReadStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	bill.Flows = getNegFlows(bill.Flows)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %w", err)
	}
	return nil
}

func (k Keeper) ChargeBucketReadStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketReadStoreBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %w", err)
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

func (k Keeper) GetObjectLockFee(ctx sdk.Context, primarySpId uint32, priceTime int64, payloadSize uint64) (amount sdkmath.Int, err error) {
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: primarySpId,
		PriceTime: priceTime,
	})
	if err != nil {
		return amount, fmt.Errorf("get store price failed: %w", err)
	}
	chargeSize, err := k.GetObjectChargeSize(ctx, payloadSize, priceTime)
	if err != nil {
		return amount, fmt.Errorf("get charge size error: %w", err)
	}
	secondarySPNum := int64(k.GetExpectSecondarySPNumForECObject(ctx, priceTime))
	rate := price.PrimaryStorePrice.Add(price.SecondaryStorePrice.MulInt64(secondarySPNum)).MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, priceTime)
	if err != nil {
		return amount, fmt.Errorf("get versioned reserve time error: %w", err)
	}
	amount = rate.Mul(sdkmath.NewIntFromUint64(versionedParams.ReserveTime))
	return amount, nil
}

func (k Keeper) GetObjectChargeSize(ctx sdk.Context, payloadSize uint64, ts int64) (size uint64, err error) {
	params, err := k.GetVersionedParamsWithTs(ctx, ts)
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
