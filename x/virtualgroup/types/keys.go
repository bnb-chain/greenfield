package types

import (
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/internal/sequence"
)

const (
	// ModuleName defines the module name
	ModuleName = "virtualgroup"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_virtualgroup"

	// TStoreKey defines the transient store key
	TStoreKey = "transient_storage"

	// LVGName string for derive the virtual payment account for LVG
	LVGName = "local_virtual_group"

	// GVGName string for derive the virtual payment account for GVG
	GVGName = "global_virtual_group"

	// GVGFamilyName string for derive the virtual payment account for GVG family
	GVGFamilyName = "global_virtual_group_family"
)

var (
	ParamsKey = []byte{0x01}

	LVGKey       = []byte{0x21}
	GVGKey       = []byte{0x22}
	GVGFamilyKey = []byte{0x23}

	LVGSequencePrefix       = []byte{0x31}
	GVGSequencePrefix       = []byte{0x32}
	GVGFamilySequencePrefix = []byte{0x33}
)

func GetLVGKey(bucketID math.Uint, lvgID uint32) []byte {
	var uint256Seq sequence.Sequence[math.Uint]
	var uint32Seq sequence.Sequence[uint32]
	return append(LVGKey, append(uint256Seq.EncodeSequence(bucketID), uint32Seq.EncodeSequence(lvgID)...)...)
}

func GetGVGKey(spID uint32, gvgID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(GVGKey, append(uint32Seq.EncodeSequence(spID), uint32Seq.EncodeSequence(gvgID)...)...)
}

func GetGVGFamilyKey(spID uint32, familyID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(GVGFamilyKey, append(uint32Seq.EncodeSequence(spID), uint32Seq.EncodeSequence(familyID)...)...)
}
