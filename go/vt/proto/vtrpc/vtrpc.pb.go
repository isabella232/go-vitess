// Code generated by protoc-gen-go.
// source: vtrpc.proto
// DO NOT EDIT!

/*
Package vtrpc is a generated protocol buffer package.

It is generated from these files:
	vtrpc.proto

It has these top-level messages:
	CallerID
	RPCError
*/
package vtrpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Code represents canonical error codes. The names, numbers and comments
// must match the ones defined by grpc:
// https://godoc.org/google.golang.org/grpc/codes.
type Code int32

const (
	// OK is returned on success.
	Code_OK Code = 0
	// CANCELED indicates the operation was cancelled (typically by the caller).
	Code_CANCELED Code = 1
	// UNKNOWN error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	Code_UNKNOWN Code = 2
	// INVALID_ARGUMENT indicates client specified an invalid argument.
	// Note that this differs from FAILED_PRECONDITION. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	Code_INVALID_ARGUMENT Code = 3
	// DEADLINE_EXCEEDED means operation expired before completion.
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully. For
	// example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	Code_DEADLINE_EXCEEDED Code = 4
	// NOT_FOUND means some requested entity (e.g., file or directory) was
	// not found.
	Code_NOT_FOUND Code = 5
	// ALREADY_EXISTS means an attempt to create an entity failed because one
	// already exists.
	Code_ALREADY_EXISTS Code = 6
	// PERMISSION_DENIED indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use RESOURCE_EXHAUSTED
	// instead for those errors).  It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
	Code_PERMISSION_DENIED Code = 7
	// UNAUTHENTICATED indicates the request does not have valid
	// authentication credentials for the operation.
	Code_UNAUTHENTICATED Code = 16
	// RESOURCE_EXHAUSTED indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	Code_RESOURCE_EXHAUSTED Code = 8
	// FAILED_PRECONDITION indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	//
	// A litmus test that may help a service implementor in deciding
	// between FAILED_PRECONDITION, ABORTED, and UNAVAILABLE:
	//  (a) Use UNAVAILABLE if the client can retry just the failing call.
	//  (b) Use ABORTED if the client should retry at a higher-level
	//      (e.g., restarting a read-modify-write sequence).
	//  (c) Use FAILED_PRECONDITION if the client should not retry until
	//      the system state has been explicitly fixed.  E.g., if an "rmdir"
	//      fails because the directory is non-empty, FAILED_PRECONDITION
	//      should be returned since the client should not retry unless
	//      they have first fixed up the directory by deleting files from it.
	//  (d) Use FAILED_PRECONDITION if the client performs conditional
	//      REST Get/Update/Delete on a resource and the resource on the
	//      server does not match the condition. E.g., conflicting
	//      read-modify-write on the same resource.
	Code_FAILED_PRECONDITION Code = 9
	// ABORTED indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	//
	// See litmus test above for deciding between FAILED_PRECONDITION,
	// ABORTED, and UNAVAILABLE.
	Code_ABORTED Code = 10
	// OUT_OF_RANGE means operation was attempted past the valid range.
	// E.g., seeking or reading past end of file.
	//
	// Unlike INVALID_ARGUMENT, this error indicates a problem that may
	// be fixed if the system state changes. For example, a 32-bit file
	// system will generate INVALID_ARGUMENT if asked to read at an
	// offset that is not in the range [0,2^32-1], but it will generate
	// OUT_OF_RANGE if asked to read from an offset past the current
	// file size.
	//
	// There is a fair bit of overlap between FAILED_PRECONDITION and
	// OUT_OF_RANGE.  We recommend using OUT_OF_RANGE (the more specific
	// error) when it applies so that callers who are iterating through
	// a space can easily look for an OUT_OF_RANGE error to detect when
	// they are done.
	Code_OUT_OF_RANGE Code = 11
	// UNIMPLEMENTED indicates operation is not implemented or not
	// supported/enabled in this service.
	Code_UNIMPLEMENTED Code = 12
	// INTERNAL errors. Means some invariants expected by underlying
	// system has been broken.  If you see one of these errors,
	// something is very broken.
	Code_INTERNAL Code = 13
	// UNAVAILABLE indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff.
	//
	// See litmus test above for deciding between FAILED_PRECONDITION,
	// ABORTED, and UNAVAILABLE.
	Code_UNAVAILABLE Code = 14
	// DATA_LOSS indicates unrecoverable data loss or corruption.
	Code_DATA_LOSS Code = 15
)

