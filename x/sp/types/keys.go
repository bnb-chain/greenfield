package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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
	// Keys for store prefixes
	StorageProviderKey               = []byte{0x21} // prefix for each key to a storage provider
	StorageProviderByFundingAddrKey  = []byte{0x22} // prefix for each key to a storage provider index, by funding address
	StorageProviderBySealAddrKey     = []byte{0x23} // prefix for each key to a storage provider index, by seal address
	StorageProviderByApprovalAddrKey = []byte{0x24} // prefix for each key to a storage provider index, by approval address
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// GetStorageProviderKey creates the key for the provider with address
// VALUE: staking/Validator
func GetStorageProviderKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderKey, address.MustLengthPrefix(spAddr.Bytes())...)
}

// GetStorageProviderByFundingAddrKey creates the key for the storage provider with funding address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderByFundingAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderByFundingAddrKey, address.MustLengthPrefix(spAddr.Bytes())...)
}

// GetStorageProviderBySealAddrKey creates the key for the storage provider with seal address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderBySealAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderBySealAddrKey, address.MustLengthPrefix(spAddr.Bytes())...)
}

// GetStorageProviderByApprovalAddrKey creates the key for the storage provider with approval address
// VALUE: storage provider operator address ([]byte)
func GetStorageProviderByApprovalAddrKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderByApprovalAddrKey, address.MustLengthPrefix(spAddr.Bytes())...)
}

func UnmarshalStorageProvider(cdc codec.BinaryCodec, value []byte) (sp StorageProvider, err error) {
	err = cdc.Unmarshal(value, &sp)
	return sp, err
}

func MustUnmarshalStorageProvider(cdc codec.BinaryCodec, value []byte) StorageProvider {
	sp, err := UnmarshalStorageProvider(cdc, value)
	if err != nil {
		panic(err)
	}

	return sp
}

func MustMarshalStorageProvider(cdc codec.BinaryCodec, sp *StorageProvider) []byte {
	return cdc.MustMarshal(sp)
}
