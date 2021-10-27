package main

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Timer struct {
	Enable               bool     `protobuf:"varint,1,opt,name=enable,proto3" json:"enable,omitempty"`
	Alert                string   `protobuf:"bytes,2,opt,name=alert,proto3" json:"alert,omitempty"`
	Url                  string   `protobuf:"bytes,4,opt,name=url,proto3" json:"url,omitempty"`
	Month                int32    `protobuf:"varint,8,opt,name=month,proto3" json:"month,omitempty"`
	Day                  int32    `protobuf:"varint,16,opt,name=day,proto3" json:"day,omitempty"`
	Week                 int32    `protobuf:"varint,32,opt,name=week,proto3" json:"week,omitempty"`
	Hour                 int32    `protobuf:"varint,64,opt,name=hour,proto3" json:"hour,omitempty"`
	Minute               int32    `protobuf:"varint,128,opt,name=minute,proto3" json:"minute,omitempty"`
	Selfid               int64    `protobuf:"varint,129,opt,name=selfid,proto3" json:"selfid,omitempty"`
	Cron                 string   `protobuf:"bytes,130,opt,name=cron,proto3" json:"cron,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Timer) Reset()         { *m = Timer{} }
func (m *Timer) String() string { return proto.CompactTextString(m) }
func (*Timer) ProtoMessage()    {}
func (*Timer) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad0307ee16b652d2, []int{0}
}
func (m *Timer) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Timer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Timer.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Timer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Timer.Merge(m, src)
}
func (m *Timer) XXX_Size() int {
	return m.Size()
}
func (m *Timer) XXX_DiscardUnknown() {
	xxx_messageInfo_Timer.DiscardUnknown(m)
}

var xxx_messageInfo_Timer proto.InternalMessageInfo

func (m *Timer) GetEnable() bool {
	if m != nil {
		return m.Enable
	}
	return false
}

func (m *Timer) GetAlert() string {
	if m != nil {
		return m.Alert
	}
	return ""
}

func (m *Timer) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *Timer) GetMonth() int32 {
	if m != nil {
		return m.Month
	}
	return 0
}

func (m *Timer) GetDay() int32 {
	if m != nil {
		return m.Day
	}
	return 0
}

func (m *Timer) GetWeek() int32 {
	if m != nil {
		return m.Week
	}
	return 0
}

func (m *Timer) GetHour() int32 {
	if m != nil {
		return m.Hour
	}
	return 0
}

func (m *Timer) GetMinute() int32 {
	if m != nil {
		return m.Minute
	}
	return 0
}

func (m *Timer) GetSelfid() int64 {
	if m != nil {
		return m.Selfid
	}
	return 0
}

func (m *Timer) GetCron() string {
	if m != nil {
		return m.Cron
	}
	return ""
}

type TimersMap struct {
	Timers               map[string]*Timer `protobuf:"bytes,1,rep,name=timers,proto3" json:"timers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *TimersMap) Reset()         { *m = TimersMap{} }
func (m *TimersMap) String() string { return proto.CompactTextString(m) }
func (*TimersMap) ProtoMessage()    {}
func (*TimersMap) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad0307ee16b652d2, []int{1}
}
func (m *TimersMap) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TimersMap) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TimersMap.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TimersMap) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TimersMap.Merge(m, src)
}
func (m *TimersMap) XXX_Size() int {
	return m.Size()
}
func (m *TimersMap) XXX_DiscardUnknown() {
	xxx_messageInfo_TimersMap.DiscardUnknown(m)
}

var xxx_messageInfo_TimersMap proto.InternalMessageInfo

func (m *TimersMap) GetTimers() map[string]*Timer {
	if m != nil {
		return m.Timers
	}
	return nil
}

func init() {
	proto.RegisterType((*Timer)(nil), "timer.Timer")
	proto.RegisterType((*TimersMap)(nil), "timer.TimersMap")
	proto.RegisterMapType((map[string]*Timer)(nil), "timer.TimersMap.TimersEntry")
}

func init() { proto.RegisterFile("timer.proto", fileDescriptor_ad0307ee16b652d2) }

