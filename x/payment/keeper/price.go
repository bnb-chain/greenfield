package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) GetStoragePrice(ctx sdk.Context, params types.StoragePriceParams) (price types.StoragePrice, err error) {
	primarySp, err := sdk.AccAddressFromHexUnsafe(params.PrimarySp)
	if err != nil {
		return types.StoragePrice{}, fmt.Errorf("invalid primary sp address: %s", params.PrimarySp)
	}
	primarySpPrice, err := k.spKeeper.GetSpStoragePriceByTime(ctx, primarySp, params.PriceTime)
	if err != nil {
		return types.StoragePrice{}, fmt.Errorf("get sp [%s] storage price @[%d] failed: %w", params.PrimarySp, params.PriceTime, err)
	}
	secondarySpStorePrice, err := k.spKeeper.GetSecondarySpStorePriceByTime(ctx, params.PriceTime)
	if err != nil {
		return types.StoragePrice{}, fmt.Errorf("get secondary sp store price failed: %w, price time: %d", err, params.PriceTime)
	}
	storePrice := types.StoragePrice{
		PrimaryStorePrice:   primarySpPrice.StorePrice,
		SecondaryStorePrice: secondarySpStorePrice.StorePrice,
		ReadPrice:           primarySpPrice.ReadPrice,
	}
	return storePrice, nil
}
