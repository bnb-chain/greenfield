package types

import sdkmath "cosmossdk.io/math"

type ReadPacket uint64

const (
	ReadPacketLevelFree ReadPacket = iota
	ReadPacketLevel1GB
	ReadPacketLevel10GB
)

type FlowChange struct {
	Addr          string
	Rate          sdkmath.Int
	StaticBalance sdkmath.Int
}
