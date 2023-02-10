package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetReadPrice priceTime is kept to retrieve the price of ReadPacket at historical time
func (k Keeper) GetReadPrice(ctx sdk.Context, readQuota storagetypes.ReadQuota, _priceTime int64) (sdkmath.Int, error) {
	return k.GetReadPriceV0(readQuota)
}

func (k Keeper) GetReadPriceV0(readPacket storagetypes.ReadQuota) (price sdkmath.Int, err error) {
	switch readPacket {
	case storagetypes.READ_QUOTA_FREE:
		price = sdkmath.NewInt(0)
	case storagetypes.READ_QUOTA_1G:
		price = sdkmath.NewInt(2)
	case storagetypes.READ_QUOTA_10G:
		price = sdkmath.NewInt(4)
	default:
		err = fmt.Errorf("invalid read packet level: %d", readPacket)
	}
	return
}

func (k Keeper) GetStorePrice(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) types.StorePrice {
	// price may change with objectInfo.CreateAt
	return k.GetStorePriceV0(ctx, bucketInfo, objectInfo)
}

func (k Keeper) GetStorePriceV0(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, objectInfo *storagetypes.ObjectInfo) types.StorePrice {
	// A simple mock price: 4 per byte per second for primary SP and 1 per byte per second for 6 secondary SPs
	storePrice := types.StorePrice{
		UserPayRate: sdkmath.NewInt(100),
	}
	if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_INIT {
		// TODO: WARNING HARDCODE Here. Need refine according to the params of storage module
		if len(objectInfo.SecondarySpAddresses) != 6 {
			panic("there should be 6 secondary sps")
		}
		storePrice.Flows = []storagetypes.OutFlowInUSD{
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
