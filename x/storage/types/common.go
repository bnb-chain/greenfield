package types

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSecondarySpSignDoc(objectID math.Uint, gvgId math.Uint, checkSums [][]byte) *SecondarySpSignDoc {
	css := make([]byte, 0)
	for _, cs := range checkSums {
		css = append(css, cs...)
	}
	return &SecondarySpSignDoc{
		GlobalGroupId: gvgId,
		ObjectId:      objectID,
		Checksum:      css,
	}
}

// GetSignBytes get the event hash
func (c *SecondarySpSignDoc) GetSignBytes() [32]byte {
	bts, err := rlp.EncodeToBytes(c)
	if err != nil {
		panic(err)
	}

	btsHash := sdk.Keccak256Hash(bts)
	return btsHash
}
