package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// PaymentAccountCountKeyPrefix is the prefix to retrieve all PaymentAccountCount
	PaymentAccountCountKeyPrefix = "PaymentAccountCount/value/"
)

// PaymentAccountCountKey returns the store key to retrieve a PaymentAccountCount from the index fields
func PaymentAccountCountKey(
	owner string,
) []byte {
	var key []byte

	ownerBytes := []byte(owner)
	key = append(key, ownerBytes...)
	key = append(key, []byte("/")...)

	return key
}
