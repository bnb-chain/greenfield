package v1

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	"gopkg.in/yaml.v2"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params defines the parameters for the module.
type Params struct {
	// deposit_denom defines the staking coin denomination.
	DepositDenom string `protobuf:"bytes,1,opt,name=deposit_denom,json=depositDenom,proto3" json:"deposit_denom,omitempty"`
	// store price, in bnb wei per charge byte
	//nolint
	GvgStakingPerBytes github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=gvg_staking_per_bytes,json=gvgStakingPerBytes,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"gvg_staking_per_bytes"`
	// the max number of lvg which allowed in a bucket
	MaxLocalVirtualGroupNumPerBucket uint32 `protobuf:"varint,3,opt,name=max_local_virtual_group_num_per_bucket,json=maxLocalVirtualGroupNumPerBucket,proto3" json:"max_local_virtual_group_num_per_bucket,omitempty"`
	// the max number of gvg which can exist in a family
	MaxGlobalVirtualGroupNumPerFamily uint32 `protobuf:"varint,4,opt,name=max_global_virtual_group_num_per_family,json=maxGlobalVirtualGroupNumPerFamily,proto3" json:"max_global_virtual_group_num_per_family,omitempty"`
	// if the store size reach the exceed, the family is not allowed to sever more buckets
	MaxStoreSizePerFamily uint64 `protobuf:"varint,5,opt,name=max_store_size_per_family,json=maxStoreSizePerFamily,proto3" json:"max_store_size_per_family,omitempty"`
}

func (m *Params) String() string {
	out, _ := yaml.Marshal(m)
	return string(out)
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_d8ecf89dd5128885, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetDepositDenom() string {
	if m != nil {
		return m.DepositDenom
	}
	return ""
}

func (m *Params) GetMaxLocalVirtualGroupNumPerBucket() uint32 {
	if m != nil {
		return m.MaxLocalVirtualGroupNumPerBucket
	}
	return 0
}

func (m *Params) GetMaxGlobalVirtualGroupNumPerFamily() uint32 {
	if m != nil {
		return m.MaxGlobalVirtualGroupNumPerFamily
	}
	return 0
}

func (m *Params) GetMaxStoreSizePerFamily() uint64 {
	if m != nil {
		return m.MaxStoreSizePerFamily
	}
	return 0
}

//func init() {
//	proto.RegisterType((*Params)(nil), "greenfield.virtualgroup.Params")
//}
//
//func init() {
//	proto.RegisterFile("greenfield/virtualgroup/params.proto", fileDescriptor_d8ecf89dd5128885)
//}

var fileDescriptor_d8ecf89dd5128885 = []byte{
	// 414 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0x3f, 0x6f, 0xd4, 0x30,
	0x18, 0x87, 0x63, 0x5a, 0x2a, 0xb0, 0xe8, 0x12, 0x51, 0x91, 0x76, 0x48, 0xc2, 0x1f, 0x95, 0x5b,
	0x2e, 0x19, 0x60, 0x40, 0x88, 0xe9, 0x84, 0xa8, 0x2a, 0xa1, 0xea, 0x94, 0x93, 0x18, 0x58, 0x2c,
	0x27, 0x71, 0x5d, 0xeb, 0x62, 0x3b, 0xb2, 0x9d, 0x53, 0xae, 0x9f, 0x80, 0x91, 0x91, 0xb1, 0x1f,
	0x82, 0x0f, 0x71, 0xe3, 0x89, 0x09, 0x31, 0x9c, 0xd0, 0xdd, 0xc2, 0xc7, 0x40, 0x76, 0x22, 0x08,
	0x42, 0x9d, 0x92, 0xfc, 0xf2, 0xf8, 0xd1, 0xfb, 0xbe, 0x7e, 0xe1, 0x33, 0xaa, 0x08, 0x11, 0x97,
	0x8c, 0x54, 0x65, 0xba, 0x60, 0xca, 0x34, 0xb8, 0xa2, 0x4a, 0x36, 0x75, 0x5a, 0x63, 0x85, 0xb9,
	0x4e, 0x6a, 0x25, 0x8d, 0xf4, 0x1f, 0xfd, 0xa5, 0x92, 0x21, 0x75, 0x72, 0x5c, 0x48, 0xcd, 0xa5,
	0x46, 0x0e, 0x4b, 0xbb, 0x8f, 0xee, 0xcc, 0xc9, 0x43, 0x2a, 0xa9, 0xec, 0x72, 0xfb, 0xd6, 0xa5,
	0x4f, 0x3e, 0xed, 0xc1, 0x83, 0xa9, 0x53, 0xfb, 0x4f, 0xe1, 0x61, 0x49, 0x6a, 0xa9, 0x99, 0x41,
	0x25, 0x11, 0x92, 0x07, 0x20, 0x06, 0xa3, 0xfb, 0xd9, 0x83, 0x3e, 0x7c, 0x6b, 0x33, 0x5f, 0xc2,
	0x23, 0xba, 0xa0, 0x48, 0x1b, 0x3c, 0x67, 0x82, 0xa2, 0x9a, 0x28, 0x94, 0x2f, 0x0d, 0xd1, 0xc1,
	0x1d, 0x0b, 0x4f, 0xde, 0xac, 0x36, 0x91, 0xf7, 0x63, 0x13, 0x9d, 0x52, 0x66, 0xae, 0x9a, 0x3c,
	0x29, 0x24, 0xef, 0xab, 0xe8, 0x1f, 0x63, 0x5d, 0xce, 0x53, 0xb3, 0xac, 0x89, 0x4e, 0xce, 0x85,
	0xf9, 0xf6, 0x75, 0x0c, 0xfb, 0x22, 0xcf, 0x85, 0xc9, 0x7c, 0xba, 0xa0, 0xb3, 0xce, 0x3c, 0x25,
	0x6a, 0x62, 0xbd, 0xfe, 0x14, 0x9e, 0x72, 0xdc, 0xa2, 0x4a, 0x16, 0xb8, 0x42, 0x7d, 0xaf, 0xc8,
	0x35, 0x8b, 0x44, 0xc3, 0xbb, 0x02, 0x9a, 0x62, 0x4e, 0x4c, 0xb0, 0x17, 0x83, 0xd1, 0x61, 0x16,
	0x73, 0xdc, 0xbe, 0xb7, 0xf0, 0x87, 0x8e, 0x3d, 0xb3, 0xe8, 0x45, 0xc3, 0xad, 0xd0, 0x71, 0x7e,
	0x06, 0x9f, 0x5b, 0x23, 0xad, 0x64, 0x7e, 0xab, 0xf2, 0x12, 0x73, 0x56, 0x2d, 0x83, 0x7d, 0xa7,
	0x7c, 0xcc, 0x71, 0x7b, 0xe6, 0xe8, 0xff, 0x9d, 0xef, 0x1c, 0xe8, 0xbf, 0x82, 0xc7, 0xd6, 0xa9,
	0x8d, 0x54, 0x04, 0x69, 0x76, 0x4d, 0x86, 0x96, 0xbb, 0x31, 0x18, 0xed, 0x67, 0x47, 0x1c, 0xb7,
	0x33, 0xfb, 0x7f, 0xc6, 0xae, 0xc9, 0x9f, 0x93, 0xaf, 0xef, 0x7d, 0xb9, 0x89, 0xbc, 0x5f, 0x37,
	0x11, 0x98, 0x5c, 0xac, 0xb6, 0x21, 0x58, 0x6f, 0x43, 0xf0, 0x73, 0x1b, 0x82, 0xcf, 0xbb, 0xd0,
	0x5b, 0xef, 0x42, 0xef, 0xfb, 0x2e, 0xf4, 0x3e, 0xbe, 0x1c, 0x4c, 0x33, 0x17, 0xf9, 0xb8, 0xb8,
	0xc2, 0x4c, 0xa4, 0x83, 0x4d, 0x69, 0xff, 0xdd, 0x15, 0x37, 0xdf, 0xfc, 0xc0, 0xdd, 0xf0, 0x8b,
	0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x12, 0x99, 0xe2, 0xe9, 0x53, 0x02, 0x00, 0x00,
}

func (this *Params) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.DepositDenom != that1.DepositDenom {
		return false
	}
	if !this.GvgStakingPerBytes.Equal(that1.GvgStakingPerBytes) {
		return false
	}
	if this.MaxLocalVirtualGroupNumPerBucket != that1.MaxLocalVirtualGroupNumPerBucket {
		return false
	}
	if this.MaxGlobalVirtualGroupNumPerFamily != that1.MaxGlobalVirtualGroupNumPerFamily {
		return false
	}
	if this.MaxStoreSizePerFamily != that1.MaxStoreSizePerFamily {
		return false
	}
	return true
}
func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.MaxStoreSizePerFamily != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxStoreSizePerFamily))
		i--
		dAtA[i] = 0x28
	}
	if m.MaxGlobalVirtualGroupNumPerFamily != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxGlobalVirtualGroupNumPerFamily))
		i--
		dAtA[i] = 0x20
	}
	if m.MaxLocalVirtualGroupNumPerBucket != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxLocalVirtualGroupNumPerBucket))
		i--
		dAtA[i] = 0x18
	}
	{
		size := m.GvgStakingPerBytes.Size()
		i -= size
		if _, err := m.GvgStakingPerBytes.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.DepositDenom) > 0 {
		i -= len(m.DepositDenom)
		copy(dAtA[i:], m.DepositDenom)
		i = encodeVarintParams(dAtA, i, uint64(len(m.DepositDenom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DepositDenom)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	l = m.GvgStakingPerBytes.Size()
	n += 1 + l + sovParams(uint64(l))
	if m.MaxLocalVirtualGroupNumPerBucket != 0 {
		n += 1 + sovParams(uint64(m.MaxLocalVirtualGroupNumPerBucket))
	}
	if m.MaxGlobalVirtualGroupNumPerFamily != 0 {
		n += 1 + sovParams(uint64(m.MaxGlobalVirtualGroupNumPerFamily))
	}
	if m.MaxStoreSizePerFamily != 0 {
		n += 1 + sovParams(uint64(m.MaxStoreSizePerFamily))
	}
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

//lint:ignore U1000 Ignore unused function for it is auto generated ealier
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DepositDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DepositDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GvgStakingPerBytes", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.GvgStakingPerBytes.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxLocalVirtualGroupNumPerBucket", wireType)
			}
			m.MaxLocalVirtualGroupNumPerBucket = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxLocalVirtualGroupNumPerBucket |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxGlobalVirtualGroupNumPerFamily", wireType)
			}
			m.MaxGlobalVirtualGroupNumPerFamily = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxGlobalVirtualGroupNumPerFamily |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxStoreSizePerFamily", wireType)
			}
			m.MaxStoreSizePerFamily = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxStoreSizePerFamily |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowParams
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowParams
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
