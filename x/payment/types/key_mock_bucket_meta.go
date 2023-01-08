package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // MockBucketMetaKeyPrefix is the prefix to retrieve all MockBucketMeta
	MockBucketMetaKeyPrefix = "MockBucketMeta/value/"
)

// MockBucketMetaKey returns the store key to retrieve a MockBucketMeta from the index fields
func MockBucketMetaKey(
bucketName string,
) []byte {
	var key []byte
    
    bucketNameBytes := []byte(bucketName)
    key = append(key, bucketNameBytes...)
    key = append(key, []byte("/")...)
    
	return key
}