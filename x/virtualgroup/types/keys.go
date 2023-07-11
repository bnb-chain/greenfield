package types

import (
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

	// GVGVirtualPaymentAccountName string for derive the virtual payment account for GVG
	GVGVirtualPaymentAccountName = "global_virtual_group"

	// GVGFamilyName string for derive the virtual payment account for GVG family
	GVGFamilyName = "global_virtual_group_family"

	// NoSpecifiedFamilyId defines
	NoSpecifiedFamilyId = uint32(0)
)

var (
	ParamsKey = []byte{0x01}

	GVGKey       = []byte{0x21}
	GVGFamilyKey = []byte{0x22}

	GVGSequencePrefix       = []byte{0x32}
	GVGFamilySequencePrefix = []byte{0x33}

	GVGStatisticsWithinSPKey = []byte{0x41}

	SwapOutFamilyKey = []byte{0x51}
	SwapOutGVGKey    = []byte{0x61}
)

func GetGVGKey(gvgID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(GVGKey, uint32Seq.EncodeSequence(gvgID)...)
}

func GetGVGFamilyKey(spID uint32, familyID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(GVGFamilyKey, append(uint32Seq.EncodeSequence(spID), uint32Seq.EncodeSequence(familyID)...)...)
}

func GetGVGFamilyPrefixKey(spID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(GVGFamilyKey, uint32Seq.EncodeSequence(spID)...)
}

func GetGVGStatisticsWithinSPKey(spID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(GVGStatisticsWithinSPKey, uint32Seq.EncodeSequence(spID)...)
}

func GetSwapOutFamilyKey(globalVirtualGroupFamilyID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(SwapOutFamilyKey, uint32Seq.EncodeSequence(globalVirtualGroupFamilyID)...)
}

func GetSwapOutGVGKey(globalVirtualGroupID uint32) []byte {
	var uint32Seq sequence.Sequence[uint32]
	return append(SwapOutGVGKey, uint32Seq.EncodeSequence(globalVirtualGroupID)...)
}
