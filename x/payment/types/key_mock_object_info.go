package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // MockObjectInfoKeyPrefix is the prefix to retrieve all MockObjectInfo
	MockObjectInfoKeyPrefix = "MockObjectInfo/value/"
)

// MockObjectInfoKey returns the store key to retrieve a MockObjectInfo from the index fields
func MockObjectInfoKey(
bucketName string,
objectName string,
) []byte {
	var key []byte
    
    bucketNameBytes := []byte(bucketName)
    key = append(key, bucketNameBytes...)
    key = append(key, []byte("/")...)
    
    objectNameBytes := []byte(objectName)
    key = append(key, objectNameBytes...)
    key = append(key, []byte("/")...)
    
	return key
}