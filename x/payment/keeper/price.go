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
	spStoragePrice, err := k.GetSpStoragePriceByTime(ctx, spAddr, priceTime)
	if err != nil {
		return sdkmath.NewInt(0), fmt.Errorf("get sp storage price failed: %w", err)
	}
	rate := spStoragePrice.ReadQuotaPrice.Mul(sdkmath.NewIntFromUint64(readQuota)).QuoRaw(types.PriceUint)
	return rate, nil
}

func (k Keeper) GetStorePrice(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) types.StorePrice {
	// A simple mock price: 4 per byte per second for primary SP and 1 per byte per second for 6 secondary SPs
	storePrice := types.StorePrice{
		UserPayRate: sdkmath.NewInt(100),
	}
	if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_INIT {
		// TODO: WARNING HARDCODE Here. Need refine according to the params of storage module
		if len(objectInfo.SecondarySpAddresses) != 6 {
			panic("there should be 6 secondary sps")
		}
		storePrice.Flows = []types.OutFlow{
			{SpAddress: bucketInfo.PrimarySpAddress, Rate: sdkmath.NewInt(40)},
			{SpAddress: objectInfo.SecondarySpAddresses[0], Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySpAddresses[1], Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySpAddresses[2], Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySpAddresses[3], Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySpAddresses[4], Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySpAddresses[5], Rate: sdkmath.NewInt(10)},
		}
	}
	return storePrice
}
