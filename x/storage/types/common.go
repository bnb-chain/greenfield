package types

import (
	"bytes"
	"crypto/sha256"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewSecondarySpSealObjectSignDoc creating the doc for all secondary sps bls signings,
// checksums is the hash of slice of integrity hash(objectInfo checkSums) contributed by secondary sps
func NewSecondarySpSealObjectSignDoc(objectID math.Uint, gvgId uint32, checksum []byte) *SecondarySpSealObjectSignDoc {
	return &SecondarySpSealObjectSignDoc{
		GlobalVirtualGroupId: gvgId,
		ObjectId:             objectID,
		Checksum:             checksum,
	}
}

func (c *SecondarySpSealObjectSignDoc) GetSignBytes() [32]byte {
	bts, err := rlp.EncodeToBytes(c)
	if err != nil {
		panic(err)
	}

	btsHash := sdk.Keccak256Hash(bts)
	return btsHash
}

// GenerateHash generates sha256 hash of checksums from objectInfo
func GenerateHash(checksumList [][]byte) []byte {
	hash := sha256.New()
	checksumBytesTotal := bytes.Join(checksumList, []byte(""))
	hash.Write(checksumBytesTotal)
	return hash.Sum(nil)
}

func NewMigrationBucketSignDoc(srcSPID, dspSPID uint32, bucketID math.Uint) *MigrationBucketSignDoc {
	return &MigrationBucketSignDoc{
		SrcSpId:  srcSPID,
		DstSpId:  dspSPID,
		BucketId: bucketID,
	}
}

func (mbs *MigrationBucketSignDoc) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(mbs)
	return sdk.MustSortJSON(bz)
}
