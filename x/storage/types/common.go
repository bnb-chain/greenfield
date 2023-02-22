package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSecondarySpSignDoc(spAddress sdk.AccAddress, checksum []byte) *SecondarySpSignDoc {
	return &SecondarySpSignDoc{
		SpAddress: spAddress.String(),
		Checksum:  checksum,
	}
}

func (sr *SecondarySpSignDoc) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(sr)
	return sdk.MustSortJSON(bz)
}
