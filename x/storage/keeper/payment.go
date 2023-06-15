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

func (k Keeper) ChargeInitialReadFee(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) error {
	if bucketInfo.ChargedReadQuota == 0 {
		return nil
	}
	bucketInfo.BillingInfo.PriceTime = ctx.BlockTime().Unix()
	bill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return fmt.Errorf("charge initial read fee failed, get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, bill)
	if err != nil {
		// TODO: handle this
		ctx.Logger().Error("charge initial read fee failed", "err", err.Error())
	}
	return nil
}

func (k Keeper) ChargeDeleteBucket(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) error {
	bill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return err
	}
	if len(bill) == 0 {
		return nil
	}
	// TODO: fixme
	// should only remain at most 2 flows: charged_read_quota fee and tax
	//if len(bill.Flows) > 2 {
	//	panic(fmt.Sprintf("unexpected left flow number: %d", len(bill.Flows)))
	//}
	for _, f := range bill {
		f.Flows = GetNegFlows(f.Flows)
	}
	err = k.paymentKeeper.ApplyUserFlowsList(ctx, bill)
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
		lvgObjectsSize := bi.BillingInfo.LvgObjectsSize
		lvgObjectsSize = append(lvgObjectsSize, storagetypes.LVGObjectsSize{
			LvgId:           objectInfo.LocalVirtualGroupId,
			TotalChargeSize: chargeSize,
		})
		bi.BillingInfo.LvgObjectsSize = MergeSecondarySpObjectsSize(lvgObjectsSize)
		return nil
	})
}

func (k Keeper) ChargeDeleteObject(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) error {
	chargeSize, err := k.GetChargeSize(ctx, objectInfo.PayloadSize, objectInfo.CreateAt)
	if err != nil {
		return fmt.Errorf("get charge size error: %w", err)
	}
	return k.ChargeViaBucketChange(ctx, bucketInfo, func(bi *storagetypes.BucketInfo) error {
		bi.BillingInfo.TotalChargeSize -= chargeSize
		toBeSub := []storagetypes.LVGObjectsSize{
			{
				LvgId:           objectInfo.LocalVirtualGroupId,
				TotalChargeSize: chargeSize},
		}
		bi.BillingInfo.LvgObjectsSize = SubSecondarySpObjectsSize(bi.BillingInfo.LvgObjectsSize, toBeSub)
		return nil
	})
}

func (k Keeper) ChargeViaBucketChange(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, changeFunc func(bucketInfo *storagetypes.BucketInfo) error) error {
	// get previous bill
	prevBill, err := k.GetBucketBill(ctx, bucketInfo)
	if err != nil {
		return fmt.Errorf("charge via bucket change failed, get bucket bill failed, bucket: %s, err: %s", bucketInfo.BucketName, err.Error())
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
		// TODO: handle this
		ctx.Logger().Error("charge via bucket change failed", "err", err.Error())
	}
	return nil
}