var Code_name = map[int32]string{
	0:  "OK",
	1:  "CANCELED",
	2:  "UNKNOWN",
	3:  "INVALID_ARGUMENT",
	4:  "DEADLINE_EXCEEDED",
	5:  "NOT_FOUND",
	6:  "ALREADY_EXISTS",
	7:  "PERMISSION_DENIED",
	16: "UNAUTHENTICATED",
	8:  "RESOURCE_EXHAUSTED",
	9:  "FAILED_PRECONDITION",
	10: "ABORTED",
	11: "OUT_OF_RANGE",
	12: "UNIMPLEMENTED",
	13: "INTERNAL",
	14: "UNAVAILABLE",
	15: "DATA_LOSS",
}
var Code_value = map[string]int32{
	"OK":                  0,
	"CANCELED":            1,
	"UNKNOWN":             2,
	"INVALID_ARGUMENT":    3,
	"DEADLINE_EXCEEDED":   4,
	"NOT_FOUND":           5,
	"ALREADY_EXISTS":      6,
	"PERMISSION_DENIED":   7,
	"UNAUTHENTICATED":     16,
	"RESOURCE_EXHAUSTED":  8,
	"FAILED_PRECONDITION": 9,
	"ABORTED":             10,
	"OUT_OF_RANGE":        11,
	"UNIMPLEMENTED":       12,
	"INTERNAL":            13,
	"UNAVAILABLE":         14,
	"DATA_LOSS":           15,
}

func (x Code) String() string {
	return proto.EnumName(Code_name, int32(x))
}
func (Code) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// LegacyErrorCode is the enum values for Errors. This type is deprecated.
// Use Code instead. Background: In the initial design, we thought
// that we may end up with a different list of canonical error codes
// than the ones defined by grpc. In hindisght, we realize that
// the grpc error codes are fairly generic and mostly sufficient.
// In order to avoid confusion, this type will be deprecated in
// favor of the new Code that matches exactly what grpc defines.
// Some names below have a _LEGACY suffix. This is to prevent
// name collisions with Code.
type LegacyErrorCode int32

