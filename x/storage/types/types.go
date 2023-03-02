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

func EncodeSequence(u Uint) []byte {
	return u.Bytes()
}

func DecodeSequence(bz []byte) Uint {
	u := sdkmath.NewUint(0)
	return u.SetBytes(bz)
}

func (m *BucketInfo) ToNFTMetadata() (*MetaData, error) {
	return toNFTMetaData(*m, []string{"PaymentOutFlows"})
}

func (m *ObjectInfo) ToNFTMetadata() (*MetaData, error) {
	return toNFTMetaData(*m, []string{"Checksums"})
}

func (m *GroupInfo) ToNFTMetadata() (*MetaData, error) {
	return toNFTMetaData(*m, nil)
}

func toNFTMetaData(m interface{}, omitFields []string) (*MetaData, error) {
	attributes := make([]Trait, 0)
	v := reflect.ValueOf(m)
	typ := v.Type()

OUTER:
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := typ.Field(i).Name
		for _, n := range omitFields {
			if fieldName == n {
				continue OUTER
			}
		}
		value := fmt.Sprintf("%v", field.Interface())
		attributes = append(attributes,
			Trait{
				TraitType: fieldName,
				Value:     value,
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
