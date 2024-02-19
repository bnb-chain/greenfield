package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (k Keeper) SettleAndDistributeGVGFamily(ctx sdk.Context, sp *sptypes.StorageProvider, family *types.GlobalVirtualGroupFamily) error {
	paymentAddress := sdk.MustAccAddressFromHex(family.GetVirtualPaymentAddress())
	totalBalance, err := k.paymentKeeper.QueryDynamicBalance(ctx, paymentAddress)
	if err != nil {
		return fmt.Errorf("fail to query balance: %s, err: %s", paymentAddress.String(), err.Error())
	}
	if !totalBalance.IsPositive() {
		return nil
	}

	err = k.paymentKeeper.Withdraw(ctx, paymentAddress, sdk.MustAccAddressFromHex(sp.FundingAddress), totalBalance)
	if err != nil {
		return fmt.Errorf("fail to send coins: %s %s", paymentAddress, sp.FundingAddress)
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventSettleGlobalVirtualGroupFamily{
		Id:               family.Id,
		SpId:             sp.Id,
		SpFundingAddress: sp.FundingAddress,
		Amount:           totalBalance,
	})
	if err != nil {
		ctx.Logger().Error("fail to send event for settlement", "vfg", family.Id, "err", err)
	}

	return nil
}

func (k Keeper) SettleAndDistributeGVG(ctx sdk.Context, gvg *types.GlobalVirtualGroup) error {
	paymentAddress := sdk.MustAccAddressFromHex(gvg.GetVirtualPaymentAddress())
	totalBalance, err := k.paymentKeeper.QueryDynamicBalance(ctx, paymentAddress)
	if err != nil {
		return fmt.Errorf("fail to query balance: %s, err: %s", paymentAddress.String(), err.Error())
	}

	amount := totalBalance.QuoRaw(int64(len(gvg.SecondarySpIds)))
	if !amount.IsPositive() {
		return nil
	}

	fundingAddresses := make([]string, 0)
	for _, spID := range gvg.SecondarySpIds {
		sp, found := k.spKeeper.GetStorageProvider(ctx, spID)
		if !found {
			return fmt.Errorf("fail to find secondary sp: %d", spID)
		}
		err = k.paymentKeeper.Withdraw(ctx, paymentAddress, sdk.MustAccAddressFromHex(sp.FundingAddress), amount)
		if err != nil {
			return fmt.Errorf("fail to send coins: %s %s", paymentAddress, sp.FundingAddress)
		}

		fundingAddresses = append(fundingAddresses, sp.FundingAddress)
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventSettleGlobalVirtualGroup{
		Id:                 gvg.Id,
		SpIds:              gvg.SecondarySpIds,
		SpFundingAddresses: fundingAddresses,
		Amount:             amount,
	})
	if err != nil {
		ctx.Logger().Error("fail to send event for settlement", "gvg", gvg.Id, "err", err)
	}

	return nil
}
