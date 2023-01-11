package types

import sdkmath "cosmossdk.io/math"

type ReadPacket uint64

const (
	ReadPacketLevelFree ReadPacket = iota
	ReadPacketLevel1GB
	ReadPacketLevel10GB
)

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