var fileDescriptor_ad0307ee16b652d2 = []byte{
	// 293 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0x9d, 0xb6, 0x09, 0xed, 0xc4, 0x43, 0x58, 0x45, 0x07, 0x91, 0x10, 0x72, 0xca, 0xa9,
	0x87, 0xea, 0x41, 0x3c, 0x89, 0x20, 0x9e, 0xbc, 0x2c, 0xbe, 0x40, 0x6a, 0x57, 0x1a, 0x9a, 0x3f,
	0x65, 0xb3, 0x51, 0x72, 0x53, 0x8f, 0x3e, 0x81, 0x8f, 0xe4, 0xd1, 0x27, 0x10, 0x89, 0x2f, 0x22,
	0x3b, 0x1b, 0xa4, 0xb7, 0xef, 0xf7, 0x9b, 0x8f, 0x24, 0x33, 0xc1, 0xc0, 0xe4, 0xa5, 0xd2, 0xf3,
	0xad, 0xae, 0x4d, 0x2d, 0x3c, 0x86, 0xe4, 0x1b, 0xd0, 0xbb, 0xb7, 0x49, 0x1c, 0xa1, 0xaf, 0xaa,
	0x6c, 0x59, 0x28, 0x82, 0x18, 0xd2, 0xa9, 0x1c, 0x48, 0x1c, 0xa2, 0x97, 0x15, 0x4a, 0x1b, 0x1a,
	0xc5, 0x90, 0xce, 0xa4, 0x03, 0x11, 0xe2, 0xb8, 0xd5, 0x05, 0x4d, 0xd8, 0xd9, 0x68, 0x7b, 0x65,
	0x5d, 0x99, 0x35, 0x4d, 0x63, 0x48, 0x3d, 0xe9, 0xc0, 0xf6, 0x56, 0x59, 0x47, 0x21, 0x3b, 0x1b,
	0x85, 0xc0, 0xc9, 0xb3, 0x52, 0x1b, 0x8a, 0x59, 0x71, 0xb6, 0x6e, 0x5d, 0xb7, 0x9a, 0xae, 0x9c,
	0xb3, 0x59, 0x1c, 0xa3, 0x5f, 0xe6, 0x55, 0x6b, 0x14, 0xbd, 0x00, 0xeb, 0x01, 0xed, 0xa0, 0x51,
	0xc5, 0x63, 0xbe, 0xa2, 0x57, 0x3b, 0x18, 0xcb, 0x01, 0xc5, 0x01, 0x4e, 0x1e, 0x74, 0x5d, 0xd1,
	0x1b, 0xf0, 0x57, 0x31, 0x24, 0xef, 0x80, 0x33, 0x5e, 0xb0, 0xb9, 0xcb, 0xb6, 0xe2, 0x1c, 0x7d,
	0xde, 0xbb, 0x21, 0x88, 0xc7, 0x69, 0xb0, 0x38, 0x9d, 0xbb, 0x9b, 0xfc, 0x37, 0x86, 0x74, 0x53,
	0x19, 0xdd, 0xc9, 0xa1, 0x7b, 0x72, 0x8b, 0xc1, 0x8e, 0xb6, 0x3b, 0x6d, 0x54, 0x47, 0xee, 0x2d,
	0x36, 0x8a, 0x04, 0xbd, 0xa7, 0xac, 0x68, 0x15, 0xdf, 0x28, 0x58, 0xec, 0xef, 0x3e, 0x55, 0xba,
	0xd1, 0xe5, 0xe8, 0x02, 0xae, 0xc3, 0xcf, 0x3e, 0x82, 0xaf, 0x3e, 0x82, 0x9f, 0x3e, 0x82, 0x8f,
	0xdf, 0x68, 0x6f, 0xe9, 0xf3, 0xdf, 0x38, 0xfb, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x04, 0x03, 0x99,
	0x28, 0x9c, 0x01, 0x00, 0x00,
}

