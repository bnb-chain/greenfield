package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// StreamRecordKeyPrefix is the prefix to retrieve all StreamRecord
	StreamRecordKeyPrefix = "StreamRecord/value/"
)

// StreamRecordKey returns the store key to retrieve a StreamRecord from the index fields
func StreamRecordKey(
	account string,
) []byte {
	var key []byte

	accountBytes := []byte(account)
	key = append(key, accountBytes...)
	key = append(key, []byte("/")...)

	return key
}