func (k Keeper) GetBucketBill(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo) ([]types.UserFlows, error) {
	flows := []types.UserFlows{}

	if bucketInfo.BillingInfo.TotalChargeSize == 0 && bucketInfo.ChargedReadQuota == 0 {
		return flows, nil
	}
	primarySp, found := k.spKeeper.GetStorageProvider(ctx, bucketInfo.PrimarySpId)
	if !found {
		return flows, fmt.Errorf("get storage provider failed: %d", bucketInfo.PrimarySpId)
	}
	price, err := k.paymentKeeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: primarySp.OperatorAddress,
		PriceTime: bucketInfo.BillingInfo.PriceTime,
	})
	if err != nil {
		return flows, fmt.Errorf("get storage price failed: %w", err)
	}

	params := k.paymentKeeper.GetParams(ctx)

	gvgFamily, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.PrimarySpId, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return flows, fmt.Errorf("get GVG family failed: %d", bucketInfo.GlobalVirtualGroupFamilyId)
	}
	gvgFamilyFlows := types.UserFlows{From: sdk.MustAccAddressFromHex(gvgFamily.VirtualPaymentAddress)}
	userFlows := types.UserFlows{From: sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)}

	readFlowRate := price.ReadPrice.MulInt(sdkmath.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	// read flow: 1. payment account -> GVG family -> primary sp, 2. payment account -> GVG family -> validator tax pool
	if readFlowRate.IsPositive() {
		userFlows.Flows = append(userFlows.Flows, types.OutFlow{
			ToAddress: gvgFamily.VirtualPaymentAddress,
			Rate:      readFlowRate,
		})
		gvgFamilyFlows.Flows = append(gvgFamilyFlows.Flows, types.OutFlow{
			ToAddress: primarySp.FundingAddress,
			Rate:      readFlowRate,
		})

		validatorTaxRate := params.ValidatorTaxRate.MulInt(readFlowRate).TruncateInt()
		if validatorTaxRate.IsPositive() {
			userFlows.Flows = append(userFlows.Flows, types.OutFlow{
				ToAddress: gvgFamily.VirtualPaymentAddress,
				Rate:      validatorTaxRate,
			})
			gvgFamilyFlows.Flows = append(gvgFamilyFlows.Flows, types.OutFlow{
				ToAddress: types.ValidatorTaxPoolAddress.String(),
				Rate:      validatorTaxRate,
			})
		}
	}

	// store flows: 1. payment account -> GVG -> sps, 2. payment account -> GVG -> validator tax pool
	gvgFlows := make([]types.UserFlows, 0)
	storeFlowRate := sdkmath.ZeroInt()
	for _, lvgStoreSize := range bucketInfo.BillingInfo.LvgObjectsSize {
		lvg, found := k.virtualGroupKeeper.GetLVG(ctx, bucketInfo.Id, lvgStoreSize.LvgId)
		if !found {
			return flows, fmt.Errorf("get LVG failed: %d", lvgStoreSize.LvgId)
		}

		gvg, found := k.virtualGroupKeeper.GetGVG(ctx, primarySp.Id, lvg.GlobalVirtualGroupId)
		if !found {
			return flows, fmt.Errorf("get GVG failed: %d", lvg.GlobalVirtualGroupId)
		}
		gvgFlow := types.UserFlows{From: sdk.MustAccAddressFromHex(gvg.VirtualPaymentAddress)}

		// primary sp
		primaryStoreFlowRate := price.PrimaryStorePrice.MulInt(sdkmath.NewIntFromUint64(lvgStoreSize.TotalChargeSize)).TruncateInt()
		storeFlowRate = storeFlowRate.Add(primaryStoreFlowRate)
		if primaryStoreFlowRate.IsPositive() {
			userFlows.Flows = append(userFlows.Flows, types.OutFlow{
				ToAddress: gvg.VirtualPaymentAddress,
				Rate:      primaryStoreFlowRate,
			})
			gvgFlow.Flows = append(gvgFlow.Flows, types.OutFlow{
				ToAddress: primarySp.FundingAddress,
				Rate:      primaryStoreFlowRate,
			})
		}

		//secondary sp
		secondaryStoreFlowRate := price.SecondaryStorePrice.MulInt(sdkmath.NewIntFromUint64(lvgStoreSize.TotalChargeSize)).TruncateInt()
		if secondaryStoreFlowRate.IsPositive() {
			for _, id := range gvg.SecondarySpIds {
				sp, found := k.spKeeper.GetStorageProvider(ctx, id)
				if !found {
					return flows, fmt.Errorf("get sp failed: %d", id)
				}
				userFlows.Flows = append(userFlows.Flows, types.OutFlow{
					ToAddress: gvg.VirtualPaymentAddress,
					Rate:      secondaryStoreFlowRate,
				})
				gvgFlow.Flows = append(gvgFlow.Flows, types.OutFlow{
					ToAddress: sp.FundingAddress,
					Rate:      secondaryStoreFlowRate,
				})
				storeFlowRate = storeFlowRate.Add(secondaryStoreFlowRate)
			}
		}

		// validator tax pool
		validatorTaxRate := params.ValidatorTaxRate.MulInt(storeFlowRate).TruncateInt()
		if validatorTaxRate.IsPositive() {
			userFlows.Flows = append(userFlows.Flows, types.OutFlow{
				ToAddress: gvg.VirtualPaymentAddress,
				Rate:      validatorTaxRate,
			})
			gvgFlow.Flows = append(userFlows.Flows, types.OutFlow{
				ToAddress: types.ValidatorTaxPoolAddress.String(),
				Rate:      validatorTaxRate,
			})
		}
		gvgFlows = append(gvgFlows, gvgFlow)
	}

	flows = append(flows, userFlows, gvgFamilyFlows)
	flows = append(flows, gvgFlows...)

	return flows, nil
}

func (k Keeper) ChargeAccordingToBillChange(ctx sdk.Context, prevFlows, currentFlows []types.UserFlows) error {
	for _, pre := range prevFlows {
		pre.Flows = GetNegFlows(pre.Flows)
	}
	flows := make([]types.UserFlows, 0)
	flows = append(flows, prevFlows...)
	flows = append(flows, currentFlows...)
	err := k.paymentKeeper.ApplyUserFlowsList(ctx, flows)
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

func MergeSecondarySpObjectsSize(list []storagetypes.LVGObjectsSize) []storagetypes.LVGObjectsSize {
	if len(list) <= 1 {
		return list
	}
	helperMap := make(map[uint32]uint64)
	for _, objectsSize := range list {
		helperMap[objectsSize.LvgId] += objectsSize.TotalChargeSize
	}
	res := make([]storagetypes.LVGObjectsSize, 0, len(helperMap))
	for id, size := range helperMap {
		if size == 0 {
			continue
		}
		res = append(res, storagetypes.LVGObjectsSize{
			LvgId:           id,
			TotalChargeSize: size,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].LvgId < res[j].LvgId
	})
	return res
}

func SubSecondarySpObjectsSize(prev []storagetypes.LVGObjectsSize, toBeSub []storagetypes.LVGObjectsSize) []storagetypes.LVGObjectsSize {
	if len(toBeSub) == 0 {
		return prev
	}
	helperMap := make(map[uint32]uint64)
	// merge prev
	for _, objectsSize := range prev {
		helperMap[objectsSize.LvgId] += objectsSize.TotalChargeSize
	}
	// sub toBeSub
	for _, objectsSize := range toBeSub {
		helperMap[objectsSize.LvgId] -= objectsSize.TotalChargeSize
	}
	// merge the result
	res := make([]storagetypes.LVGObjectsSize, 0, len(helperMap))
	for sp, size := range helperMap {
		if size == 0 {
			continue
		}
		res = append(res, storagetypes.LVGObjectsSize{
			LvgId:           sp,
			TotalChargeSize: size,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].LvgId < res[j].LvgId
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
