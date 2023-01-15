package types

import sdkmath "cosmossdk.io/math"

type StreamRecordChange struct {
	Addr          string
	Rate          sdkmath.Int
	StaticBalance sdkmath.Int
}

func NewDefaultStreamRecordChangeWithAddr(addr string) StreamRecordChange {
	return StreamRecordChange{
		Addr:          addr,
		Rate:          sdkmath.ZeroInt(),
		StaticBalance: sdkmath.ZeroInt(),
	}
}
