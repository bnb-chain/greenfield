package types

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "sp"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_sp"
)

var (
	ParamsKey = []byte{0x01}

	StorageProviderKey               = []byte{0x21} // prefix for each key to a storage provider
	StorageProviderByOperatorAddrKey = []byte{0x23} // prefix for each key to a storage provider index, by operator address
	StorageProviderByFundingAddrKey  = []byte{0x24} // prefix for each key to a storage provider index, by funding address
	StorageProviderBySealAddrKey     = []byte{0x25} // prefix for each key to a storage provider index, by seal address
	StorageProviderByApprovalAddrKey = []byte{0x26} // prefix for each key to a storage provider index, by approval address
	StorageProviderByGcAddrKey       = []byte{0x27} // prefix for each key to a storage provider index, by gc address
	SpStoragePriceKeyPrefix          = []byte{0x28}
	SecondarySpStorePriceKeyPrefix   = []byte{0x29}
	StorageProviderByBlsPubKeyKey    = []byte{0x30} // prefix for each key to a storage provider index, by bls pub key

	StorageProviderSequenceKey = []byte{0x31}
)

// GetStorageProviderKey creates the key for the provider with address
// VALUE: staking/Validator
func GetStorageProviderKey(id []byte) []byte {
	return append(StorageProviderKey, id...)

}

func GetStorageProviderByOperatorAddrKey(operatorAddr sdk.AccAddress) []byte {
	return append(StorageProviderByOperatorAddrKey, operatorAddr.Bytes()...)

}

// GetStorageProviderByFundingAddrKey creates the key for the storage provider with funding address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderByFundingAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderByFundingAddrKey, spAddr.Bytes()...)
}

// GetStorageProviderBySealAddrKey creates the key for the storage provider with seal address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderBySealAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderBySealAddrKey, spAddr.Bytes()...)
}

// GetStorageProviderByApprovalAddrKey creates the key for the storage provider with approval address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderByApprovalAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderByApprovalAddrKey, spAddr.Bytes()...)
}

// GetStorageProviderByGcAddrKey creates the key for the storage provider with approval address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderByGcAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderByGcAddrKey, spAddr.Bytes()...)
}

// GetStorageProviderByBlsKeyKey creates the key for the storage provider with bls pub key
func GetStorageProviderByBlsKeyKey(blsPk []byte) []byte {
	return append(StorageProviderByBlsPubKeyKey, blsPk...)
}

func UnmarshalStorageProvider(cdc codec.BinaryCodec, value []byte) (sp *StorageProvider, err error) {
	sp = &StorageProvider{}
	err = cdc.Unmarshal(value, sp)
	return sp, err
}

func MustUnmarshalStorageProvider(cdc codec.BinaryCodec, value []byte) *StorageProvider {
	sp, err := UnmarshalStorageProvider(cdc, value)
	if err != nil {
		panic(err)
	}

	return sp
}

func MustMarshalStorageProvider(cdc codec.BinaryCodec, sp *StorageProvider) []byte {
	return cdc.MustMarshal(sp)
}

func SpStoragePriceKey(
	spId uint32,
	timestamp int64,
) []byte {
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, spId)

	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timestamp))

	var key []byte
	key = append(key, idBytes...)
	key = append(key, timeBytes...)

	return key
}

func ParseSpStoragePriceKey(key []byte) (spId uint32, timestamp int64) {
	spId = binary.BigEndian.Uint32(key[0:4])
	timestamp = int64(binary.BigEndian.Uint64(key[4:]))
	return
}

func SecondarySpStorePriceKey(
	timestamp int64,
) []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timestamp))
	return timeBytes
}

func ParseSecondarySpStorePriceKey(key []byte) (timestamp int64) {
	timestamp = int64(binary.BigEndian.Uint64(key))
	return
}
