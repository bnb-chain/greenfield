package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// PaymentAccountKeyPrefix is the prefix to retrieve all PaymentAccount
	PaymentAccountKeyPrefix = "PaymentAccount/value/"
)

// PaymentAccountKey returns the store key to retrieve a PaymentAccount from the index fields
func PaymentAccountKey(
	addr string,
) []byte {
	var key []byte

	addrBytes := []byte(addr)
	key = append(key, addrBytes...)
	key = append(key, []byte("/")...)

	return key
}
