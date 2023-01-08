package types

type ReadPacketLevel uint64

const (
	ReadPacketLevelFree ReadPacketLevel = iota
	ReadPacketLevel1GB
	ReadPacketLevel10GB
)
