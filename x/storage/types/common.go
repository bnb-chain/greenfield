package types

import (
	"bytes"
	"crypto/sha256"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewSecondarySpSealObjectSignDoc creates the doc for all secondary sps bls signings,
// checksums is the hash of slice of integrity hash(objectInfo checkSums) contributed by secondary sps
func NewSecondarySpSealObjectSignDoc(chainID string, gvgID uint32, objectID math.Uint, checksum []byte) *SecondarySpSealObjectSignDoc {
	return &SecondarySpSealObjectSignDoc{
		ChainId:              chainID,
		GlobalVirtualGroupId: gvgID,
		ObjectId:             objectID,
		Checksum:             checksum,
	}
}

func (c *SecondarySpSealObjectSignDoc) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(c))
}

func (c *SecondarySpSealObjectSignDoc) GetBlsSignHash() [32]byte {
	return sdk.Keccak256Hash(c.GetSignBytes())
}

// GenerateHash generates sha256 hash of checksums from objectInfo
func GenerateHash(checksumList [][]byte) []byte {
	hash := sha256.New()
	checksumBytesTotal := bytes.Join(checksumList, []byte(""))
	hash.Write(checksumBytesTotal)
	return hash.Sum(nil)
}

func NewSecondarySpMigrationBucketSignDoc(chainID string, bucketID math.Uint, spID, srcGVGID, dstGVGID uint32) *SecondarySpMigrationBucketSignDoc {
	return &SecondarySpMigrationBucketSignDoc{
		ChainId:                 chainID,
		BucketId:                bucketID,
		DstPrimarySpId:          spID,
		SrcGlobalVirtualGroupId: srcGVGID,
		DstGlobalVirtualGroupId: dstGVGID,
	}
}

func (c *SecondarySpMigrationBucketSignDoc) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(c))
}

func (c *SecondarySpMigrationBucketSignDoc) GetBlsSignHash() [32]byte {
	return sdk.Keccak256Hash(c.GetSignBytes())
}
