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

func (k Keeper) ChargeInitialReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	if bucketInfo.ChargedReadQuota == 0 {
		return nil
	}
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("charge initial read fee failed, get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	if err != nil {
		ctx.Logger().Error("charge initial read fee failed", "err", err.Error())
		return err
	}
	return nil
}

func (k Keeper) ChargeDeleteBucket(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {
	bill, err := k.GetBucketBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return err
	}
	if len(bill.Flows) == 0 {
		return nil
	}
	//should only remain at most 2 flows: charged_read_quota fee to gvg family and tax to validator pool
	if len(bill.Flows) > 2 {
		panic(fmt.Sprintf("unexpected left flow number: %d", len(bill.Flows)))
	}
	bill.Flows = GetNegFlows(bill.Flows)
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{bill})
	return err
}

func (k Keeper) UpdateBucketInfoAndCharge(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, newPaymentAddr string, newReadQuota uint64) error {
	if bucketInfo.PaymentAddress != newPaymentAddr && bucketInfo.ChargedReadQuota != newReadQuota {
		return fmt.Errorf("payment address and read quota can not be changed at the same time")
	}
	err := k.ChargeViaBucketChange(ctx, bucketInfo, internalBucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.PaymentAddress = newPaymentAddr
		bi.ChargedReadQuota = newReadQuota
		return nil
	})
	return err
}

func (k Keeper) LockStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	paymentAddr := sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	primarySp, found := k.spKeeper.GetStorageProvider(ctx, bucketInfo.PrimarySpId)
	if !found {
		return fmt.Errorf("get storage provider failed: %d", bucketInfo.PrimarySpId)
	}
	amount, err := k.GetObjectLockFee(ctx, primarySp.OperatorAddress, objectInfo.CreateAt, objectInfo.PayloadSize)
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
	primarySp, found := k.spKeeper.GetStorageProvider(ctx, bucketInfo.PrimarySpId)
	if !found {
		return fmt.Errorf("get storage provider failed: %d", bucketInfo.PrimarySpId)
	}
	lockedBalance, err := k.GetObjectLockFee(ctx, primarySp.OperatorAddress, objectInfo.CreateAt, objectInfo.PayloadSize)
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

func (k Keeper) UnlockAndChargeStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	// unlock store fee
	err := k.UnlockStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return fmt.Errorf("unlock store fee failed: %w", err)
	}

	return k.ChargeStoreFee(ctx, bucketInfo, internalBucketInfo, objectInfo)
}

func (k Keeper) ChargeStoreFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	if err != nil {
		return fmt.Errorf("get charge size error: %w", err)
	}
	return k.ChargeViaBucketChange(ctx, bucketInfo, internalBucketInfo, func(bi *storagetypes.BucketInfo) error {
		internalBucketInfo.TotalChargeSize += chargeSize
		toMerge := internalBucketInfo.LocalVirtualGroups
		toMerge = append(toMerge, &storagetypes.LocalVirtualGroup{
			Id:              objectInfo.LocalVirtualGroupId,
			TotalChargeSize: chargeSize,
		})
		internalBucketInfo.LocalVirtualGroups = MergeLvgChargeSize(toMerge)
		return nil
	})
}

func (k Keeper) ChargeDeleteObject(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	if err != nil {
		return fmt.Errorf("get charge size error: %w", err)
	}
	return k.ChargeViaBucketChange(ctx, bucketInfo, internalBucketInfo, func(bi *storagetypes.BucketInfo) error {
		internalBucketInfo.TotalChargeSize -= chargeSize
		toSub := []*storagetypes.LocalVirtualGroup{
			{
				Id:              objectInfo.LocalVirtualGroupId,
				TotalChargeSize: chargeSize},
		}
		internalBucketInfo.LocalVirtualGroups = SubLvgChargeSize(internalBucketInfo.LocalVirtualGroups, toSub)
		return nil
	})
}