const (
	// SUCCESS_LEGACY is returned from a successful call.
	LegacyErrorCode_SUCCESS_LEGACY LegacyErrorCode = 0
	// CANCELLED_LEGACY means that the context was cancelled (and noticed in the app layer,
	// as opposed to the RPC layer).
	LegacyErrorCode_CANCELLED_LEGACY LegacyErrorCode = 1
	// UNKNOWN_ERROR_LEGACY includes:
	// 1. MySQL error codes that we don't explicitly handle.
	// 2. MySQL response that wasn't as expected. For example, we might expect a MySQL
	//  timestamp to be returned in a particular way, but it wasn't.
	// 3. Anything else that doesn't fall into a different bucket.
	LegacyErrorCode_UNKNOWN_ERROR_LEGACY LegacyErrorCode = 2
	// BAD_INPUT_LEGACY is returned when an end-user either sends SQL that couldn't be parsed correctly,
	// or tries a query that isn't supported by Vitess.
	LegacyErrorCode_BAD_INPUT_LEGACY LegacyErrorCode = 3
	// DEADLINE_EXCEEDED_LEGACY is returned when an action is taking longer than a given timeout.
	LegacyErrorCode_DEADLINE_EXCEEDED_LEGACY LegacyErrorCode = 4
	// INTEGRITY_ERROR_LEGACY is returned on integrity error from MySQL, usually due to
	// duplicate primary keys.
	LegacyErrorCode_INTEGRITY_ERROR_LEGACY LegacyErrorCode = 5
	// PERMISSION_DENIED_LEGACY errors are returned when a user requests access to something
	// that they don't have permissions for.
	LegacyErrorCode_PERMISSION_DENIED_LEGACY LegacyErrorCode = 6
	// RESOURCE_EXHAUSTED_LEGACY is returned when a query exceeds its quota in some dimension
	// and can't be completed due to that. Queries that return RESOURCE_EXHAUSTED
	// should not be retried, as it could be detrimental to the server's health.
	// Examples of errors that will cause the RESOURCE_EXHAUSTED code:
	// 1. TxPoolFull: this is retried server-side, and is only returned as an error
	//  if the server-side retries failed.
	// 2. Query is killed due to it taking too long.
	LegacyErrorCode_RESOURCE_EXHAUSTED_LEGACY LegacyErrorCode = 7
	// QUERY_NOT_SERVED_LEGACY means that a query could not be served right now.
	// Client can interpret it as: "the tablet that you sent this query to cannot
	// serve the query right now, try a different tablet or try again later."
	// This could be due to various reasons: QueryService is not serving, should
	// not be serving, wrong shard, wrong tablet type, blacklisted table, etc.
	// Clients that receive this error should usually retry the query, but after taking
	// the appropriate steps to make sure that the query will get sent to the correct
	// tablet.
	LegacyErrorCode_QUERY_NOT_SERVED_LEGACY LegacyErrorCode = 8
	// NOT_IN_TX_LEGACY means that we're not currently in a transaction, but we should be.
	LegacyErrorCode_NOT_IN_TX_LEGACY LegacyErrorCode = 9
	// INTERNAL_ERROR_LEGACY means some invariants expected by underlying
	// system has been broken.  If you see one of these errors,
	// something is very broken.
	LegacyErrorCode_INTERNAL_ERROR_LEGACY LegacyErrorCode = 10
	// TRANSIENT_ERROR_LEGACY is used for when there is some error that we expect we can
	// recover from automatically - often due to a resource limit temporarily being
	// reached. Retrying this error, with an exponential backoff, should succeed.
	// Clients should be able to successfully retry the query on the same backends.
	// Examples of things that can trigger this error:
	// 1. Query has been throttled
	// 2. VtGate could have request backlog
	LegacyErrorCode_TRANSIENT_ERROR_LEGACY LegacyErrorCode = 11
	// UNAUTHENTICATED_LEGACY errors are returned when a user requests access to something,
	// and we're unable to verify the user's authentication.
	LegacyErrorCode_UNAUTHENTICATED_LEGACY LegacyErrorCode = 12
)

var LegacyErrorCode_name = map[int32]string{
	0:  "SUCCESS_LEGACY",
	1:  "CANCELLED_LEGACY",
	2:  "UNKNOWN_ERROR_LEGACY",
	3:  "BAD_INPUT_LEGACY",
	4:  "DEADLINE_EXCEEDED_LEGACY",
	5:  "INTEGRITY_ERROR_LEGACY",
	6:  "PERMISSION_DENIED_LEGACY",
	7:  "RESOURCE_EXHAUSTED_LEGACY",
	8:  "QUERY_NOT_SERVED_LEGACY",
	9:  "NOT_IN_TX_LEGACY",
	10: "INTERNAL_ERROR_LEGACY",
	11: "TRANSIENT_ERROR_LEGACY",
	12: "UNAUTHENTICATED_LEGACY",
}
var LegacyErrorCode_value = map[string]int32{
	"SUCCESS_LEGACY":            0,
	"CANCELLED_LEGACY":          1,
	"UNKNOWN_ERROR_LEGACY":      2,
	"BAD_INPUT_LEGACY":          3,
	"DEADLINE_EXCEEDED_LEGACY":  4,
	"INTEGRITY_ERROR_LEGACY":    5,
	"PERMISSION_DENIED_LEGACY":  6,
	"RESOURCE_EXHAUSTED_LEGACY": 7,
	"QUERY_NOT_SERVED_LEGACY":   8,
	"NOT_IN_TX_LEGACY":          9,
	"INTERNAL_ERROR_LEGACY":     10,
	"TRANSIENT_ERROR_LEGACY":    11,
	"UNAUTHENTICATED_LEGACY":    12,
}

