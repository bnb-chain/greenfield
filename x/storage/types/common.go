package types

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewSecondarySpSignDoc creating the dec for all secondary sps bls signing,
// checksums is a slice of integrity hash contributed by secondary sp
func NewSecondarySpSignDoc(objectID math.Uint, gvgId uint32, checksums [][]byte) *SecondarySpSignDoc {
	return &SecondarySpSignDoc{
		GlobalGroupId: gvgId,
		ObjectId:      objectID,
		Checksums:     checksums,
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
