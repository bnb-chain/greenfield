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

func EncodeSequence(u Uint) []byte {
	return u.Bytes()
}

func DecodeSequence(bz []byte) Uint {
	u := sdkmath.NewUint(0)
	return u.SetBytes(bz)
}

func (m *BucketInfo) ToNFTMetadata() (*MetaData, error) {
	return toNFTMetaData(*m)
}

func (m *ObjectInfo) ToNFTMetadata() (*MetaData, error) {
	return toNFTMetaData(*m)
}

func (m *GroupInfo) ToNFTMetadata() (*MetaData, error) {
	return toNFTMetaData(*m)
}

func toNFTMetaData(m interface{}) (*MetaData, error) {
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
	name, err := getNFTName(m)
	if err != nil {
		return nil, err
	}
	return &MetaData{
		Name:       name,
		Attributes: attributes,
	}, nil
}

func getNFTName(m interface{}) (isMetaData_Name, error) {
	switch o := m.(type) {
	case BucketInfo:
		return &MetaData_BucketName{o.BucketName}, nil
	case ObjectInfo:
		return &MetaData_ObjectName{o.ObjectName}, nil
	case GroupInfo:
		return &MetaData_GroupName{o.GroupName}, nil
	default:
		return nil, ErrInvalidNFTType
	}
}