func (x LegacyErrorCode) String() string {
	return proto.EnumName(LegacyErrorCode_name, int32(x))
}
func (LegacyErrorCode) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// CallerID is passed along RPCs to identify the originating client
// for a request. It is not meant to be secure, but only
// informational.  The client can put whatever info they want in these
// fields, and they will be trusted by the servers. The fields will
// just be used for logging purposes, and to easily find a client.
// VtGate propagates it to VtTablet, and VtTablet may use this
// information for monitoring purposes, to display on dashboards, or
// for blacklisting purposes.
type CallerID struct {
	// principal is the effective user identifier. It is usually filled in
	// with whoever made the request to the appserver, if the request
	// came from an automated job or another system component.
	// If the request comes directly from the Internet, or if the Vitess client
	// takes action on its own accord, it is okay for this field to be absent.
	Principal string `protobuf:"bytes,1,opt,name=principal" json:"principal,omitempty"`
	// component describes the running process of the effective caller.
	// It can for instance be the hostname:port of the servlet initiating the
	// database call, or the container engine ID used by the servlet.
	Component string `protobuf:"bytes,2,opt,name=component" json:"component,omitempty"`
	// subcomponent describes a component inisde the immediate caller which
	// is responsible for generating is request. Suggested values are a
	// servlet name or an API endpoint name.
	Subcomponent string `protobuf:"bytes,3,opt,name=subcomponent" json:"subcomponent,omitempty"`
}

func (m *CallerID) Reset()                    { *m = CallerID{} }
func (m *CallerID) String() string            { return proto.CompactTextString(m) }
func (*CallerID) ProtoMessage()               {}
func (*CallerID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *CallerID) GetPrincipal() string {
	if m != nil {
		return m.Principal
	}
	return ""
}

func (m *CallerID) GetComponent() string {
	if m != nil {
		return m.Component
	}
	return ""
}

func (m *CallerID) GetSubcomponent() string {
	if m != nil {
		return m.Subcomponent
	}
	return ""
}

