package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetReadPrice priceTime is kept to retrieve the price of ReadPacket at historical time
func (k Keeper) GetReadPrice(ctx sdk.Context, readPacket types.ReadPacket, _priceTime int64) (sdkmath.Int, error) {
	return k.GetReadPriceV0(readPacket)
}

func (k Keeper) GetReadPriceV0(readPacket types.ReadPacket) (price sdkmath.Int, err error) {
	switch readPacket {
	case types.READ_PACKET_FREE:
		price = sdkmath.NewInt(0)
	case types.READ_PACKET_1GB:
		price = sdkmath.NewInt(2)
	case types.READ_PACKET_10GB:
		price = sdkmath.NewInt(4)
	default:
		err = fmt.Errorf("invalid read packet level: %d", readPacket)
	}
	return
}

func (k Keeper) GetStorePrice(ctx sdk.Context, bucketMeta *types.MockBucketMeta, objectInfo *types.MockObjectInfo) types.StorePrice {
	// price may change with objectInfo.CreateAt
	return k.GetStorePriceV0(ctx, bucketMeta, objectInfo)
}

func (k Keeper) GetStorePriceV0(ctx sdk.Context, bucketMeta *types.MockBucketMeta, objectInfo *types.MockObjectInfo) types.StorePrice {
	// A simple mock price: 4 per byte per second for primary SP and 1 per byte per second for 6 secondary SPs
	storePrice := types.StorePrice{
		UserPayRate: sdkmath.NewInt(100),
	}
	if objectInfo.ObjectState != types.OBJECT_STATE_INIT {
		if len(objectInfo.SecondarySPs) != 6 {
			panic("there should be 6 secondary sps")
		}
		storePrice.Flows = []types.OutFlowInUSD{
			{SpAddress: bucketMeta.SpAddress, Rate: sdkmath.NewInt(40)},
			{SpAddress: objectInfo.SecondarySPs[0].Id, Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySPs[1].Id, Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySPs[2].Id, Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySPs[3].Id, Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySPs[4].Id, Rate: sdkmath.NewInt(10)},
			{SpAddress: objectInfo.SecondarySPs[5].Id, Rate: sdkmath.NewInt(10)},
		}
	}
	return storePrice
}
