package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// BnbPricePriceKeyPrefix is the prefix to retrieve all BnbPricePrice
	BnbPricePriceKeyPrefix = "BnbPricePrice/value/"
)

// BnbPricePriceKey returns the store key to retrieve a BnbPricePrice from the index fields
func BnbPricePriceKey(
	time int64,
) []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(time))
	return timeBytes
}
