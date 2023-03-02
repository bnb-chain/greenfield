package keeper

import (
	"fmt"

	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetStoragePrice(ctx sdk.Context, params types.StoragePriceParams) (price types.StoragePrice, err error) {
	primarySpPrice, err := k.spKeeper.GetSpStoragePriceByTime(ctx, params.PrimarySp, params.PriceTime)
	if err != nil {
		return types.StoragePrice{}, fmt.Errorf("get sp storage price failed: %w", err)
	}
	secondarySpStorePrice, err := k.spKeeper.GetSecondarySpStorePriceByTime(ctx, params.PriceTime)
	if err != nil {
		return types.StoragePrice{}, fmt.Errorf("get secondary sp store price failed: %w", err)
	}
	storePrice := types.StoragePrice{
		PrimaryStorePrice:   primarySpPrice.StorePrice,
		SecondaryStorePrice: secondarySpStorePrice.StorePrice,
		ReadPrice:           primarySpPrice.ReadPrice,
	}
	return storePrice, nil
}
