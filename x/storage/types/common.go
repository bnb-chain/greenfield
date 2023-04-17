package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSecondarySpSignDoc(spAddress sdk.AccAddress, objectID math.Uint, checksum []byte) *SecondarySpSignDoc {
	return &SecondarySpSignDoc{
		SpAddress: spAddress.String(),
		ObjectId:  objectID,
		Checksum:  checksum,
	}
}

func (sr *SecondarySpSignDoc) GetSignBytes() []byte {
	panic("GetSignBytes")

	bz := ModuleCdc.MustMarshalJSON(sr)
	return sdk.MustSortJSON(bz)
}
