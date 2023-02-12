package types

import (
	sdkmath "cosmossdk.io/math"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type StreamRecordChange struct {
	Addr                string
	RateChange          sdkmath.Int
	StaticBalanceChange sdkmath.Int
	LockBalanceChange   sdkmath.Int
}

func NewDefaultStreamRecordChangeWithAddr(addr string) *StreamRecordChange {
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

type StorePrice struct {
	UserPayRate sdkmath.Int
	Flows       []storagetypes.OutFlowInUSD
}

type BNBPrice struct {
	Num       sdkmath.Int
	Precision sdkmath.Int
}

func (price BNBPrice) Equal(other BNBPrice) bool {
	return price.Num.Equal(other.Num) && price.Precision.Equal(other.Precision)
}
