package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
)

func SubmitBNBPrice(priceTime int64, price sdkmath.Int) {

}

// GetBNBPrice return the price of BNB at priceTime
// price = num / precision
// e.g. num = 27740000000, precision = 100000000, price = 27740000000 / 100000000 = 277.4
func GetBNBPrice(_priceTime int64) (num, precision sdkmath.Int) {
	return sdkmath.NewInt(27740000000), sdkmath.NewInt(100000000)
}

// GetReadPrice priceTime is kept to retrieve the price of ReadPacket at historical time
func GetReadPrice(readPacket types.ReadPacket, priceTime int64) (sdkmath.Int, error) {
	priceInUSD, err := GetReadPriceV0(readPacket)
	if err != nil {
		return sdkmath.ZeroInt(), fmt.Errorf("get read price failed: %w", err)
	}
	if priceInUSD.IsZero() {
		return priceInUSD, nil
	}
	priceNum, pricePrecision := GetBNBPrice(priceTime)
	priceInBNB := priceInUSD.Mul(pricePrecision).Quo(priceNum)
	return priceInBNB, nil
}

func GetReadPriceV0(readPacket types.ReadPacket) (price sdkmath.Int, err error) {
	switch readPacket {
	case types.ReadPacketLevelFree:
		price = sdkmath.NewInt(0)
		break
	case types.ReadPacketLevel1GB:
		price = sdkmath.NewInt(1e17)
		break
	case types.ReadPacketLevel10GB:
		price = sdkmath.NewInt(1e18)
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
