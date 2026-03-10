package types

import (
	"context"
	"fmt"
	"io"
	math_bits "math/bits"

	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

var _ proto.Message = &MsgDeleteOwnRecord{}

type MsgDeleteOwnRecord struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *MsgDeleteOwnRecord) Reset()         { *m = MsgDeleteOwnRecord{} }
func (m *MsgDeleteOwnRecord) String() string { return proto.CompactTextString(m) }
func (*MsgDeleteOwnRecord) ProtoMessage()    {}

func (m *MsgDeleteOwnRecord) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type MsgDeleteOwnRecordResponse struct{}

func (m *MsgDeleteOwnRecordResponse) Reset()         { *m = MsgDeleteOwnRecordResponse{} }
func (m *MsgDeleteOwnRecordResponse) String() string { return proto.CompactTextString(m) }
func (*MsgDeleteOwnRecordResponse) ProtoMessage()    {}

// --- MsgServer ---

type MsgServer interface {
	DeleteOwnRecord(context.Context, *MsgDeleteOwnRecord) (*MsgDeleteOwnRecordResponse, error)
}

type UnimplementedMsgServer struct{}

func (*UnimplementedMsgServer) DeleteOwnRecord(context.Context, *MsgDeleteOwnRecord) (*MsgDeleteOwnRecordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteOwnRecord not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_DeleteOwnRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgDeleteOwnRecord)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).DeleteOwnRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/portalchain.constitution.Msg/DeleteOwnRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).DeleteOwnRecord(ctx, req.(*MsgDeleteOwnRecord))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "portalchain.constitution.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DeleteOwnRecord",
			Handler:    _Msg_DeleteOwnRecord_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "portalchain/constitution/tx.proto",
}

// --- MsgClient ---

type MsgClient interface {
	DeleteOwnRecord(ctx context.Context, in *MsgDeleteOwnRecord, opts ...grpc.CallOption) (*MsgDeleteOwnRecordResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) DeleteOwnRecord(ctx context.Context, in *MsgDeleteOwnRecord, opts ...grpc.CallOption) (*MsgDeleteOwnRecordResponse, error) {
	out := new(MsgDeleteOwnRecordResponse)
	err := c.cc.Invoke(ctx, "/portalchain.constitution.Msg/DeleteOwnRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// --- Proto Registration ---

func init() {
	proto.RegisterType((*MsgDeleteOwnRecord)(nil), "portalchain.constitution.MsgDeleteOwnRecord")
	proto.RegisterType((*MsgDeleteOwnRecordResponse)(nil), "portalchain.constitution.MsgDeleteOwnRecordResponse")
}

// --- Binary Marshal/Unmarshal ---

func (m *MsgDeleteOwnRecord) Marshal() ([]byte, error) {
	size := m.Size()
	dAtA := make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgDeleteOwnRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgDeleteOwnRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarint(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0x0a
	}
	return len(dAtA) - i, nil
}

func (m *MsgDeleteOwnRecord) Size() int {
	if m == nil {
		return 0
	}
	var n int
	l := len(m.Address)
	if l > 0 {
		n += 1 + l + sov(uint64(l))
	}
	return n
}

func (m *MsgDeleteOwnRecord) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return errIntOverflow
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
			return fmt.Errorf("proto: MsgDeleteOwnRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgDeleteOwnRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return errIntOverflow
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
				return errInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return errInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return errInvalidLength
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

func (m *MsgDeleteOwnRecordResponse) Marshal() ([]byte, error) {
	size := m.Size()
	dAtA := make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgDeleteOwnRecordResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgDeleteOwnRecordResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	return len(dAtA) - i, nil
}

func (m *MsgDeleteOwnRecordResponse) Size() int {
	if m == nil {
		return 0
	}
	return 0
}

func (m *MsgDeleteOwnRecordResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return errIntOverflow
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
			return fmt.Errorf("proto: MsgDeleteOwnRecordResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgDeleteOwnRecordResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return errInvalidLength
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

// --- Helpers ---

func encodeVarint(dAtA []byte, offset int, v uint64) int {
	offset -= sov(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func sov(x uint64) int {
	return (math_bits.Len64(x|1) + 6) / 7
}

func skip(dAtA []byte) (int, error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, errIntOverflow
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
					return 0, errIntOverflow
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
					return 0, errIntOverflow
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
				return 0, errInvalidLength
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, errUnexpectedEndOfGroup
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, errInvalidLength
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	errInvalidLength        = fmt.Errorf("proto: negative length found during unmarshaling")
	errIntOverflow          = fmt.Errorf("proto: integer overflow")
	errUnexpectedEndOfGroup = fmt.Errorf("proto: unexpected end of group")
)
