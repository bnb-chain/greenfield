package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// FlowKeyPrefix is the prefix to retrieve all Flow
	FlowKeyPrefix = "Flow/value/"
)

// FlowKey returns the store key to retrieve a Flow from the index fields
func FlowKey(
	from string,
	to string,
) []byte {
	var key []byte

	fromBytes := []byte(from)
	key = append(key, fromBytes...)
	key = append(key, []byte("/")...)

	toBytes := []byte(to)
	key = append(key, toBytes...)
	key = append(key, []byte("/")...)

	return key
}
