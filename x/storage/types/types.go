package types

import (
	"fmt"
	"reflect"

	sdkmath "cosmossdk.io/math"
)

const (
	SecondarySPNum = 6
	// MinChargeSize is the minimum size to charge for a storage object
	MinChargeSize = 1024
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

func EncodeSequence(u Uint) []byte {
	return u.Bytes()
}

func DecodeSequence(bz []byte) Uint {
	u := sdkmath.NewUint(0)
	return u.SetBytes(bz)
}

func (m *BucketInfo) ToNFTMetadata() *BucketMetaData {
	return &BucketMetaData{
		BucketName: m.BucketName,
		Attributes: getNFTAttributes(*m),
	}
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
	if di.BucketIds == nil || (di.BucketIds != nil && len(di.BucketIds.Id) == 0) {
		isBucketIdsEmpty = true
	}
	if di.ObjectIds == nil || (di.ObjectIds != nil && len(di.ObjectIds.Id) == 0) {
		isObjectIdsEmpty = true
	}
	if di.GroupIds == nil || (di.GroupIds != nil && len(di.GroupIds.Id) == 0) {
		isGroupIdsEmpty = true
	}
	if isBucketIdsEmpty && isObjectIdsEmpty && isGroupIdsEmpty {
		return true
	}
	return false
}
