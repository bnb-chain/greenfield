package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (k Keeper) SettleAndDistributeGVGFamily(ctx sdk.Context, spID uint32, family *types.GlobalVirtualGroupFamily) error {
	paymentAddress := sdk.MustAccAddressFromHex(family.GetVirtualPaymentAddress())
	totalBalance, err := k.paymentKeeper.QueryDynamicBalance(ctx, paymentAddress)
	if err != nil {
		return fmt.Errorf("fail to query balance: %s, err: %s", paymentAddress.String(), err.Error())
	}
	if !totalBalance.IsPositive() {
		return nil
	}

	sp, found := k.spKeeper.GetStorageProvider(ctx, spID)
	if !found {
		return fmt.Errorf("fail to find primary sp: %d", spID)
	}
	err = k.paymentKeeper.Withdraw(ctx, paymentAddress, sdk.MustAccAddressFromHex(sp.FundingAddress), totalBalance)
	if err != nil {
		return fmt.Errorf("fail to send coins: %s %s", paymentAddress, sp.FundingAddress)
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
	for _, spID := range gvg.SecondarySpIds {
		sp, found := k.spKeeper.GetStorageProvider(ctx, spID)
		if !found {
			return fmt.Errorf("fail to find secondary sp: %d", spID)
		}
		err = k.paymentKeeper.Withdraw(ctx, paymentAddress, sdk.MustAccAddressFromHex(sp.FundingAddress), amount)
		if err != nil {
			return fmt.Errorf("fail to send coins: %s %s", paymentAddress, sp.FundingAddress)
		}
	}
	return nil
}
