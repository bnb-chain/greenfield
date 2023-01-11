package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	defaultSingleBnbPrice := SingleBnbPrice{0, 27740000000}
	defaultBnbPrice := BnbPrice{
		Prices: []*SingleBnbPrice{&defaultSingleBnbPrice},
	}
	return &GenesisState{
		StreamRecordList:        []StreamRecord{},
		PaymentAccountCountList: []PaymentAccountCount{},
		PaymentAccountList:      []PaymentAccount{},
		MockBucketMetaList:      []MockBucketMeta{},
		FlowList:                []Flow{},
		BnbPrice:                &defaultBnbPrice,
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

	// Check for duplicated index in mockBucketMeta
	mockBucketMetaIndexMap := make(map[string]struct{})

	for _, elem := range gs.MockBucketMetaList {
		index := string(MockBucketMetaKey(elem.BucketName))
		if _, ok := mockBucketMetaIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for mockBucketMeta")
		}
		mockBucketMetaIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in flow
	flowIndexMap := make(map[string]struct{})

	for _, elem := range gs.FlowList {
		index := string(FlowKey(elem.From, elem.To))
		if _, ok := flowIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for flow")
		}
		flowIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