func (m *Timer) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Timer) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Timer) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Cron) > 0 {
		i -= len(m.Cron)
		copy(dAtA[i:], m.Cron)
		i = encodeVarintTimer(dAtA, i, uint64(len(m.Cron)))
		i--
		dAtA[i] = 0x8
		i--
		dAtA[i] = 0x92
	}
	if m.Selfid != 0 {
		i = encodeVarintTimer(dAtA, i, uint64(m.Selfid))
		i--
		dAtA[i] = 0x8
		i--
		dAtA[i] = 0x88
	}
	if m.Minute != 0 {
		i = encodeVarintTimer(dAtA, i, uint64(m.Minute))
		i--
		dAtA[i] = 0x8
		i--
		dAtA[i] = 0x80
	}
	if m.Hour != 0 {
		i = encodeVarintTimer(dAtA, i, uint64(m.Hour))
		i--
		dAtA[i] = 0x4
		i--
		dAtA[i] = 0x80
	}
	if m.Week != 0 {
		i = encodeVarintTimer(dAtA, i, uint64(m.Week))
		i--
		dAtA[i] = 0x2
		i--
		dAtA[i] = 0x80
	}
	if m.Day != 0 {
		i = encodeVarintTimer(dAtA, i, uint64(m.Day))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x80
	}
	if m.Month != 0 {
		i = encodeVarintTimer(dAtA, i, uint64(m.Month))
		i--
		dAtA[i] = 0x40
	}
	if len(m.Url) > 0 {
		i -= len(m.Url)
		copy(dAtA[i:], m.Url)
		i = encodeVarintTimer(dAtA, i, uint64(len(m.Url)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Alert) > 0 {
		i -= len(m.Alert)
		copy(dAtA[i:], m.Alert)
		i = encodeVarintTimer(dAtA, i, uint64(len(m.Alert)))
		i--
		dAtA[i] = 0x12
	}
	if m.Enable {
		i--
		if m.Enable {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *TimersMap) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TimersMap) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TimersMap) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Timers) > 0 {
		for k := range m.Timers {
			v := m.Timers[k]
			baseI := i
			if v != nil {
				{
					size, err := v.MarshalToSizedBuffer(dAtA[:i])
					if err != nil {
						return 0, err
					}
					i -= size
					i = encodeVarintTimer(dAtA, i, uint64(size))
				}
				i--
				dAtA[i] = 0x12
			}
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintTimer(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintTimer(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintTimer(dAtA []byte, offset int, v uint64) int {
	offset -= sovTimer(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Timer) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Enable {
		n += 2
	}
	l = len(m.Alert)
	if l > 0 {
		n += 1 + l + sovTimer(uint64(l))
	}
	l = len(m.Url)
	if l > 0 {
		n += 1 + l + sovTimer(uint64(l))
	}
	if m.Month != 0 {
		n += 1 + sovTimer(uint64(m.Month))
	}
	if m.Day != 0 {
		n += 2 + sovTimer(uint64(m.Day))
	}
	if m.Week != 0 {
		n += 2 + sovTimer(uint64(m.Week))
	}
	if m.Hour != 0 {
		n += 2 + sovTimer(uint64(m.Hour))
	}
	if m.Minute != 0 {
		n += 2 + sovTimer(uint64(m.Minute))
	}
	if m.Selfid != 0 {
		n += 2 + sovTimer(uint64(m.Selfid))
	}
	l = len(m.Cron)
	if l > 0 {
		n += 2 + l + sovTimer(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *TimersMap) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Timers) > 0 {
		for k, v := range m.Timers {
			_ = k
			_ = v
			l = 0
			if v != nil {
				l = v.Size()
				l += 1 + sovTimer(uint64(l))
			}
			mapEntrySize := 1 + len(k) + sovTimer(uint64(len(k))) + l
			n += mapEntrySize + 1 + sovTimer(uint64(mapEntrySize))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovTimer(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTimer(x uint64) (n int) {
	return sovTimer(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Timer) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTimer
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
			return fmt.Errorf("proto: Timer: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Timer: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Enable", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Enable = bool(v != 0)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Alert", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
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
				return ErrInvalidLengthTimer
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTimer
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Alert = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Url", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
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
				return ErrInvalidLengthTimer
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTimer
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Url = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Month", wireType)
			}
			m.Month = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Month |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 16:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Day", wireType)
			}
			m.Day = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Day |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 32:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Week", wireType)
			}
			m.Week = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Week |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 64:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hour", wireType)
			}
			m.Hour = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Hour |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 128:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Minute", wireType)
			}
			m.Minute = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Minute |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 129:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Selfid", wireType)
			}
			m.Selfid = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Selfid |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 130:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cron", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
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
				return ErrInvalidLengthTimer
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTimer
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cron = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTimer(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTimer
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TimersMap) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTimer
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
			return fmt.Errorf("proto: TimersMap: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TimersMap: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timers", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimer
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
				return ErrInvalidLengthTimer
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTimer
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Timers == nil {
				m.Timers = make(map[string]*Timer)
			}
			var mapkey string
			var mapvalue *Timer
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTimer
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
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTimer
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthTimer
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthTimer
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTimer
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthTimer
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return ErrInvalidLengthTimer
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &Timer{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipTimer(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthTimer
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Timers[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTimer(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTimer
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTimer(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTimer
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
					return 0, ErrIntOverflowTimer
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
					return 0, ErrIntOverflowTimer
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
				return 0, ErrInvalidLengthTimer
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTimer
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTimer
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTimer        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTimer          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTimer = fmt.Errorf("proto: unexpected end of group")
)
