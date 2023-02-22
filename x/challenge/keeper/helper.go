package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandomObjectKey generates a random object key for challenge.
func RandomObjectKey(randaoMix []byte) []byte {
	bucketKey := sdk.Keccak256(randaoMix[:32])
	objectKey := sdk.Keccak256(randaoMix[32:])
	return append(bucketKey, objectKey...)
}

// CalculateSegments calculates the number of segments for the payload size.
func CalculateSegments(payloadSize, segmentSize uint64) uint64 {
	segments := payloadSize / segmentSize
	if payloadSize%segmentSize == 0 {
		return segments
	}
	segments++
	return segments
}

// RandomSegmentIndex generates a random segment index for challenge.
func RandomSegmentIndex(randaoMix []byte, segments uint64) uint32 {
	number := new(big.Int)
	number.SetBytes(sdk.Keccak256(randaoMix)[:32])
	number = big.NewInt(0).Abs(number)
	index := big.NewInt(0).Mod(number, big.NewInt(int64(segments)))
	return uint32(index.Uint64())
}

// RandomRedundancyIndex generates a random redundancy index (storage provider) for challenge.
func RandomRedundancyIndex(randaoMix []byte, sps uint64) int32 {
	number := new(big.Int)
	number.SetBytes(sdk.Keccak256(randaoMix)[32:])
	number = big.NewInt(0).Abs(number)
	index := big.NewInt(0).Mod(number, big.NewInt(int64(sps)))
	return int32(index.Uint64())
}
