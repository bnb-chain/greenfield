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
	VersionedParams VersionedParams `protobuf:"bytes,1,opt,name=versioned_params,json=versionedParams,proto3" json:"versioned_params"`
	// The maximum number of payment accounts that can be created by one user
	PaymentAccountCountLimit uint64 `protobuf:"varint,2,opt,name=payment_account_count_limit,json=paymentAccountCountLimit,proto3" json:"payment_account_count_limit,omitempty" yaml:"payment_account_count_limit"`
	// Time duration threshold of forced settlement.
	// If dynamic balance is less than NetOutFlowRate * forcedSettleTime, the account can be forced settled.
	ForcedSettleTime uint64 `protobuf:"varint,3,opt,name=forced_settle_time,json=forcedSettleTime,proto3" json:"forced_settle_time,omitempty" yaml:"forced_settle_time"`
	// the maximum number of flows that will be auto forced settled in one block
	MaxAutoSettleFlowCount uint64 `protobuf:"varint,4,opt,name=max_auto_settle_flow_count,json=maxAutoSettleFlowCount,proto3" json:"max_auto_settle_flow_count,omitempty" yaml:"max_auto_settle_flow_count"`
	// the maximum number of flows that will be auto resumed in one block
	MaxAutoResumeFlowCount uint64 `protobuf:"varint,5,opt,name=max_auto_resume_flow_count,json=maxAutoResumeFlowCount,proto3" json:"max_auto_resume_flow_count,omitempty" yaml:"max_auto_resume_flow_count"`
	// The denom of fee charged in payment module
	FeeDenom string `protobuf:"bytes,6,opt,name=fee_denom,json=feeDenom,proto3" json:"fee_denom,omitempty" yaml:"fee_denom"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd7d37632356c8f4, []int{0}
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

func (m *Params) GetVersionedParams() VersionedParams {
	if m != nil {
		return m.VersionedParams
	}
	return VersionedParams{}
}

func (m *Params) GetPaymentAccountCountLimit() uint64 {
	if m != nil {
		return m.PaymentAccountCountLimit
	}
	return 0
}

func (m *Params) GetForcedSettleTime() uint64 {
	if m != nil {
		return m.ForcedSettleTime
	}
	return 0
}

func (m *Params) GetMaxAutoSettleFlowCount() uint64 {
	if m != nil {
		return m.MaxAutoSettleFlowCount
	}
	return 0
}

func (m *Params) GetMaxAutoResumeFlowCount() uint64 {
	if m != nil {
		return m.MaxAutoResumeFlowCount
	}
	return 0
}

func (m *Params) GetFeeDenom() string {
	if m != nil {
		return m.FeeDenom
	}
	return ""
}

// VersionedParams defines the parameters with multiple versions, each version is stored with different timestamp.
type VersionedParams struct {
	// Time duration which the buffer balance need to be reserved for NetOutFlow e.g. 6 month
	ReserveTime uint64 `protobuf:"varint,1,opt,name=reserve_time,json=reserveTime,proto3" json:"reserve_time,omitempty" yaml:"reserve_time"`
	// The tax rate to pay for validators in storage payment. The default value is 1%(0.01)
	ValidatorTaxRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,opt,name=validator_tax_rate,json=validatorTaxRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"validator_tax_rate"`
}

func (m *VersionedParams) Reset()         { *m = VersionedParams{} }
func (m *VersionedParams) String() string { return proto.CompactTextString(m) }
func (*VersionedParams) ProtoMessage()    {}
func (*VersionedParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd7d37632356c8f4, []int{1}
}
func (m *VersionedParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VersionedParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_VersionedParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *VersionedParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VersionedParams.Merge(m, src)
}
func (m *VersionedParams) XXX_Size() int {
	return m.Size()
}
func (m *VersionedParams) XXX_DiscardUnknown() {
	xxx_messageInfo_VersionedParams.DiscardUnknown(m)
}

var xxx_messageInfo_VersionedParams proto.InternalMessageInfo

func (m *VersionedParams) GetReserveTime() uint64 {
	if m != nil {
		return m.ReserveTime
	}
	return 0
}

//func init() {
//	proto.RegisterType((*Params)(nil), "greenfield.payment.Params")
//	proto.RegisterType((*VersionedParams)(nil), "greenfield.payment.VersionedParams")
//}
//
//func init() { proto.RegisterFile("greenfield/payment/params.proto", fileDescriptor_bd7d37632356c8f4) }

