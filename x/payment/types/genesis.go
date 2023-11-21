package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		StreamRecordList:        []StreamRecord{},
		PaymentAccountCountList: []PaymentAccountCount{},
		PaymentAccountList:      []PaymentAccount{},
		AutoSettleRecordList:    []AutoSettleRecord{},
		Params:                  DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in streamRecord
	streamRecordIndexMap := make(map[string]struct{})

	for _, elem := range gs.StreamRecordList {
		index := string(StreamRecordKey(sdk.MustAccAddressFromHex(elem.Account)))
		if _, ok := streamRecordIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for streamRecord")
		}
		streamRecordIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in paymentAccountCount
	paymentAccountCountIndexMap := make(map[string]struct{})

	for _, elem := range gs.PaymentAccountCountList {
		index := string(PaymentAccountCountKey(sdk.MustAccAddressFromHex(elem.Owner)))
		if _, ok := paymentAccountCountIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for paymentAccountCount")
		}
		paymentAccountCountIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in paymentAccount
	paymentAccountIndexMap := make(map[string]struct{})

	for _, elem := range gs.PaymentAccountList {
		index := string(PaymentAccountKey(sdk.MustAccAddressFromHex(elem.Addr)))
		if _, ok := paymentAccountIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for paymentAccount")
		}
		paymentAccountIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in autoSettleRecord
	autoSettleRecordIndexMap := make(map[string]struct{})

	for _, elem := range gs.AutoSettleRecordList {
		index := string(AutoSettleRecordKey(elem.Timestamp, sdk.MustAccAddressFromHex(elem.Addr)))
		if _, ok := autoSettleRecordIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for autoSettleRecord")
		}
		autoSettleRecordIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
