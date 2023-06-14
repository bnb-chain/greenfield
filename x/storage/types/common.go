package types

import (
	"bytes"
	"cosmossdk.io/math"
	"crypto/sha256"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewSecondarySpSignDoc creating the doc for all secondary sps bls signings,
// checksums is the integrity hash of slice of integrity hash(objectInfo checkSums) contributed by secondary sps
func NewSecondarySpSignDoc(objectID math.Uint, gvgId uint32, checksums []byte) *SecondarySpSignDoc {
	return &SecondarySpSignDoc{
		GlobalVirtualGroupId: gvgId,
		ObjectId:             objectID,
		Checksums:            checksums,
	}
}

func (c *SecondarySpSignDoc) GetSignBytes() [32]byte {
	bts, err := rlp.EncodeToBytes(c)
	if err != nil {
		panic(err)
	}

	btsHash := sdk.Keccak256Hash(bts)
	return btsHash
}

// GenerateIntegrityHash generates integrity hash of all piece data checksum
func GenerateIntegrityHash(checksumList [][]byte) []byte {
	hash := sha256.New()
	checksumBytesTotal := bytes.Join(checksumList, []byte(""))
	hash.Write(checksumBytesTotal)
	return hash.Sum(nil)
}
