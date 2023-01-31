package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// BnbPriceKeyPrefix is the prefix to retrieve all BnbPrice
	BnbPriceKeyPrefix = "BnbPrice/value/"
)

// BnbPriceKey returns the store key to retrieve a BnbPrice from the index fields
func BnbPriceKey(
	time int64,
) []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(time))
	return timeBytes
}
