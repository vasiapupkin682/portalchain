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
// MsgRemoveAgent — proto-compatible message
// ---------------------------------------------------------------------------

const TypeMsgRemoveAgent = "remove_agent"

var _ sdk.Msg = &MsgRemoveAgent{}

type MsgRemoveAgent struct {
	Authority     string `protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	AgentAddress  string `protobuf:"bytes,2,opt,name=agent_address,json=agentAddress,proto3" json:"agent_address,omitempty"`
	Reason        string `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	AgentConsent  []byte `protobuf:"bytes,4,opt,name=agent_consent,json=agentConsent,proto3" json:"agent_consent,omitempty"`
	ConsentExpiry int64  `protobuf:"varint,5,opt,name=consent_expiry,json=consentExpiry,proto3" json:"consent_expiry,omitempty"`
}

type MsgRemoveAgentResponse struct{}

func init() {
	proto.RegisterType((*MsgRemoveAgent)(nil), "portalchain.poi.MsgRemoveAgent")
	proto.RegisterType((*MsgRemoveAgentResponse)(nil), "portalchain.poi.MsgRemoveAgentResponse")
}

// ---------------------------------------------------------------------------
// proto.Message interface
// ---------------------------------------------------------------------------

func (m *MsgRemoveAgent) Reset()         { *m = MsgRemoveAgent{} }
func (m *MsgRemoveAgent) String() string { return proto.CompactTextString(m) }
func (*MsgRemoveAgent) ProtoMessage()    {}

func (m *MsgRemoveAgentResponse) Reset()         { *m = MsgRemoveAgentResponse{} }
func (m *MsgRemoveAgentResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRemoveAgentResponse) ProtoMessage()    {}

// ---------------------------------------------------------------------------
// sdk.Msg interface
// ---------------------------------------------------------------------------

func (msg *MsgRemoveAgent) Route() string { return RouterKey }
func (msg *MsgRemoveAgent) Type() string  { return TypeMsgRemoveAgent }

func (msg *MsgRemoveAgent) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgRemoveAgent) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveAgent) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrapf(ErrInvalidValidator, "invalid authority address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.AgentAddress); err != nil {
		return sdkerrors.Wrapf(ErrInvalidValidator, "invalid agent address: %s", err)
	}
	if msg.Reason == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "reason cannot be empty")
	}
	if len(msg.AgentConsent) == 0 {
		return sdkerrors.Wrap(ErrAgentConsentInvalid, "agent consent signature cannot be empty")
	}
	if msg.ConsentExpiry <= 0 {
		return sdkerrors.Wrap(ErrAgentConsentInvalid, "consent_expiry must be a positive unix timestamp")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Proto binary serialization (gogoproto-compatible)
// ---------------------------------------------------------------------------

func (m *MsgRemoveAgent) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRemoveAgent) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRemoveAgent) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.ConsentExpiry != 0 {
		i = encodeVarintRemoveAgent(dAtA, i, uint64(m.ConsentExpiry))
		i--
		dAtA[i] = 0x28
	}
	if len(m.AgentConsent) > 0 {
		i -= len(m.AgentConsent)
		copy(dAtA[i:], m.AgentConsent)
		i = encodeVarintRemoveAgent(dAtA, i, uint64(len(m.AgentConsent)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintRemoveAgent(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.AgentAddress) > 0 {
		i -= len(m.AgentAddress)
		copy(dAtA[i:], m.AgentAddress)
		i = encodeVarintRemoveAgent(dAtA, i, uint64(len(m.AgentAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintRemoveAgent(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0x0a
	}
	return len(dAtA) - i, nil
}

func encodeVarintRemoveAgent(dAtA []byte, offset int, v uint64) int {
	offset -= sovRemoveAgent(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func (m *MsgRemoveAgent) Size() (n int) {
	if m == nil {
		return 0
	}
	l := len(m.Authority)
	if l > 0 {
		n += 1 + l + sovRemoveAgent(uint64(l))
	}
	l = len(m.AgentAddress)
	if l > 0 {
		n += 1 + l + sovRemoveAgent(uint64(l))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovRemoveAgent(uint64(l))
	}
	l = len(m.AgentConsent)
	if l > 0 {
		n += 1 + l + sovRemoveAgent(uint64(l))
	}
	if m.ConsentExpiry != 0 {
		n += 1 + sovRemoveAgent(uint64(m.ConsentExpiry))
	}
	return n
}

func sovRemoveAgent(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func (m *MsgRemoveAgent) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRemoveAgent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRemoveAgent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
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
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AgentAddress", wireType)
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
			m.AgentAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
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
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AgentConsent", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AgentConsent = append(m.AgentConsent[:0], dAtA[iNdEx:postIndex]...)
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsentExpiry", wireType)
			}
			m.ConsentExpiry = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ConsentExpiry |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipRemoveAgent(dAtA[iNdEx:])
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

func skipRemoveAgent(dAtA []byte) (n int, err error) {
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

// MsgRemoveAgentResponse serialization (empty message)

func (m *MsgRemoveAgentResponse) Marshal() (dAtA []byte, err error)  { return nil, nil }
func (m *MsgRemoveAgentResponse) MarshalTo(dAtA []byte) (int, error) { return 0, nil }
func (m *MsgRemoveAgentResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	return len(dAtA), nil
}
func (m *MsgRemoveAgentResponse) Size() (n int) { return 0 }

func (m *MsgRemoveAgentResponse) Unmarshal(dAtA []byte) error {
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
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRemoveAgentResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		iNdEx = preIndex
		skippy, err := skipRemoveAgent(dAtA[iNdEx:])
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
	return nil
}

// ---------------------------------------------------------------------------
// Extended MsgServer interface (adds RemoveAgent to the proto-generated one)
// ---------------------------------------------------------------------------

type FullMsgServer interface {
	MsgServer
	RemoveAgent(context.Context, *MsgRemoveAgent) (*MsgRemoveAgentResponse, error)
}

type UnimplementedFullMsgServer struct {
	UnimplementedMsgServer
}

func (*UnimplementedFullMsgServer) RemoveAgent(ctx context.Context, req *MsgRemoveAgent) (*MsgRemoveAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveAgent not implemented")
}

// ---------------------------------------------------------------------------
// gRPC handler and combined service descriptor
// ---------------------------------------------------------------------------

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

// _FullMsg_serviceDesc replaces _Msg_serviceDesc for service registration.
// It has the SAME ServiceName so that existing gRPC routing paths still work,
// and includes both the original SubmitEpochReport and the new RemoveAgent.
var _FullMsg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "portalchain.poi.Msg",
	HandlerType: (*FullMsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitEpochReport",
			Handler:    _Msg_SubmitEpochReport_Handler,
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
