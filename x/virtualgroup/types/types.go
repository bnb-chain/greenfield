package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
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

func (f *GlobalVirtualGroupFamily) MustRemoveGVG(gvgID uint32) {
	err := f.RemoveGVG(gvgID)
	if err != nil {
		panic(fmt.Sprintf("remove gvg from family failed. err: %s", err))
	}
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