// RPCError is an application-level error structure returned by
// VtTablet (and passed along by VtGate if appropriate).
// We use this so the clients don't have to parse the error messages,
// but instead can depend on the value of the code.
type RPCError struct {
	LegacyCode LegacyErrorCode `protobuf:"varint,1,opt,name=legacy_code,json=legacyCode,enum=vtrpc.LegacyErrorCode" json:"legacy_code,omitempty"`
	Message    string          `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	Code       Code            `protobuf:"varint,3,opt,name=code,enum=vtrpc.Code" json:"code,omitempty"`
}

func (m *RPCError) Reset()                    { *m = RPCError{} }
func (m *RPCError) String() string            { return proto.CompactTextString(m) }
func (*RPCError) ProtoMessage()               {}
func (*RPCError) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RPCError) GetLegacyCode() LegacyErrorCode {
	if m != nil {
		return m.LegacyCode
	}
	return LegacyErrorCode_SUCCESS_LEGACY
}

func (m *RPCError) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *RPCError) GetCode() Code {
	if m != nil {
		return m.Code
	}
	return Code_OK
}

func init() {
	proto.RegisterType((*CallerID)(nil), "vtrpc.CallerID")
	proto.RegisterType((*RPCError)(nil), "vtrpc.RPCError")
	proto.RegisterEnum("vtrpc.Code", Code_name, Code_value)
	proto.RegisterEnum("vtrpc.LegacyErrorCode", LegacyErrorCode_name, LegacyErrorCode_value)
}

func init() { proto.RegisterFile("vtrpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 584 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x93, 0x4d, 0x4f, 0x1b, 0x3b,
	0x14, 0x86, 0xc9, 0x07, 0xf9, 0x38, 0x13, 0x88, 0x39, 0x7c, 0x85, 0x7b, 0xb9, 0xba, 0x57, 0xac,
	0xae, 0x58, 0xb0, 0x68, 0x17, 0x5d, 0x3b, 0xe3, 0x43, 0xb0, 0x18, 0x3c, 0xa9, 0xc7, 0xa6, 0x64,
	0x65, 0x85, 0x30, 0x42, 0xa9, 0x02, 0x13, 0x4d, 0x52, 0xa4, 0x6e, 0xfa, 0xb3, 0xfa, 0x9b, 0xfa,
	0x33, 0x2a, 0x4f, 0x32, 0x45, 0x81, 0x55, 0xe2, 0xf7, 0x39, 0x3e, 0x7e, 0xcf, 0x7b, 0x34, 0x10,
	0xbc, 0x2c, 0xf3, 0xf9, 0xe4, 0x62, 0x9e, 0x67, 0xcb, 0x0c, 0xb7, 0x8b, 0xc3, 0xd9, 0x57, 0x68,
	0x85, 0xe3, 0xd9, 0x2c, 0xcd, 0xa5, 0xc0, 0x53, 0x68, 0xcf, 0xf3, 0xe9, 0xf3, 0x64, 0x3a, 0x1f,
	0xcf, 0x7a, 0x95, 0xff, 0x2a, 0xff, 0xb7, 0xf5, 0xab, 0xe0, 0xe9, 0x24, 0x7b, 0x9a, 0x67, 0xcf,
	0xe9, 0xf3, 0xb2, 0x57, 0x5d, 0xd1, 0x3f, 0x02, 0x9e, 0x41, 0x67, 0xf1, 0xed, 0xfe, 0xb5, 0xa0,
	0x56, 0x14, 0x6c, 0x68, 0x67, 0x3f, 0xa0, 0xa5, 0x87, 0x21, 0xe5, 0x79, 0x96, 0xe3, 0x27, 0x08,
	0x66, 0xe9, 0xe3, 0x78, 0xf2, 0xdd, 0x4d, 0xb2, 0x87, 0xb4, 0x78, 0x6d, 0xf7, 0xc3, 0xd1, 0xc5,
	0xca, 0x61, 0x54, 0x90, 0xa2, 0x30, 0xcc, 0x1e, 0x52, 0x0d, 0xab, 0x52, 0xff, 0x1f, 0x7b, 0xd0,
	0x7c, 0x4a, 0x17, 0x8b, 0xf1, 0x63, 0xba, 0x36, 0x51, 0x1e, 0xf1, 0x5f, 0xa8, 0x17, 0xbd, 0x6a,
	0x45, 0xaf, 0x60, 0xdd, 0xab, 0x68, 0x50, 0x80, 0xf3, 0x9f, 0x55, 0xa8, 0x17, 0x3d, 0x1a, 0x50,
	0x8d, 0xaf, 0xd9, 0x16, 0x76, 0xa0, 0x15, 0x72, 0x15, 0x52, 0x44, 0x82, 0x55, 0x30, 0x80, 0xa6,
	0x55, 0xd7, 0x2a, 0xfe, 0xa2, 0x58, 0x15, 0x0f, 0x80, 0x49, 0x75, 0xcb, 0x23, 0x29, 0x1c, 0xd7,
	0x03, 0x7b, 0x43, 0xca, 0xb0, 0x1a, 0x1e, 0xc2, 0x9e, 0x20, 0x2e, 0x22, 0xa9, 0xc8, 0xd1, 0x5d,
	0x48, 0x24, 0x48, 0xb0, 0x3a, 0xee, 0x40, 0x5b, 0xc5, 0xc6, 0x5d, 0xc6, 0x56, 0x09, 0xb6, 0x8d,
	0x08, 0xbb, 0x3c, 0xd2, 0xc4, 0xc5, 0xc8, 0xd1, 0x9d, 0x4c, 0x4c, 0xc2, 0x1a, 0xfe, 0xe6, 0x90,
	0xf4, 0x8d, 0x4c, 0x12, 0x19, 0x2b, 0x27, 0x48, 0x49, 0x12, 0xac, 0x89, 0xfb, 0xd0, 0xb5, 0x8a,
	0x5b, 0x73, 0x45, 0xca, 0xc8, 0x90, 0x1b, 0x12, 0x8c, 0xe1, 0x11, 0xa0, 0xa6, 0x24, 0xb6, 0x3a,
	0xf4, 0xaf, 0x5c, 0x71, 0x9b, 0x78, 0xbd, 0x85, 0xc7, 0xb0, 0x7f, 0xc9, 0x65, 0x44, 0xc2, 0x0d,
	0x35, 0x85, 0xb1, 0x12, 0xd2, 0xc8, 0x58, 0xb1, 0xb6, 0x77, 0xce, 0xfb, 0xb1, 0xf6, 0x55, 0x80,
	0x0c, 0x3a, 0xb1, 0x35, 0x2e, 0xbe, 0x74, 0x9a, 0xab, 0x01, 0xb1, 0x00, 0xf7, 0x60, 0xc7, 0x2a,
	0x79, 0x33, 0x8c, 0xc8, 0x8f, 0x41, 0x82, 0x75, 0xfc, 0xe4, 0x52, 0x19, 0xd2, 0x8a, 0x47, 0x6c,
	0x07, 0xbb, 0x10, 0x58, 0xc5, 0x6f, 0xb9, 0x8c, 0x78, 0x3f, 0x22, 0xb6, 0xeb, 0x07, 0x12, 0xdc,
	0x70, 0x17, 0xc5, 0x49, 0xc2, 0xba, 0xe7, 0xbf, 0xaa, 0xd0, 0x7d, 0xb3, 0x13, 0x3f, 0x64, 0x62,
	0xc3, 0x90, 0x92, 0xc4, 0x45, 0x34, 0xe0, 0xe1, 0x88, 0x6d, 0xf9, 0xd0, 0x56, 0x79, 0x7a, 0x8f,
	0x6b, 0xb5, 0x82, 0x3d, 0x38, 0x58, 0xe7, 0xea, 0x48, 0xeb, 0x58, 0x97, 0xa4, 0x08, 0xb9, 0xcf,
	0x85, 0x93, 0x6a, 0x68, 0x4d, 0xa9, 0xd6, 0xf0, 0x14, 0x7a, 0xef, 0x42, 0x2e, 0x69, 0x1d, 0xff,
	0x82, 0x23, 0xef, 0x7c, 0xa0, 0xa5, 0x19, 0x6d, 0xf6, 0xdb, 0xf6, 0x37, 0xdf, 0x85, 0x5c, 0xd2,
	0x06, 0xfe, 0x03, 0x27, 0xef, 0x63, 0x2d, 0x71, 0x13, 0xff, 0x86, 0xe3, 0xcf, 0x96, 0xf4, 0xc8,
	0xf9, 0x55, 0x26, 0xa4, 0x6f, 0x5f, 0x61, 0xcb, 0x3b, 0xf5, 0xb2, 0x54, 0xce, 0xdc, 0x95, 0x6a,
	0x1b, 0x4f, 0xe0, 0xb0, 0x4c, 0x71, 0xd3, 0x0a, 0x78, 0x9b, 0x46, 0x73, 0x95, 0x48, 0x52, 0x66,
	0x93, 0x05, 0x9e, 0xbd, 0x59, 0x7a, 0xc9, 0x3a, 0xfd, 0x3d, 0xe8, 0x4e, 0xb3, 0x8b, 0x97, 0xe9,
	0x32, 0x5d, 0x2c, 0x56, 0x5f, 0xea, 0x7d, 0xa3, 0xf8, 0xf9, 0xf8, 0x3b, 0x00, 0x00, 0xff, 0xff,
	0xdd, 0xc2, 0x4b, 0x19, 0xbf, 0x03, 0x00, 0x00,
}
