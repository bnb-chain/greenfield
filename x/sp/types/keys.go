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
	StorageProviderKey = []byte{0x21} // prefix for each key to a storage provider
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// GetStorageProviderKey creates the key for the provider with address
// VALUE: staking/Validator
func GetStorageProviderKey(spAddr sdk.AccAddress) []byte {
	return append(StorageProviderKey, address.MustLengthPrefix(spAddr.Bytes())...)
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
