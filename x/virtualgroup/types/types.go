package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Int  = sdkmath.Int
	Uint = sdkmath.Uint
)

func (f *GlobalVirtualGroupFamily) AppendGVG(gvgID uint32) {
	f.GlobalVirtualGroupIds = append(f.GlobalVirtualGroupIds, gvgID)
}

func (f *GlobalVirtualGroupFamily) Contains(gvgID uint32) bool {
	for _, id := range f.GlobalVirtualGroupIds {
		if id == gvgID {
			return true
		}
	}
	return false
}

func (f *GlobalVirtualGroupFamily) RemoveGVG(gvgID uint32) error {
	for i, id := range f.GlobalVirtualGroupIds {
		if id == gvgID {
			f.GlobalVirtualGroupIds = append(f.GlobalVirtualGroupIds[:i], f.GlobalVirtualGroupIds[i+1:]...)
			return nil
		}
	}
	return ErrGVGNotExist
}

func (g *GlobalVirtualGroupsBindingOnBucket) AppendGVGAndLVG(gvgID, lvgID uint32) {
	g.GlobalVirtualGroupIds = append(g.GlobalVirtualGroupIds, gvgID)
	g.LocalVirtualGroupIds = append(g.LocalVirtualGroupIds, lvgID)
}

func (g *GlobalVirtualGroupsBindingOnBucket) GetLVGIDByGVGID(gvgID uint32) uint32 {
	for i, gID := range g.GlobalVirtualGroupIds {
		if gID == gvgID {
			return g.LocalVirtualGroupIds[i]
		}
	}
	return 0
}

func (g *GlobalVirtualGroupsBindingOnBucket) GetGVGIDByLVGID(lvgID uint32) uint32 {
	for i, lID := range g.LocalVirtualGroupIds {
		if lID == lvgID {
			return g.GlobalVirtualGroupIds[i]
		}
	}
	return 0
}

func NewMigrationBucketSignDoc(bucketID sdkmath.Uint, spID, lvgID, srcGVGID, dstGVGID uint32) *MigrationBucketSignDoc {
	return &MigrationBucketSignDoc{
		BucketId:                bucketID,
		DstPrimarySpId:          spID,
		LocalVirtualGroupId:     lvgID,
		SrcGlobalVirtualGroupId: srcGVGID,
		DstGlobalVirtualGroupId: dstGVGID,
	}
}

func (c *MigrationBucketSignDoc) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(c))
}

func (c *MigrationBucketSignDoc) GetBlsSignHash() [32]byte {
	return sdk.Keccak256Hash(c.GetSignBytes())
}