var fileDescriptor_bd7d37632356c8f4 = []byte{
	// 509 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0xcd, 0x6e, 0xd3, 0x4c,
	0x14, 0x8d, 0xbf, 0x2f, 0x44, 0x64, 0x8a, 0xd4, 0xc8, 0x54, 0xe0, 0x06, 0x61, 0x07, 0x23, 0xaa,
	0x6c, 0x62, 0x0b, 0xd8, 0x55, 0x6c, 0x6a, 0x22, 0x24, 0x04, 0x0b, 0x64, 0x22, 0x16, 0x6c, 0x46,
	0x13, 0xfb, 0x3a, 0x35, 0x78, 0x3c, 0xd1, 0x78, 0x9c, 0x3a, 0xcf, 0xc0, 0x86, 0x87, 0x61, 0xc3,
	0x1b, 0x74, 0x59, 0xb1, 0x42, 0x2c, 0x2c, 0x94, 0xbc, 0x81, 0x9f, 0x00, 0x65, 0xc6, 0x69, 0x12,
	0xa2, 0xb2, 0x99, 0x9f, 0x7b, 0xce, 0x9c, 0x73, 0x75, 0xe7, 0x5e, 0x64, 0x4d, 0x38, 0x40, 0x1a,
	0xc5, 0x90, 0x84, 0xee, 0x94, 0xcc, 0x29, 0xa4, 0xc2, 0x9d, 0x12, 0x4e, 0x68, 0xe6, 0x4c, 0x39,
	0x13, 0x4c, 0xd7, 0x37, 0x04, 0xa7, 0x26, 0x74, 0x8f, 0x03, 0x96, 0x51, 0x96, 0x61, 0xc9, 0x70,
	0xd5, 0x45, 0xd1, 0xbb, 0x47, 0x13, 0x36, 0x61, 0x2a, 0xbe, 0x3a, 0xa9, 0xa8, 0xfd, 0xa5, 0x89,
	0x5a, 0xef, 0xa4, 0xaa, 0x3e, 0x42, 0x9d, 0x19, 0xf0, 0x2c, 0x66, 0x29, 0x84, 0x58, 0x39, 0x19,
	0x5a, 0x4f, 0xeb, 0x1f, 0x3c, 0x7b, 0xec, 0xec, 0x5b, 0x39, 0x1f, 0xd6, 0x5c, 0xf5, 0xdc, 0x6b,
	0x5e, 0x96, 0x56, 0xc3, 0x3f, 0x9c, 0xed, 0x86, 0x75, 0x40, 0x0f, 0xea, 0x17, 0x98, 0x04, 0x01,
	0xcb, 0x53, 0x81, 0xd5, 0x9a, 0xc4, 0x34, 0x16, 0xc6, 0x7f, 0x3d, 0xad, 0xdf, 0xf4, 0x4e, 0xaa,
	0xd2, 0xb2, 0xe7, 0x84, 0x26, 0xa7, 0xf6, 0x3f, 0xc8, 0xb6, 0x6f, 0xd4, 0xe8, 0x99, 0x02, 0x5f,
	0xae, 0x96, 0xb7, 0x2b, 0x48, 0x7f, 0x83, 0xf4, 0x88, 0xf1, 0x00, 0x42, 0x9c, 0x81, 0x10, 0x09,
	0x60, 0x11, 0x53, 0x30, 0xfe, 0x97, 0xea, 0x0f, 0xab, 0xd2, 0x3a, 0x56, 0xea, 0xfb, 0x1c, 0xdb,
	0xef, 0xa8, 0xe0, 0x7b, 0x19, 0x1b, 0xc5, 0x14, 0x74, 0x82, 0xba, 0x94, 0x14, 0x98, 0xe4, 0x82,
	0xad, 0xa9, 0x51, 0xc2, 0x2e, 0x54, 0x2e, 0x46, 0x53, 0x8a, 0x3e, 0xa9, 0x4a, 0xeb, 0x91, 0x12,
	0xbd, 0x99, 0x6b, 0xfb, 0xf7, 0x28, 0x29, 0xce, 0x72, 0xc1, 0x94, 0xfa, 0xab, 0x84, 0x5d, 0xc8,
	0xa4, 0x77, 0x2c, 0x38, 0x64, 0x39, 0xdd, 0xb1, 0xb8, 0x75, 0xa3, 0xc5, 0x1e, 0x77, 0x63, 0xe1,
	0x4b, 0x68, 0x63, 0xf1, 0x14, 0xb5, 0x23, 0x00, 0x1c, 0x42, 0xca, 0xa8, 0xd1, 0xea, 0x69, 0xfd,
	0xb6, 0x77, 0x54, 0x95, 0x56, 0xa7, 0xae, 0xc4, 0x1a, 0xb2, 0xfd, 0xdb, 0x11, 0xc0, 0x50, 0x1e,
	0xbf, 0x6b, 0xe8, 0xf0, 0xaf, 0x7f, 0xd5, 0x4f, 0xd1, 0x1d, 0x0e, 0x19, 0xf0, 0x59, 0x5d, 0x53,
	0x4d, 0xe6, 0x76, 0xbf, 0x2a, 0xad, 0xbb, 0x4a, 0x69, 0x1b, 0xb5, 0xfd, 0x83, 0xfa, 0x2a, 0x0b,
	0xf9, 0x09, 0xe9, 0x33, 0x92, 0xc4, 0x21, 0x11, 0x8c, 0x63, 0x41, 0x0a, 0xcc, 0x89, 0x00, 0xf9,
	0xe7, 0x6d, 0xef, 0xc5, 0xaa, 0x5f, 0x7e, 0x95, 0xd6, 0xc9, 0x24, 0x16, 0xe7, 0xf9, 0xd8, 0x09,
	0x18, 0xad, 0x1b, 0xb6, 0xde, 0x06, 0x59, 0xf8, 0xd9, 0x15, 0xf3, 0x29, 0x64, 0xce, 0x10, 0x82,
	0x1f, 0xdf, 0x06, 0xa8, 0xee, 0xe7, 0x21, 0x04, 0x7e, 0xe7, 0x5a, 0x77, 0x44, 0x0a, 0x9f, 0x08,
	0xf0, 0x5e, 0x5f, 0x2e, 0x4c, 0xed, 0x6a, 0x61, 0x6a, 0xbf, 0x17, 0xa6, 0xf6, 0x75, 0x69, 0x36,
	0xae, 0x96, 0x66, 0xe3, 0xe7, 0xd2, 0x6c, 0x7c, 0x74, 0xb7, 0x1c, 0xc6, 0xe9, 0x78, 0x10, 0x9c,
	0x93, 0x38, 0x75, 0xb7, 0xc6, 0xab, 0xb8, 0x1e, 0x30, 0x69, 0x37, 0x6e, 0xc9, 0xd9, 0x78, 0xfe,
	0x27, 0x00, 0x00, 0xff, 0xff, 0xad, 0xa0, 0x65, 0x82, 0x83, 0x03, 0x00, 0x00,
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
	if len(m.FeeDenom) > 0 {
		i -= len(m.FeeDenom)
		copy(dAtA[i:], m.FeeDenom)
		i = encodeVarintParams(dAtA, i, uint64(len(m.FeeDenom)))
		i--
		dAtA[i] = 0x32
	}
	if m.MaxAutoResumeFlowCount != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxAutoResumeFlowCount))
		i--
		dAtA[i] = 0x28
	}
	if m.MaxAutoSettleFlowCount != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxAutoSettleFlowCount))
		i--
		dAtA[i] = 0x20
	}
	if m.ForcedSettleTime != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.ForcedSettleTime))
		i--
		dAtA[i] = 0x18
	}
	if m.PaymentAccountCountLimit != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.PaymentAccountCountLimit))
		i--
		dAtA[i] = 0x10
	}
	{
		size, err := m.VersionedParams.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *VersionedParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VersionedParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VersionedParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.ValidatorTaxRate.Size()
		i -= size
		if _, err := m.ValidatorTaxRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if m.ReserveTime != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.ReserveTime))
		i--
		dAtA[i] = 0x8
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
	l = m.VersionedParams.Size()
	n += 1 + l + sovParams(uint64(l))
	if m.PaymentAccountCountLimit != 0 {
		n += 1 + sovParams(uint64(m.PaymentAccountCountLimit))
	}
	if m.ForcedSettleTime != 0 {
		n += 1 + sovParams(uint64(m.ForcedSettleTime))
	}
	if m.MaxAutoSettleFlowCount != 0 {
		n += 1 + sovParams(uint64(m.MaxAutoSettleFlowCount))
	}
	if m.MaxAutoResumeFlowCount != 0 {
		n += 1 + sovParams(uint64(m.MaxAutoResumeFlowCount))
	}
	l = len(m.FeeDenom)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	return n
}

func (m *VersionedParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ReserveTime != 0 {
		n += 1 + sovParams(uint64(m.ReserveTime))
	}
	l = m.ValidatorTaxRate.Size()
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
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
				return fmt.Errorf("proto: wrong wireType = %d for field VersionedParams", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.VersionedParams.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PaymentAccountCountLimit", wireType)
			}
			m.PaymentAccountCountLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PaymentAccountCountLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ForcedSettleTime", wireType)
			}
			m.ForcedSettleTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ForcedSettleTime |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxAutoSettleFlowCount", wireType)
			}
			m.MaxAutoSettleFlowCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxAutoSettleFlowCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxAutoResumeFlowCount", wireType)
			}
			m.MaxAutoResumeFlowCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxAutoResumeFlowCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeeDenom", wireType)
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
			m.FeeDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
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
func (m *VersionedParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: VersionedParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VersionedParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReserveTime", wireType)
			}
			m.ReserveTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReserveTime |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorTaxRate", wireType)
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
			if err := m.ValidatorTaxRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
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
