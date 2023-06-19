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
