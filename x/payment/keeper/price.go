package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetReadPrice priceTime is kept to retrieve the price of ReadPacket at historical time
func (k Keeper) GetReadPrice(ctx sdk.Context, readPacket types.ReadPacket, priceTime int64) (sdkmath.Int, error) {
	priceInUSD, err := GetReadPriceV0(readPacket)
	if err != nil {
		return sdkmath.ZeroInt(), fmt.Errorf("get read price failed: %w", err)
	}
	if priceInUSD.IsZero() {
		return priceInUSD, nil
	}
	priceNum, pricePrecision, err := k.GetBNBPriceByTime(ctx, priceTime)
	if err != nil {
		return sdkmath.ZeroInt(), fmt.Errorf("get bnb price failed: %w", err)
	}
	priceInBNB := priceInUSD.Mul(pricePrecision).Quo(priceNum)
	return priceInBNB, nil
}

func GetReadPriceV0(readPacket types.ReadPacket) (price sdkmath.Int, err error) {
	switch readPacket {
	case types.ReadPacketFree:
		price = sdkmath.NewInt(0)
		break
	case types.ReadPacket1GB:
		price = sdkmath.NewInt(1)
		break
	case types.ReadPacket10GB:
		price = sdkmath.NewInt(10)
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