func (k Keeper) ChargeViaBucketChange(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo,
	changeFunc func(bucketInfo *storagetypes.BucketInfo) error) error {
	// get previous bill
	prevBill, err := k.GetBucketBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("charge via bucket change failed, get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	// change bucket billing info
	if err = changeFunc(bucketInfo); err != nil {
		return errors.Wrapf(err, "change bucket billing info failed")
	}
	internalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	// calculate new bill
	newBill, err := k.GetBucketBill(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return fmt.Errorf("get new bucket bill failed: %w", err)
	}

	// charge according to bill change
	err = k.ChargeAccordingToBillChange(ctx, prevBill, newBill)
	if err != nil {
		ctx.Logger().Error("charge via bucket change failed", "err", err.Error())
		return err
	}
	return nil
}

func (k Keeper) GetBucketBill(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo, settle ...bool) (userFlows types.UserFlows, err error) {
	doSettle := settle != nil && settle[0]
	userFlows.From = sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	if internalBucketInfo.TotalChargeSize == 0 && bucketInfo.ChargedReadQuota == 0 {
		return userFlows, nil
	}
	primarySp, found := k.spKeeper.GetStorageProvider(ctx, bucketInfo.PrimarySpId)
	if !found {
		return userFlows, fmt.Errorf("get storage provider failed: %d", bucketInfo.PrimarySpId)
	}
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: primarySp.OperatorAddress,
		PriceTime: internalBucketInfo.PriceTime,
	})
	if err != nil {
		return userFlows, fmt.Errorf("get storage price failed: %w", err)
	}

	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.PrimarySpId, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return userFlows, fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}

	if doSettle {
		err := k.virtualGroupKeeper.SettleAndDistributeGVGFamily(ctx, primarySp.Id, gvgFamily)
		if err != nil {
			return userFlows, fmt.Errorf("settle GVG family failed: %d, err: %s", gvgFamily.Id, err.Error())
		}
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
			return userFlows, fmt.Errorf("get GVG failed: %d", lvg.GlobalVirtualGroupId)
		}

		if doSettle {
			err := k.virtualGroupKeeper.SettleAndDistributeGVG(ctx, gvg)
			if err != nil {
				return userFlows, fmt.Errorf("settle GVG failed: %d, err: %s", gvg.Id, err.Error())
			}
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

func (k Keeper) ChargeAccordingToBillChange(ctx sdk.Context, prevFlows, currentFlows types.UserFlows) error {
	prevFlows.Flows = GetNegFlows(prevFlows.Flows)
	err := k.paymentKeeper.ApplyUserFlowsList(ctx, []types.UserFlows{prevFlows, currentFlows})
	if err != nil {
		return fmt.Errorf("apply user flows list failed: %w", err)
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

func MergeLvgChargeSize(list []*storagetypes.LocalVirtualGroup) []*storagetypes.LocalVirtualGroup {
	if len(list) <= 1 {
		return list
	}
	helperMap := make(map[uint32]uint64)
	for _, objectsSize := range list {
		helperMap[objectsSize.Id] += objectsSize.TotalChargeSize
	}
	res := make([]*storagetypes.LocalVirtualGroup, 0, len(helperMap))
	for id, size := range helperMap {
		if size == 0 {
			continue
		}
		res = append(res, &storagetypes.LocalVirtualGroup{
			Id:              id,
			TotalChargeSize: size,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Id < res[j].Id
	})
	return res
}

func SubLvgChargeSize(prev []*storagetypes.LocalVirtualGroup, toBeSub []*storagetypes.LocalVirtualGroup) []*storagetypes.LocalVirtualGroup {
	if len(toBeSub) == 0 {
		return prev
	}
	helperMap := make(map[uint32]uint64)
	// merge prev
	for _, objectsSize := range prev {
		helperMap[objectsSize.Id] += objectsSize.TotalChargeSize
	}
	// sub toBeSub
	for _, objectsSize := range toBeSub {
		helperMap[objectsSize.Id] -= objectsSize.TotalChargeSize
	}
	// merge the result
	res := make([]*storagetypes.LocalVirtualGroup, 0, len(helperMap))
	for id, size := range helperMap {
		if size == 0 {
			continue
		}
		res = append(res, &storagetypes.LocalVirtualGroup{
			Id:              id,
			TotalChargeSize: size,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Id < res[j].Id
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
	secondarySPNum := int64(k.GetExpectSecondarySPNumForECObject(ctx, priceTime))
	rate := price.PrimaryStorePrice.Add(price.SecondaryStorePrice.MulInt64(secondarySPNum)).MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	versionedParams, err := k.paymentKeeper.GetVersionedParamsWithTs(ctx, priceTime)
	if err != nil {
		return amount, fmt.Errorf("get versioned reserve time error: %w", err)
	}
	amount = rate.Mul(sdkmath.NewIntFromUint64(versionedParams.ReserveTime))
	return amount, nil
}

func (k Keeper) GetChargeSize(ctx sdk.Context, payloadSize uint64, ts int64) (size uint64, err error) {
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

func (k Keeper) UpdateLVGChargeSize(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo,
	internalBucketInfo *storagetypes.InternalBucketInfo) error {

	totalChargeSize := uint64(0)
	for _, lvg := range internalBucketInfo.LocalVirtualGroups {
		chargeSize, err := k.GetChargeSize(ctx, lvg.StoredSize, internalBucketInfo.PriceTime)
		if err != nil {
			return err
		}
		totalChargeSize = totalChargeSize + chargeSize
		lvg.TotalChargeSize = chargeSize
	}
	internalBucketInfo.TotalChargeSize = totalChargeSize
	return nil
}

func (k Keeper) ChargeBucketMigration(ctx sdk.Context, oldBucketInfo, newBucketInfo *storagetypes.BucketInfo,
	oldInternalBucketInfo, newInternalBucketInfo *storagetypes.InternalBucketInfo) error {
	// settle and get previous bill
	prevBill, err := k.GetBucketBill(ctx, oldBucketInfo, oldInternalBucketInfo, true)
	if err != nil {
		return fmt.Errorf("settle and get bucket bill failed, bucket: %s, err: %s", oldBucketInfo.BucketName, err.Error())
	}

	// update billing info
	newInternalBucketInfo.PriceTime = ctx.BlockTime().Unix()
	err = k.UpdateLVGChargeSize(ctx, newBucketInfo, newInternalBucketInfo)
	if err != nil {
		return fmt.Errorf("fail to update lvg charge size, bucket: %s, err: %s", oldBucketInfo.BucketName, err.Error())
	}

	// calculate new bill
	newBill, err := k.GetBucketBill(ctx, newBucketInfo, newInternalBucketInfo)
	if err != nil {
		return fmt.Errorf("get new bucket bill failed: %w", err)
	}

	// charge according to bill change
	err = k.ChargeAccordingToBillChange(ctx, prevBill, newBill)
	if err != nil {
		ctx.Logger().Error("charge via bucket change failed", "err", err.Error())
		return err
	}
	return nil
}
