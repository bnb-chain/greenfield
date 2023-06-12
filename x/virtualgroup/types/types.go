package types

func (f *GlobalVirtualGroupFamily) AppendGVG(gvgID uint32) {
	f.GlobalVirtualGroupIds = append(f.GlobalVirtualGroupIds, gvgID)
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
