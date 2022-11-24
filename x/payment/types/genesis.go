package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		StreamRecordList:        []StreamRecord{},
		PaymentAccountCountList: []PaymentAccountCount{},
		PaymentAccountList:      []PaymentAccount{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in streamRecord
	streamRecordIndexMap := make(map[string]struct{})

	for _, elem := range gs.StreamRecordList {
		index := string(StreamRecordKey(elem.Account))
		if _, ok := streamRecordIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for streamRecord")
		}
		streamRecordIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in paymentAccountCount
	paymentAccountCountIndexMap := make(map[string]struct{})

	for _, elem := range gs.PaymentAccountCountList {
		index := string(PaymentAccountCountKey(elem.Owner))
		if _, ok := paymentAccountCountIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for paymentAccountCount")
		}
		paymentAccountCountIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in paymentAccount
	paymentAccountIndexMap := make(map[string]struct{})

	for _, elem := range gs.PaymentAccountList {
		index := string(PaymentAccountKey(elem.Addr))
		if _, ok := paymentAccountIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for paymentAccount")
		}
		paymentAccountIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
