package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StreamRecordChange struct {
	Addr                sdk.AccAddress
	RateChange          sdkmath.Int
	StaticBalanceChange sdkmath.Int
	LockBalanceChange   sdkmath.Int
}

func NewDefaultStreamRecordChangeWithAddr(addr sdk.AccAddress) *StreamRecordChange {
	return &StreamRecordChange{
		Addr:                addr,
		RateChange:          sdkmath.ZeroInt(),
		StaticBalanceChange: sdkmath.ZeroInt(),
		LockBalanceChange:   sdkmath.ZeroInt(),
	}
}

func (change *StreamRecordChange) WithRateChange(rateChange sdkmath.Int) *StreamRecordChange {
	change.RateChange = rateChange
	return change
}

func (change *StreamRecordChange) WithStaticBalanceChange(staticBalanceChange sdkmath.Int) *StreamRecordChange {
	change.StaticBalanceChange = staticBalanceChange
	return change
}

func (change *StreamRecordChange) WithLockBalanceChange(lockBalanceChange sdkmath.Int) *StreamRecordChange {
	change.LockBalanceChange = lockBalanceChange
	return change
}

type StoragePriceParams struct {
	PrimarySp string
	PriceTime int64
}

type StoragePrice struct {
	ReadPrice           sdk.Dec
	PrimaryStorePrice   sdk.Dec
	SecondaryStorePrice sdk.Dec
}

type UserFlows struct {
	From  sdk.AccAddress
	Flows []OutFlow
}
