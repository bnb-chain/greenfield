package keeper

import (
	"fmt"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SettleAndDistributeGVGFamily(ctx sdk.Context, spID uint32, family *types.GlobalVirtualGroupFamily) error {
	paymentAddress := sdk.MustAccAddressFromHex(family.GetVirtualPaymentAddress())
	streamRecord, found := k.paymentKeeper.GetStreamRecord(ctx, paymentAddress)
	if !found {
		return nil
	}

	totalBalance := streamRecord.StaticBalance
	diff := ctx.BlockTime().Unix() - streamRecord.CrudTimestamp
	totalBalance = totalBalance.Add(streamRecord.NetflowRate.MulRaw(diff))
	if !totalBalance.IsPositive() {
		return nil
	}

	change := paymenttypes.NewDefaultStreamRecordChangeWithAddr(paymentAddress).WithStaticBalanceChange(totalBalance.Neg())
	err := k.paymentKeeper.UpdateStreamRecord(ctx, streamRecord, change, false)
	if err != nil {
		return fmt.Errorf("fail to settle gvg family: %d, err: %s", family.Id, err.Error())
	}
	k.paymentKeeper.SetStreamRecord(ctx, streamRecord)

	sp, found := k.spKeeper.GetStorageProvider(ctx, spID)
	if !found {
		return fmt.Errorf("fail to find primary sp: %d", spID)
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx,
		paymenttypes.ModuleName,
		sdk.MustAccAddressFromHex(sp.FundingAddress),
		sdk.NewCoins(sdk.NewCoin(k.paymentKeeper.GetParams(ctx).FeeDenom, totalBalance)))
	if err != nil {
		return fmt.Errorf("fail to send coins: %s %s", paymentAddress, sp.FundingAddress)
	}

	return nil
}

func (k Keeper) SettleAndDistributeGVG(ctx sdk.Context, gvg *types.GlobalVirtualGroup) error {
	paymentAddress := sdk.MustAccAddressFromHex(gvg.GetVirtualPaymentAddress())
	streamRecord, found := k.paymentKeeper.GetStreamRecord(ctx, paymentAddress)
	if !found {
		return nil
	}

	totalBalance := streamRecord.StaticBalance
	diff := ctx.BlockTime().Unix() - streamRecord.CrudTimestamp
	totalBalance = totalBalance.Add(streamRecord.NetflowRate.MulRaw(diff))
	if !totalBalance.IsPositive() {
		return nil
	}

	change := paymenttypes.NewDefaultStreamRecordChangeWithAddr(paymentAddress).WithStaticBalanceChange(totalBalance)
	err := k.paymentKeeper.UpdateStreamRecord(ctx, streamRecord, change, true)
	if err != nil {
		return fmt.Errorf("fail to settle gvg gvg: %d, err: %s", gvg.Id, err.Error())
	}
	k.paymentKeeper.SetStreamRecord(ctx, streamRecord)

	amount := totalBalance.QuoRaw(int64(len(gvg.SecondarySpIds)))
	if amount.IsPositive() {
		coins := sdk.NewCoins(sdk.NewCoin(k.paymentKeeper.GetParams(ctx).FeeDenom, amount))
		for _, spID := range gvg.SecondarySpIds {
			sp, found := k.spKeeper.GetStorageProvider(ctx, spID)
			if !found {
				return fmt.Errorf("fail to find primary sp: %d", spID)
			}
			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx,
				paymenttypes.ModuleName,
				sdk.MustAccAddressFromHex(sp.FundingAddress),
				coins)
			if err != nil {
				return fmt.Errorf("fail to send coins: %s %s", paymentAddress, sp.FundingAddress)
			}
		}
	}

	return nil
}
