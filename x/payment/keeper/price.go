package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
)

func SubmitBNBPrice(priceTime int64, price sdkmath.Int) {

}

func GetBNBPrice(priceTime int64) sdkmath.Int {
	return sdkmath.NewInt(1)
}

// GetReadPrice priceTime is kept to retrieve the price of ReadPacket at historical time
func GetReadPrice(readPacket types.ReadPacket, _priceTime int64) (sdkmath.Int, error) {
	return GetReadPriceV0(readPacket)
}

func GetReadPriceV0(readPacket types.ReadPacket) (price sdkmath.Int, err error) {
	switch readPacket {
	case types.ReadPacketLevelFree:
		price = sdkmath.NewInt(0)
		break
	case types.ReadPacketLevel1GB:
		price = sdkmath.NewInt(10)
		break
	case types.ReadPacketLevel10GB:
		price = sdkmath.NewInt(100)
		break
	default:
		err = fmt.Errorf("invalid read packet level: %d", readPacket)
	}
	return
}

//func GetStorePrice(objectMeta types.MockBucketMeta, priceTime int64) sdkmath.Int {
//
//}
//
//func GetStorePriceV0(size uint64, priceTime int64) sdkmath.Int {
//	return sdkmath.NewInt(100)
//}
