package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// AutoSettleRecordKeyPrefix is the prefix to retrieve all AutoSettleRecord
	AutoSettleRecordKeyPrefix = "AutoSettleRecord/value/"
)

// AutoSettleRecordKey returns the store key to retrieve a AutoSettleRecord from the index fields
func AutoSettleRecordKey(
	timestamp int64,
	addr string,
) []byte {
	var key []byte

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	key = append(key, timestampBytes...)
	key = append(key, []byte("/")...)

	addrBytes := []byte(addr)
	key = append(key, addrBytes...)
	key = append(key, []byte("/")...)

	return key
}
