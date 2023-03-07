package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName defines the module name
	ModuleName = "payment"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_payment"
)

var (
	// GovernanceAddress used to receive fee of storage system, and pay for the potential debt from late forced settlement
	GovernanceAddress = sdk.AccAddress(address.Module(ModuleName, []byte("governance")))

	AutoSettleRecordKeyPrefix    = []byte{0x01}
	StreamRecordKeyPrefix        = []byte{0x02}
	PaymentAccountCountKeyPrefix = []byte{0x03}
	PaymentAccountKeyPrefix      = []byte{0x04}
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

	addrBytes := []byte(addr)
	key = append(key, addrBytes...)

	return key
}

func ParseAutoSettleRecordKey(key []byte) (res AutoSettleRecord) {
	res.Timestamp = int64(binary.BigEndian.Uint64(key[0:8]))
	res.Addr = string(key[8:])
	return
}

// PaymentAccountKey returns the store key to retrieve a PaymentAccount from the index fields
func PaymentAccountKey(
	addr string,
) []byte {
	var key []byte

	addrBytes := []byte(addr)
	key = append(key, addrBytes...)

	return key
}

// PaymentAccountCountKey returns the store key to retrieve a PaymentAccountCount from the index fields
func PaymentAccountCountKey(
	owner string,
) []byte {
	var key []byte

	ownerBytes := []byte(owner)
	key = append(key, ownerBytes...)

	return key
}

// StreamRecordKey returns the store key to retrieve a StreamRecord from the index fields
func StreamRecordKey(
	account string,
) []byte {
	var key []byte

	accountBytes := []byte(account)
	key = append(key, accountBytes...)

	return key
}
