// Code generated by protoc-gen-go.
// source: vtctldata.proto
// DO NOT EDIT!

/*
Package vtctldata is a generated protocol buffer package.

It is generated from these files:
	vtctldata.proto

It has these top-level messages:
	ExecuteVtctlCommandRequest
	ExecuteVtctlCommandResponse
*/
package vtctldata

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import logutil "github.com/youtube/vitess/go/vt/proto/logutil"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// ExecuteVtctlCommandRequest is the payload for ExecuteVtctlCommand.
// timeouts are in nanoseconds.
type ExecuteVtctlCommandRequest struct {
	Args          []string `protobuf:"bytes,1,rep,name=args" json:"args,omitempty"`
	ActionTimeout int64    `protobuf:"varint,2,opt,name=action_timeout" json:"action_timeout,omitempty"`
}

func (m *ExecuteVtctlCommandRequest) Reset()                    { *m = ExecuteVtctlCommandRequest{} }
func (m *ExecuteVtctlCommandRequest) String() string            { return proto.CompactTextString(m) }
func (*ExecuteVtctlCommandRequest) ProtoMessage()               {}
func (*ExecuteVtctlCommandRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// ExecuteVtctlCommandResponse is streamed back by ExecuteVtctlCommand.
type ExecuteVtctlCommandResponse struct {
	Event *logutil.Event `protobuf:"bytes,1,opt,name=event" json:"event,omitempty"`
}

func (m *ExecuteVtctlCommandResponse) Reset()                    { *m = ExecuteVtctlCommandResponse{} }
func (m *ExecuteVtctlCommandResponse) String() string            { return proto.CompactTextString(m) }
func (*ExecuteVtctlCommandResponse) ProtoMessage()               {}
func (*ExecuteVtctlCommandResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ExecuteVtctlCommandResponse) GetEvent() *logutil.Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func init() {
	proto.RegisterType((*ExecuteVtctlCommandRequest)(nil), "vtctldata.ExecuteVtctlCommandRequest")
	proto.RegisterType((*ExecuteVtctlCommandResponse)(nil), "vtctldata.ExecuteVtctlCommandResponse")
}

var fileDescriptor0 = []byte{
	// 162 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0x2b, 0x49, 0x2e,
	0xc9, 0x49, 0x49, 0x2c, 0x49, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x84, 0x0b, 0x48,
	0xf1, 0xe6, 0xe4, 0xa7, 0x97, 0x96, 0x64, 0xe6, 0x40, 0x64, 0x94, 0x9c, 0xb8, 0xa4, 0x5c, 0x2b,
	0x52, 0x93, 0x4b, 0x4b, 0x52, 0xc3, 0x40, 0x4a, 0x9c, 0xf3, 0x73, 0x73, 0x13, 0xf3, 0x52, 0x82,
	0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b, 0x84, 0x78, 0xb8, 0x58, 0x12, 0x8b, 0xd2, 0x8b, 0x25, 0x18,
	0x15, 0x98, 0x35, 0x38, 0x85, 0xc4, 0xb8, 0xf8, 0x12, 0x93, 0x4b, 0x32, 0xf3, 0xf3, 0xe2, 0x4b,
	0x32, 0x73, 0x53, 0xf3, 0x4b, 0x4b, 0x24, 0x98, 0x14, 0x18, 0x35, 0x98, 0x95, 0x6c, 0xb8, 0xa4,
	0xb1, 0x9a, 0x51, 0x5c, 0x90, 0x9f, 0x57, 0x9c, 0x2a, 0x24, 0xcb, 0xc5, 0x9a, 0x5a, 0x96, 0x9a,
	0x57, 0x02, 0x34, 0x85, 0x51, 0x83, 0xdb, 0x88, 0x4f, 0x0f, 0xe6, 0x02, 0x57, 0x90, 0x68, 0x12,
	0x1b, 0xd8, 0x21, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb6, 0xc9, 0xde, 0x51, 0xb5, 0x00,
	0x00, 0x00,
}
