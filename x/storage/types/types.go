package types

import (
	"fmt"
	"reflect"

	sdkmath "cosmossdk.io/math"
)

type (
	Int  = sdkmath.Int
	Uint = sdkmath.Uint
)

const (
	TagKeyTraits       = "traits"
	TagValueOmit       = "omit"
	MaxPaginationLimit = 200 // the default limit is 100 if pagination parameters is not provided
)

func (m *BucketInfo) ToNFTMetadata() *BucketMetaData {
	return &BucketMetaData{
		BucketName: m.BucketName,
		Attributes: getNFTAttributes(*m),
	}
}

func (m *BucketInfo) CheckBucketStatus() error {
	if m.BucketStatus == BUCKET_STATUS_DISCONTINUED {
		return ErrBucketDiscontinued
	} else if m.BucketStatus == BUCKET_STATUS_MIGRATING {
		return ErrBucketMigrating
	}
	return nil
}

func (m *ObjectInfo) ToNFTMetadata() *ObjectMetaData {
	return &ObjectMetaData{
		ObjectName: m.ObjectName,
		Attributes: getNFTAttributes(*m),
	}
}

func (m *GroupInfo) ToNFTMetadata() *GroupMetaData {
	return &GroupMetaData{
		GroupName:  m.GroupName,
		Attributes: getNFTAttributes(*m),
	}
}

func getNFTAttributes(m interface{}) []Trait {
	attributes := make([]Trait, 0)
	v := reflect.ValueOf(m)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if typ.Field(i).Tag.Get(TagKeyTraits) == TagValueOmit {
			continue
		}
		attributes = append(attributes,
			Trait{
				TraitType: typ.Field(i).Name,
				Value:     fmt.Sprintf("%v", v.Field(i).Interface()),
			})
	}
	return attributes
}

func (di *DeleteInfo) IsEmpty() bool {
	isBucketIdsEmpty := false
	isObjectIdsEmpty := false
	isGroupIdsEmpty := false
	if di == nil {
		return true
	}
	if di.BucketIds == nil || len(di.BucketIds.Id) == 0 {
		isBucketIdsEmpty = true
	}
	if di.ObjectIds == nil || len(di.ObjectIds.Id) == 0 {
		isObjectIdsEmpty = true
	}
	if di.GroupIds == nil || len(di.GroupIds.Id) == 0 {
		isGroupIdsEmpty = true
	}
	return isBucketIdsEmpty && isObjectIdsEmpty && isGroupIdsEmpty
}

func (b *InternalBucketInfo) GetLVGByGVGID(gvgID uint32) (*LocalVirtualGroup, bool) {
	for _, lvg := range b.LocalVirtualGroups {
		if lvg.GlobalVirtualGroupId == gvgID {
			return lvg, true
		}
	}
	return nil, false
}

func (b *InternalBucketInfo) AppendLVG(lvg *LocalVirtualGroup) {
	if len(b.LocalVirtualGroups) != 0 {
		lastLVG := b.LocalVirtualGroups[len(b.LocalVirtualGroups)-1]
		if lvg.Id <= lastLVG.Id {
			panic("Not allow to append a lvg which id is smaller than the last lvg")
		}
	}
	b.LocalVirtualGroups = append(b.LocalVirtualGroups, lvg)
}

func (b *InternalBucketInfo) GetMaxLVGID() uint32 {
	if len(b.LocalVirtualGroups) == 0 {
		return 0
	} else {
		lastLVG := b.LocalVirtualGroups[len(b.LocalVirtualGroups)-1]
		return lastLVG.Id
	}
}

func (b *InternalBucketInfo) GetLVG(lvgID uint32) (*LocalVirtualGroup, bool) {
	for _, lvg := range b.LocalVirtualGroups {
		if lvg.Id == lvgID {
			return lvg, true
		}
	}
	return nil, false
}

func (b *InternalBucketInfo) MustGetLVG(lvgID uint32) *LocalVirtualGroup {
	lvg, found := b.GetLVG(lvgID)
	if !found {
		panic("lvg not found in internal bucket info")
	}
	return lvg
}

func (b *InternalBucketInfo) DeleteLVG(lvgID uint32) {
	for i, lvg := range b.LocalVirtualGroups {
		if lvg.Id == lvgID {
			b.LocalVirtualGroups = append(b.LocalVirtualGroups[:i], b.LocalVirtualGroups[i+1:]...)
			break
		}
	}
}
