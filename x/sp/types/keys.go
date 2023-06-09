package types

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/types/address"

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
	StorageProviderByFundingAddrKey  = []byte{0x22} // prefix for each key to a storage provider index, by funding address
	StorageProviderBySealAddrKey     = []byte{0x23} // prefix for each key to a storage provider index, by seal address
	StorageProviderByApprovalAddrKey = []byte{0x24} // prefix for each key to a storage provider index, by approval address
	StorageProviderByGcAddrKey       = []byte{0x25} // prefix for each key to a storage provider index, by gc address
	SpStoragePriceKeyPrefix          = []byte{0x26}
	SecondarySpStorePriceKeyPrefix   = []byte{0x27}
	StorageProviderByBlsPubKeyKey    = []byte{0x28} // prefix for each key to a storage provider index, by bls pub key

)

// GetStorageProviderKey creates the key for the provider with address
// VALUE: staking/Validator
func GetStorageProviderKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderKey, spAddr.Bytes()...)
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

// GetStorageProviderByApprovalAddrKey creates the key for the storage provider with approval address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderByGcAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderByGcAddrKey, spAddr.Bytes()...)
}

// GetStorageProviderByBlsKeyKey creates the key for the storage provider with bls pub key
func GetStorageProviderByBlsKeyKey(blsPk []byte) []byte {
	return append(StorageProviderByBlsPubKeyKey, address.MustLengthPrefix(blsPk)...)
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
	sp sdk.AccAddress,
	UpdateTimeSec int64,
) []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(UpdateTimeSec))

	var key []byte
	key = append(key, sp...)
	key = append(key, timeBytes...)

	return key
}

func ParseSpStoragePriceKey(key []byte) (spAddr sdk.AccAddress, UpdateTimeSec int64) {
	length := len(key)
	spAddr = key[:length-8]
	UpdateTimeSec = int64(binary.BigEndian.Uint64(key[length-8 : length]))
	return
}

func SecondarySpStorePriceKey(
	UpdateTimeSec int64,
) []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(UpdateTimeSec))
	return timeBytes
}

func ParseSecondarySpStorePriceKey(key []byte) (UpdateTimeSec int64) {
	UpdateTimeSec = int64(binary.BigEndian.Uint64(key))
	return
}
