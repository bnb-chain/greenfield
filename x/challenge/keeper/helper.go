package keeper

import (
	"encoding/binary"
	"math/big"

	sdkmath "cosmossdk.io/math"
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

	seedBytes := make([]byte, 0)
	seedBytes = append(seedBytes, sdk.Keccak256(highBytes)...)
	seedBytes = append(seedBytes, sdk.Keccak256(lowBytes)...)

	for i := range randaoMix {
		seedBytes[i] = randaoMix[i] ^ seedBytes[i]
	}

	return seedBytes
}

// RandomObjectId generates a random object id for challenge.
// Be noted: id starts from 1.
func RandomObjectId(seed []byte, objectCount sdkmath.Uint) sdkmath.Uint {
	number := new(big.Int).SetBytes(sdk.Keccak256(seed))
	number = new(big.Int).Abs(number)

	id := sdkmath.NewUintFromBigInt(number).Mod(objectCount).AddUint64(1)
	return id
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
	number := new(big.Int).SetBytes(sdk.Keccak256(seed[:32]))
	number = new(big.Int).Abs(number)
	index := new(big.Int).Mod(number, big.NewInt(int64(segments)))
	return uint32(index.Uint64())
}

// RandomRedundancyIndex generates a random redundancy index (storage provider) for challenge.
// Be noted: RedundancyIndex starts from -1 (the primary sp).
func RandomRedundancyIndex(seed []byte, sps uint64) int32 {
	number := new(big.Int).SetBytes(sdk.Keccak256(seed[32:]))
	number = new(big.Int).Abs(number)
	index := new(big.Int).Mod(number, big.NewInt(int64(sps)))
	return int32(index.Uint64()) - 1
}
