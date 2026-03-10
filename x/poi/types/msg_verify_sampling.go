package types

import (
	"context"
	"fmt"
	"io"
	math_bits "math/bits"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// MsgVerifySampling — proto-compatible message
// ---------------------------------------------------------------------------

const TypeMsgVerifySampling = "verify_sampling"

var _ sdk.Msg = &MsgVerifySampling{}

type MsgVerifySampling struct {
	Verifier  string `protobuf:"bytes,1,opt,name=verifier,proto3" json:"verifier,omitempty"`
	Epoch     int64  `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Validator string `protobuf:"bytes,3,opt,name=validator,proto3" json:"validator,omitempty"`
	Passed    bool   `protobuf:"varint,4,opt,name=passed,proto3" json:"passed,omitempty"`
}

type MsgVerifySamplingResponse struct{}

func init() {
	proto.RegisterType((*MsgVerifySampling)(nil), "portalchain.poi.MsgVerifySampling")
	proto.RegisterType((*MsgVerifySamplingResponse)(nil), "portalchain.poi.MsgVerifySamplingResponse")
}

// ---------------------------------------------------------------------------
// proto.Message interface
// ---------------------------------------------------------------------------

func (m *MsgVerifySampling) Reset()         { *m = MsgVerifySampling{} }
func (m *MsgVerifySampling) String() string { return proto.CompactTextString(m) }
func (*MsgVerifySampling) ProtoMessage()    {}

func (m *MsgVerifySamplingResponse) Reset()         { *m = MsgVerifySamplingResponse{} }
func (m *MsgVerifySamplingResponse) String() string { return proto.CompactTextString(m) }
func (*MsgVerifySamplingResponse) ProtoMessage()    {}

// ---------------------------------------------------------------------------
// sdk.Msg interface
// ---------------------------------------------------------------------------

func (msg *MsgVerifySampling) Route() string { return RouterKey }
func (msg *MsgVerifySampling) Type() string  { return TypeMsgVerifySampling }

func (msg *MsgVerifySampling) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Verifier)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgVerifySampling) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgVerifySampling) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Verifier); err != nil {
		return sdkerrors.Wrapf(ErrInvalidValidator, "invalid verifier address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return sdkerrors.Wrapf(ErrInvalidValidator, "invalid validator address: %s", err)
	}
	if msg.Epoch <= 0 {
		return sdkerrors.Wrap(ErrNegativeValue, "epoch must be positive")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Proto binary serialization
// ---------------------------------------------------------------------------

func (m *MsgVerifySampling) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgVerifySampling) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgVerifySampling) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Passed {
		i--
		if m.Passed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x20
	}
	if len(m.Validator) > 0 {
		i -= len(m.Validator)
		copy(dAtA[i:], m.Validator)
		i = encodeVarintVS(dAtA, i, uint64(len(m.Validator)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Epoch != 0 {
		i = encodeVarintVS(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Verifier) > 0 {
		i -= len(m.Verifier)
		copy(dAtA[i:], m.Verifier)
		i = encodeVarintVS(dAtA, i, uint64(len(m.Verifier)))
		i--
		dAtA[i] = 0x0a
	}
	return len(dAtA) - i, nil
}

func (m *MsgVerifySampling) Size() (n int) {
	if m == nil {
		return 0
	}
	l := len(m.Verifier)
	if l > 0 {
		n += 1 + l + sovVS(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovVS(uint64(m.Epoch))
	}
	l = len(m.Validator)
	if l > 0 {
		n += 1 + l + sovVS(uint64(l))
	}
	if m.Passed {
		n += 1 + 1
	}
	return n
}

func (m *MsgVerifySampling) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgVerifySampling: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgVerifySampling: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1: // verifier
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Verifier", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Verifier = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // epoch
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Epoch |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3: // validator
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validator", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Validator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4: // passed
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Passed", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
			m.Passed = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipVS(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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

// MsgVerifySamplingResponse serialization (empty message)

func (m *MsgVerifySamplingResponse) Marshal() (dAtA []byte, err error)  { return nil, nil }
func (m *MsgVerifySamplingResponse) MarshalTo(dAtA []byte) (int, error) { return 0, nil }
func (m *MsgVerifySamplingResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	return len(dAtA), nil
}
func (m *MsgVerifySamplingResponse) Size() (n int)               { return 0 }
func (m *MsgVerifySamplingResponse) Unmarshal(dAtA []byte) error { return nil }

// ---------------------------------------------------------------------------
// Encoding helpers (scoped to this file to avoid collisions with tx.pb.go)
// ---------------------------------------------------------------------------

func encodeVarintVS(dAtA []byte, offset int, v uint64) int {
	offset -= sovVS(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func sovVS(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func skipVS(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

// ---------------------------------------------------------------------------
// Extended MsgServer interface (adds VerifySampling + RemoveAgent)
// ---------------------------------------------------------------------------

type FullMsgServer interface {
	MsgServer
	VerifySampling(context.Context, *MsgVerifySampling) (*MsgVerifySamplingResponse, error)
	RemoveAgent(context.Context, *MsgRemoveAgent) (*MsgRemoveAgentResponse, error)
}

type UnimplementedFullMsgServer struct {
	UnimplementedMsgServer
}

func (*UnimplementedFullMsgServer) VerifySampling(_ context.Context, _ *MsgVerifySampling) (*MsgVerifySamplingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifySampling not implemented")
}

func (*UnimplementedFullMsgServer) RemoveAgent(_ context.Context, _ *MsgRemoveAgent) (*MsgRemoveAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveAgent not implemented")
}

// ---------------------------------------------------------------------------
// gRPC handlers
// ---------------------------------------------------------------------------

func _Msg_VerifySampling_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgVerifySampling)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FullMsgServer).VerifySampling(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/portalchain.poi.Msg/VerifySampling",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FullMsgServer).VerifySampling(ctx, req.(*MsgVerifySampling))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RemoveAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveAgent)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FullMsgServer).RemoveAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/portalchain.poi.Msg/RemoveAgent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FullMsgServer).RemoveAgent(ctx, req.(*MsgRemoveAgent))
	}
	return interceptor(ctx, in, info, handler)
}

// ---------------------------------------------------------------------------
// Combined service descriptor (all three methods)
// ---------------------------------------------------------------------------

var _FullMsg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "portalchain.poi.Msg",
	HandlerType: (*FullMsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitEpochReport",
			Handler:    _Msg_SubmitEpochReport_Handler,
		},
		{
			MethodName: "VerifySampling",
			Handler:    _Msg_VerifySampling_Handler,
		},
		{
			MethodName: "RemoveAgent",
			Handler:    _Msg_RemoveAgent_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "portalchain/poi/tx.proto",
}

func RegisterFullMsgServer(s grpc1.Server, srv FullMsgServer) {
	s.RegisterService(&_FullMsg_serviceDesc, srv)
}
