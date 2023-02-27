package keeper

import (
	"encoding/binary"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandaoMixLength is the length of randao mix in Tendermint header
const RandaoMixLength = 64

// SeedFromRandaoMix generates seed from randao mix.
func SeedFromRandaoMix(randaoMix []byte, number uint64) []byte {
	high := ^uint64(0) - number
	highBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(highBytes, high)

	lowBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lowBytes, number)

	seedBytes := make([]byte, RandaoMixLength)

	seedBytes = append(seedBytes, sdk.Keccak256(highBytes)...)
	seedBytes = append(seedBytes, sdk.Keccak256(lowBytes)...)

	for i := range randaoMix {
		randaoMix[i] = randaoMix[i] ^ seedBytes[i]
	}

	return seedBytes
}

// RandomObjectKey generates a random object key for challenge.
func RandomObjectKey(seed []byte) []byte {
	bucketKey := sdk.Keccak256(seed[:32])
	objectKey := sdk.Keccak256(seed[32:])
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
func RandomSegmentIndex(seed []byte, segments uint64) uint32 {
	number := new(big.Int)
	number.SetBytes(sdk.Keccak256(seed)[:32])
	number = big.NewInt(0).Abs(number)
	index := big.NewInt(0).Mod(number, big.NewInt(int64(segments)))
	return uint32(index.Uint64())
}

// RandomRedundancyIndex generates a random redundancy index (storage provider) for challenge.
func RandomRedundancyIndex(seed []byte, sps uint64) int32 {
	number := new(big.Int)
	number.SetBytes(sdk.Keccak256(seed)[32:])
	number = big.NewInt(0).Abs(number)
	index := big.NewInt(0).Mod(number, big.NewInt(int64(sps)))
	return int32(index.Uint64())
}