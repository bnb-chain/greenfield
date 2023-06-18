package types

import (
	"encoding/binary"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	AutoSettleRecordKeyPrefix    = []byte{0x01}
	AutoResumeRecordKeyPrefix    = []byte{0x02}
	StreamRecordKeyPrefix        = []byte{0x03}
	PaymentAccountCountKeyPrefix = []byte{0x04}
	PaymentAccountKeyPrefix      = []byte{0x05}
	OutFlowKeyPrefix             = []byte{0x06}
	ParamsKey                    = []byte{0x07}
)

// AutoSettleRecordKey returns the store key to retrieve a AutoSettleRecord from the index fields
func AutoSettleRecordKey(
	timestamp int64,
	addr sdk.AccAddress,
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
	res.Addr = sdk.AccAddress(key[8:]).String()
	return
}

// AutoResumeRecordKey returns the store key to retrieve a AutoResumeRecord from the index fields
func AutoResumeRecordKey(
	timestamp int64,
	addr sdk.AccAddress,
) []byte {
	var key []byte

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	key = append(key, timestampBytes...)

	addrBytes := []byte(addr)
	key = append(key, addrBytes...)

	return key
}

func ParseAutoResumeRecordKey(key []byte) (res AutoResumeRecord) {
	res.Timestamp = int64(binary.BigEndian.Uint64(key[0:8]))
	res.Addr = sdk.AccAddress(key[8:]).String()
	return
}

// PaymentAccountKey returns the store key to retrieve a PaymentAccount from the index fields
func PaymentAccountKey(
	addr sdk.AccAddress,
) []byte {
	return addr
}

// PaymentAccountCountKey returns the store key to retrieve a PaymentAccountCount from the index fields
func PaymentAccountCountKey(
	owner sdk.AccAddress,
) []byte {
	return owner
}

// StreamRecordKey returns the store key to retrieve a StreamRecord from the index fields
func StreamRecordKey(
	account sdk.AccAddress,
) []byte {
	return account
}

func OutFlowKey(
	account sdk.AccAddress,
	status OutFlowStatus,
	toAddr sdk.AccAddress) []byte {
	key := account.Bytes()
	if status == OUT_FLOW_STATUS_ACTIVE {
		key = append(key, []byte{0x0}...)
	} else {
		key = append(key, []byte{0x1}...)
	}
	if toAddr != nil && !toAddr.Empty() {
		key = append(key, toAddr.Bytes()...)
	}
	return key
}

func ParseOutFlowKey(key []byte) (account sdk.AccAddress, res OutFlow) {
	account = key[0:20]
	if key[20] == byte(0) {
		res.Status = OUT_FLOW_STATUS_ACTIVE
	} else {
		res.Status = OUT_FLOW_STATUS_FROZEN
	}
	res.ToAddress = sdk.AccAddress(key[21:]).String()
	return
}

func ParseOutFlowValue(value []byte) sdkmath.Int {
	rate := sdk.ZeroInt()
	if err := rate.Unmarshal(value); err != nil {
		panic("should not happen")
	}
	return rate
}
