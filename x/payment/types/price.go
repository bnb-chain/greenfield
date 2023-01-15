package types

import sdkmath "cosmossdk.io/math"

type StreamRecordChange struct {
	Addr                string
	RateChange          sdkmath.Int
	StaticBalanceChange sdkmath.Int
	LockBalanceChange   sdkmath.Int
	AutoTransfer        bool
}

func NewDefaultStreamRecordChangeWithAddr(addr string) StreamRecordChange {
	return StreamRecordChange{
		Addr:                addr,
		RateChange:          sdkmath.ZeroInt(),
		StaticBalanceChange: sdkmath.ZeroInt(),
		LockBalanceChange:   sdkmath.ZeroInt(),
	}
}

func (change StreamRecordChange) WithRateChange(rateChange sdkmath.Int) StreamRecordChange {
	change.RateChange = rateChange
	return change
}

func (change StreamRecordChange) WithStaticBalanceChange(staticBalanceChange sdkmath.Int) StreamRecordChange {
	change.StaticBalanceChange = staticBalanceChange
	return change
}

func (change StreamRecordChange) WithLockBalanceChange(lockBalanceChange sdkmath.Int) StreamRecordChange {
	change.LockBalanceChange = lockBalanceChange
	return change
}

func (change StreamRecordChange) WithAutoTransfer(autoTransfer bool) StreamRecordChange {
	change.AutoTransfer = autoTransfer
	return change
}

type StorePriceFlow struct {
	SpAddr string
	Rate   sdkmath.Int
}

type StorePrice struct {
	UserPayRate sdkmath.Int
	Flows       []StorePriceFlow
}
