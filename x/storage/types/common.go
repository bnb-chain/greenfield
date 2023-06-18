package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSecondarySpSignDoc(spID uint32, objectID math.Uint, checksum []byte) *SecondarySpSignDoc {
	return &SecondarySpSignDoc{
		SpId:     spID,
		ObjectId: objectID,
		Checksum: checksum,
	}
}

func (sr *SecondarySpSignDoc) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(sr)
	return sdk.MustSortJSON(bz)
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
