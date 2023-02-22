package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) GetReadPrice(ctx sdk.Context, spAddr string, readQuota uint64, priceTime int64) (sdkmath.Int, error) {
	if readQuota == 0 {
		return sdkmath.NewInt(0), nil
	}
	spStoragePrice, err := k.spKeeper.GetSpStoragePriceByTime(ctx, spAddr, priceTime)
	if err != nil {
		return sdkmath.NewInt(0), fmt.Errorf("get sp storage price failed: %w", err)
	}
	rate := spStoragePrice.ReadPrice.Mul(sdkmath.NewIntFromUint64(readQuota)).QuoRaw(types.PriceUnit)
	return rate, nil
}

func (k Keeper) GetStorePrice(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) (price types.StorePrice, err error) {
	primarySpPrice, err := k.spKeeper.GetSpStoragePriceByTime(ctx, bucketInfo.PrimarySpAddress, bucketInfo.PaymentPriceTime)
	if err != nil {
		return types.StorePrice{}, fmt.Errorf("get sp storage price failed: %w", err)
	}
	secondarySpStorePrice, err := k.spKeeper.GetSecondarySpStorePriceByTime(ctx, bucketInfo.PaymentPriceTime)
	if err != nil {
		return types.StorePrice{}, fmt.Errorf("get secondary sp store price failed: %w", err)
	}
	storePrice := types.StorePrice{
		UserPayRate: primarySpPrice.StorePrice.Add(secondarySpStorePrice.StorePrice.MulRaw(6)),
	}
	if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_INIT {
		if len(objectInfo.SecondarySpAddresses) != 6 {
			panic("there should be 6 secondary sps")
		}
		storePrice.Flows = []types.OutFlow{
			{ToAddress: bucketInfo.PrimarySpAddress, Rate: primarySpPrice.StorePrice},
			{ToAddress: objectInfo.SecondarySpAddresses[0], Rate: secondarySpStorePrice.StorePrice},
			{ToAddress: objectInfo.SecondarySpAddresses[1], Rate: secondarySpStorePrice.StorePrice},
			{ToAddress: objectInfo.SecondarySpAddresses[2], Rate: secondarySpStorePrice.StorePrice},
			{ToAddress: objectInfo.SecondarySpAddresses[3], Rate: secondarySpStorePrice.StorePrice},
			{ToAddress: objectInfo.SecondarySpAddresses[4], Rate: secondarySpStorePrice.StorePrice},
			{ToAddress: objectInfo.SecondarySpAddresses[5], Rate: secondarySpStorePrice.StorePrice},
		}
	}
	return storePrice, nil
}
