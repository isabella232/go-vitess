// Code generated by protoc-gen-go.
// source: binlogdata.proto
// DO NOT EDIT!

/*
Package binlogdata is a generated protocol buffer package.

It is generated from these files:
	binlogdata.proto

It has these top-level messages:
	Charset
	BinlogTransaction
	StreamEvent
	StreamUpdateRequest
	StreamUpdateResponse
	StreamKeyRangeRequest
	StreamKeyRangeResponse
	StreamTablesRequest
	StreamTablesResponse
*/
package binlogdata

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import query "github.com/youtube/vitess/go/vt/proto/query"
import topodata "github.com/youtube/vitess/go/vt/proto/topodata"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type BinlogTransaction_Statement_Category int32

const (
	BinlogTransaction_Statement_BL_UNRECOGNIZED BinlogTransaction_Statement_Category = 0
	BinlogTransaction_Statement_BL_BEGIN        BinlogTransaction_Statement_Category = 1
	BinlogTransaction_Statement_BL_COMMIT       BinlogTransaction_Statement_Category = 2
	BinlogTransaction_Statement_BL_ROLLBACK     BinlogTransaction_Statement_Category = 3
	BinlogTransaction_Statement_BL_DML          BinlogTransaction_Statement_Category = 4
	BinlogTransaction_Statement_BL_DDL          BinlogTransaction_Statement_Category = 5
	BinlogTransaction_Statement_BL_SET          BinlogTransaction_Statement_Category = 6
)

var BinlogTransaction_Statement_Category_name = map[int32]string{
	0: "BL_UNRECOGNIZED",
	1: "BL_BEGIN",
	2: "BL_COMMIT",
	3: "BL_ROLLBACK",
	4: "BL_DML",
	5: "BL_DDL",
	6: "BL_SET",
}
var BinlogTransaction_Statement_Category_value = map[string]int32{
	"BL_UNRECOGNIZED": 0,
	"BL_BEGIN":        1,
	"BL_COMMIT":       2,
	"BL_ROLLBACK":     3,
	"BL_DML":          4,
	"BL_DDL":          5,
	"BL_SET":          6,
}

func (x BinlogTransaction_Statement_Category) String() string {
	return proto.EnumName(BinlogTransaction_Statement_Category_name, int32(x))
}
func (BinlogTransaction_Statement_Category) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{1, 0, 0}
}

// the category of this event
type StreamEvent_Category int32

const (
	StreamEvent_SE_ERR StreamEvent_Category = 0
	StreamEvent_SE_DML StreamEvent_Category = 1
	StreamEvent_SE_DDL StreamEvent_Category = 2
	StreamEvent_SE_POS StreamEvent_Category = 3
)

var StreamEvent_Category_name = map[int32]string{
	0: "SE_ERR",
	1: "SE_DML",
	2: "SE_DDL",
	3: "SE_POS",
}
var StreamEvent_Category_value = map[string]int32{
	"SE_ERR": 0,
	"SE_DML": 1,
	"SE_DDL": 2,
	"SE_POS": 3,
}

func (x StreamEvent_Category) String() string {
	return proto.EnumName(StreamEvent_Category_name, int32(x))
}
func (StreamEvent_Category) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

// Charset is the per-statement charset info from a QUERY_EVENT binlog entry.
type Charset struct {
	// @@session.character_set_client
	Client int32 `protobuf:"varint,1,opt,name=client" json:"client,omitempty"`
	// @@session.collation_connection
	Conn int32 `protobuf:"varint,2,opt,name=conn" json:"conn,omitempty"`
	// @@session.collation_server
	Server int32 `protobuf:"varint,3,opt,name=server" json:"server,omitempty"`
}

func (m *Charset) Reset()                    { *m = Charset{} }
func (m *Charset) String() string            { return proto.CompactTextString(m) }
func (*Charset) ProtoMessage()               {}
func (*Charset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// BinlogTransaction describes a transaction inside the binlogs.
type BinlogTransaction struct {
	// the statements in this transaction
	Statements []*BinlogTransaction_Statement `protobuf:"bytes,1,rep,name=statements" json:"statements,omitempty"`
	// the timestamp of the statements
	Timestamp int64 `protobuf:"varint,2,opt,name=timestamp" json:"timestamp,omitempty"`
	// the Transaction ID after this statement was applied
	TransactionId string `protobuf:"bytes,3,opt,name=transaction_id,json=transactionId" json:"transaction_id,omitempty"`
}

func (m *BinlogTransaction) Reset()                    { *m = BinlogTransaction{} }
func (m *BinlogTransaction) String() string            { return proto.CompactTextString(m) }
func (*BinlogTransaction) ProtoMessage()               {}
func (*BinlogTransaction) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *BinlogTransaction) GetStatements() []*BinlogTransaction_Statement {
	if m != nil {
		return m.Statements
	}
	return nil
}

type BinlogTransaction_Statement struct {
	// what type of statement is this?
	Category BinlogTransaction_Statement_Category `protobuf:"varint,1,opt,name=category,enum=binlogdata.BinlogTransaction_Statement_Category" json:"category,omitempty"`
	// charset of this statement, if different from pre-negotiated default.
	Charset *Charset `protobuf:"bytes,2,opt,name=charset" json:"charset,omitempty"`
	// the sql
	Sql string `protobuf:"bytes,3,opt,name=sql" json:"sql,omitempty"`
}

func (m *BinlogTransaction_Statement) Reset()                    { *m = BinlogTransaction_Statement{} }
func (m *BinlogTransaction_Statement) String() string            { return proto.CompactTextString(m) }
func (*BinlogTransaction_Statement) ProtoMessage()               {}
func (*BinlogTransaction_Statement) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

func (m *BinlogTransaction_Statement) GetCharset() *Charset {
	if m != nil {
		return m.Charset
	}
	return nil
}

// StreamEvent describes an update stream event inside the binlogs.
type StreamEvent struct {
	Category StreamEvent_Category `protobuf:"varint,1,opt,name=category,enum=binlogdata.StreamEvent_Category" json:"category,omitempty"`
	// table_name, primary_key_fields and primary_key_values are set for SE_DML
	TableName        string         `protobuf:"bytes,2,opt,name=table_name,json=tableName" json:"table_name,omitempty"`
	PrimaryKeyFields []*query.Field `protobuf:"bytes,3,rep,name=primary_key_fields,json=primaryKeyFields" json:"primary_key_fields,omitempty"`
	PrimaryKeyValues []*query.Row   `protobuf:"bytes,4,rep,name=primary_key_values,json=primaryKeyValues" json:"primary_key_values,omitempty"`
	// sql is set for SE_DDL or SE_ERR
	Sql string `protobuf:"bytes,5,opt,name=sql" json:"sql,omitempty"`
	// timestamp is set for SE_DML, SE_DDL or SE_ERR
	Timestamp int64 `protobuf:"varint,6,opt,name=timestamp" json:"timestamp,omitempty"`
	// the Transaction ID after this statement was applied
	TransactionId string `protobuf:"bytes,7,opt,name=transaction_id,json=transactionId" json:"transaction_id,omitempty"`
}

func (m *StreamEvent) Reset()                    { *m = StreamEvent{} }
func (m *StreamEvent) String() string            { return proto.CompactTextString(m) }
func (*StreamEvent) ProtoMessage()               {}
func (*StreamEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *StreamEvent) GetPrimaryKeyFields() []*query.Field {
	if m != nil {
		return m.PrimaryKeyFields
	}
	return nil
}

func (m *StreamEvent) GetPrimaryKeyValues() []*query.Row {
	if m != nil {
		return m.PrimaryKeyValues
	}
	return nil
}

// StreamUpdateRequest is the payload to StreamUpdate
type StreamUpdateRequest struct {
	// where to start
	Position string `protobuf:"bytes,1,opt,name=position" json:"position,omitempty"`
}

func (m *StreamUpdateRequest) Reset()                    { *m = StreamUpdateRequest{} }
func (m *StreamUpdateRequest) String() string            { return proto.CompactTextString(m) }
func (*StreamUpdateRequest) ProtoMessage()               {}
func (*StreamUpdateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// StreamUpdateResponse is the response from StreamUpdate
type StreamUpdateResponse struct {
	StreamEvent *StreamEvent `protobuf:"bytes,1,opt,name=stream_event,json=streamEvent" json:"stream_event,omitempty"`
}

func (m *StreamUpdateResponse) Reset()                    { *m = StreamUpdateResponse{} }
func (m *StreamUpdateResponse) String() string            { return proto.CompactTextString(m) }
func (*StreamUpdateResponse) ProtoMessage()               {}
func (*StreamUpdateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *StreamUpdateResponse) GetStreamEvent() *StreamEvent {
	if m != nil {
		return m.StreamEvent
	}
	return nil
}

// StreamKeyRangeRequest is the payload to StreamKeyRange
type StreamKeyRangeRequest struct {
	// where to start
	Position string `protobuf:"bytes,1,opt,name=position" json:"position,omitempty"`
	// what to get
	KeyRange *topodata.KeyRange `protobuf:"bytes,2,opt,name=key_range,json=keyRange" json:"key_range,omitempty"`
	// default charset on the player side
	Charset *Charset `protobuf:"bytes,3,opt,name=charset" json:"charset,omitempty"`
}

func (m *StreamKeyRangeRequest) Reset()                    { *m = StreamKeyRangeRequest{} }
func (m *StreamKeyRangeRequest) String() string            { return proto.CompactTextString(m) }
func (*StreamKeyRangeRequest) ProtoMessage()               {}
func (*StreamKeyRangeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *StreamKeyRangeRequest) GetKeyRange() *topodata.KeyRange {
	if m != nil {
		return m.KeyRange
	}
	return nil
}

func (m *StreamKeyRangeRequest) GetCharset() *Charset {
	if m != nil {
		return m.Charset
	}
	return nil
}

// StreamKeyRangeResponse is the response from StreamKeyRange
type StreamKeyRangeResponse struct {
	BinlogTransaction *BinlogTransaction `protobuf:"bytes,1,opt,name=binlog_transaction,json=binlogTransaction" json:"binlog_transaction,omitempty"`
}

func (m *StreamKeyRangeResponse) Reset()                    { *m = StreamKeyRangeResponse{} }
func (m *StreamKeyRangeResponse) String() string            { return proto.CompactTextString(m) }
func (*StreamKeyRangeResponse) ProtoMessage()               {}
func (*StreamKeyRangeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *StreamKeyRangeResponse) GetBinlogTransaction() *BinlogTransaction {
	if m != nil {
		return m.BinlogTransaction
	}
	return nil
}

// StreamTablesRequest is the payload to StreamTables
type StreamTablesRequest struct {
	// where to start
	Position string `protobuf:"bytes,1,opt,name=position" json:"position,omitempty"`
	// what to get
	Tables []string `protobuf:"bytes,2,rep,name=tables" json:"tables,omitempty"`
	// default charset on the player side
	Charset *Charset `protobuf:"bytes,3,opt,name=charset" json:"charset,omitempty"`
}

func (m *StreamTablesRequest) Reset()                    { *m = StreamTablesRequest{} }
func (m *StreamTablesRequest) String() string            { return proto.CompactTextString(m) }
func (*StreamTablesRequest) ProtoMessage()               {}
func (*StreamTablesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *StreamTablesRequest) GetCharset() *Charset {
	if m != nil {
		return m.Charset
	}
	return nil
}

// StreamTablesResponse is the response from StreamTables
type StreamTablesResponse struct {
	BinlogTransaction *BinlogTransaction `protobuf:"bytes,1,opt,name=binlog_transaction,json=binlogTransaction" json:"binlog_transaction,omitempty"`
}

func (m *StreamTablesResponse) Reset()                    { *m = StreamTablesResponse{} }
func (m *StreamTablesResponse) String() string            { return proto.CompactTextString(m) }
func (*StreamTablesResponse) ProtoMessage()               {}
func (*StreamTablesResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *StreamTablesResponse) GetBinlogTransaction() *BinlogTransaction {
	if m != nil {
		return m.BinlogTransaction
	}
	return nil
}

func init() {
	proto.RegisterType((*Charset)(nil), "binlogdata.Charset")
	proto.RegisterType((*BinlogTransaction)(nil), "binlogdata.BinlogTransaction")
	proto.RegisterType((*BinlogTransaction_Statement)(nil), "binlogdata.BinlogTransaction.Statement")
	proto.RegisterType((*StreamEvent)(nil), "binlogdata.StreamEvent")
	proto.RegisterType((*StreamUpdateRequest)(nil), "binlogdata.StreamUpdateRequest")
	proto.RegisterType((*StreamUpdateResponse)(nil), "binlogdata.StreamUpdateResponse")
	proto.RegisterType((*StreamKeyRangeRequest)(nil), "binlogdata.StreamKeyRangeRequest")
	proto.RegisterType((*StreamKeyRangeResponse)(nil), "binlogdata.StreamKeyRangeResponse")
	proto.RegisterType((*StreamTablesRequest)(nil), "binlogdata.StreamTablesRequest")
	proto.RegisterType((*StreamTablesResponse)(nil), "binlogdata.StreamTablesResponse")
	proto.RegisterEnum("binlogdata.BinlogTransaction_Statement_Category", BinlogTransaction_Statement_Category_name, BinlogTransaction_Statement_Category_value)
	proto.RegisterEnum("binlogdata.StreamEvent_Category", StreamEvent_Category_name, StreamEvent_Category_value)
}

var fileDescriptor0 = []byte{
	// 664 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xb4, 0x54, 0xcf, 0x6e, 0xd3, 0x4e,
	0x10, 0xfe, 0x25, 0x6e, 0xd3, 0x78, 0xdc, 0x3f, 0xee, 0xf6, 0x47, 0x89, 0x22, 0x2a, 0x55, 0x96,
	0x10, 0xbd, 0x10, 0x20, 0x5c, 0x50, 0xc5, 0x85, 0xa4, 0xa6, 0x8a, 0xea, 0x24, 0x68, 0x93, 0x72,
	0xe0, 0x62, 0x6d, 0x92, 0x6d, 0xb1, 0x9a, 0xd8, 0xae, 0x77, 0x1b, 0xc8, 0x43, 0x70, 0xe2, 0x49,
	0x78, 0x36, 0x5e, 0x80, 0xdd, 0xf5, 0xda, 0x71, 0xff, 0x40, 0xcb, 0x81, 0xdb, 0xcc, 0xb7, 0x33,
	0xb3, 0xdf, 0x7c, 0x33, 0xbb, 0x60, 0x8f, 0x82, 0x70, 0x1a, 0x9d, 0x4f, 0x08, 0x27, 0x8d, 0x38,
	0x89, 0x78, 0x84, 0x60, 0x89, 0xd4, 0xad, 0xcb, 0x2b, 0x9a, 0x2c, 0xd2, 0x83, 0xfa, 0x26, 0x8f,
	0xe2, 0x68, 0x19, 0xe8, 0x74, 0x61, 0xad, 0xfd, 0x99, 0x24, 0x8c, 0x72, 0xb4, 0x0b, 0x95, 0xf1,
	0x34, 0xa0, 0x21, 0xaf, 0x95, 0xf6, 0x4b, 0x07, 0xab, 0x58, 0x7b, 0x08, 0xc1, 0xca, 0x38, 0x0a,
	0xc3, 0x5a, 0x59, 0xa1, 0xca, 0x96, 0xb1, 0x8c, 0x26, 0x73, 0x9a, 0xd4, 0x8c, 0x34, 0x36, 0xf5,
	0x9c, 0x1f, 0x06, 0x6c, 0xb7, 0xd4, 0xd5, 0xc3, 0x84, 0x84, 0x8c, 0x8c, 0x79, 0x10, 0x85, 0xe8,
	0x18, 0x80, 0x71, 0xc2, 0xe9, 0x4c, 0x94, 0x63, 0xa2, 0xba, 0x71, 0x60, 0x35, 0x9f, 0x35, 0x0a,
	0xa4, 0x6f, 0xa5, 0x34, 0x06, 0x59, 0x3c, 0x2e, 0xa4, 0xa2, 0x27, 0x60, 0xf2, 0x60, 0x46, 0x05,
	0x32, 0x8b, 0x15, 0x1f, 0x03, 0x2f, 0x01, 0xf4, 0x14, 0x36, 0xf9, 0xb2, 0x84, 0x1f, 0x4c, 0x14,
	0x39, 0x13, 0x6f, 0x14, 0xd0, 0xce, 0xa4, 0xfe, 0xad, 0x0c, 0x66, 0x5e, 0x1e, 0x79, 0x50, 0x1d,
	0x0b, 0xfb, 0x3c, 0x4a, 0x16, 0xaa, 0xef, 0xcd, 0xe6, 0xcb, 0x07, 0x32, 0x6b, 0xb4, 0x75, 0x1e,
	0xce, 0x2b, 0xa0, 0xe7, 0xb0, 0x36, 0x4e, 0xe5, 0x54, 0xf4, 0xac, 0xe6, 0x4e, 0xb1, 0x98, 0x56,
	0x1a, 0x67, 0x31, 0xc8, 0x06, 0x83, 0x5d, 0x4e, 0x35, 0x4d, 0x69, 0x3a, 0x97, 0x50, 0xcd, 0xca,
	0xa2, 0x1d, 0xd8, 0x6a, 0x79, 0xfe, 0x69, 0x0f, 0xbb, 0xed, 0xfe, 0x71, 0xaf, 0xf3, 0xc9, 0x3d,
	0xb2, 0xff, 0x43, 0xeb, 0x50, 0x15, 0x60, 0xcb, 0x3d, 0xee, 0xf4, 0xec, 0x12, 0xda, 0x00, 0x53,
	0x78, 0xed, 0x7e, 0xb7, 0xdb, 0x19, 0xda, 0x65, 0xb4, 0x05, 0x96, 0x70, 0x71, 0xdf, 0xf3, 0x5a,
	0xef, 0xda, 0x27, 0xb6, 0x81, 0x00, 0x2a, 0x02, 0x38, 0xea, 0x7a, 0xf6, 0x4a, 0x66, 0x1f, 0x79,
	0xf6, 0xaa, 0xb6, 0x07, 0xee, 0xd0, 0xae, 0x38, 0x3f, 0xcb, 0x60, 0x0d, 0x78, 0x42, 0xc9, 0xcc,
	0x9d, 0x4b, 0x45, 0xde, 0xde, 0x52, 0x64, 0xbf, 0xd8, 0x44, 0x21, 0xf4, 0x2e, 0x05, 0xf6, 0x00,
	0x38, 0x19, 0x4d, 0xa9, 0x1f, 0x92, 0x19, 0x55, 0x22, 0x98, 0x62, 0x46, 0x12, 0xe9, 0x09, 0x00,
	0x1d, 0x02, 0x8a, 0x93, 0x60, 0x46, 0x92, 0x85, 0x7f, 0x41, 0x17, 0xfe, 0x59, 0x40, 0xa7, 0x13,
	0x26, 0x04, 0x90, 0x2b, 0xb1, 0xde, 0x48, 0x37, 0xf5, 0xbd, 0x04, 0xb1, 0xad, 0xe3, 0x4e, 0xe8,
	0x42, 0x01, 0x0c, 0xbd, 0xb9, 0x9e, 0x3b, 0x27, 0xd3, 0x2b, 0xca, 0x6a, 0x2b, 0x2a, 0x17, 0x74,
	0x2e, 0x8e, 0xbe, 0x14, 0x33, 0x3f, 0xaa, 0x98, 0x4c, 0xe7, 0xd5, 0x5c, 0xe7, 0xeb, 0x9b, 0x54,
	0xb9, 0x7f, 0x93, 0xd6, 0xee, 0xd8, 0x24, 0xe7, 0xb0, 0x30, 0x2c, 0xa1, 0xe8, 0xc0, 0xf5, 0x5d,
	0x8c, 0xc5, 0x8c, 0x52, 0x5b, 0xaa, 0x5e, 0xca, 0x6c, 0xa1, 0x7a, 0x59, 0xdb, 0x1f, 0xfa, 0x03,
	0xdb, 0x70, 0x5e, 0xc1, 0x4e, 0xaa, 0xe4, 0x69, 0x2c, 0x64, 0xa5, 0x98, 0x0a, 0xfe, 0x8c, 0xa3,
	0x3a, 0x54, 0xe3, 0x88, 0x05, 0xf2, 0x02, 0x25, 0xbe, 0x89, 0x73, 0xdf, 0xc1, 0xf0, 0xff, 0xf5,
	0x14, 0x16, 0x47, 0x21, 0x93, 0x9a, 0xae, 0x33, 0x85, 0xfb, 0x74, 0x9e, 0x3d, 0x5f, 0xab, 0xf9,
	0xf8, 0x37, 0x43, 0xc3, 0x16, 0x5b, 0x3a, 0xce, 0xf7, 0x12, 0x3c, 0x4a, 0x0f, 0x85, 0x5a, 0x98,
	0x84, 0xe7, 0x0f, 0x61, 0x82, 0x5e, 0x80, 0x29, 0x27, 0x90, 0xc8, 0x78, 0xbd, 0xe8, 0xa8, 0x91,
	0xff, 0x2c, 0x79, 0xa5, 0xea, 0x85, 0xb6, 0x8a, 0xef, 0xc2, 0xb8, 0xff, 0x5d, 0x38, 0x67, 0xb0,
	0x7b, 0x93, 0x94, 0xee, 0xd5, 0x03, 0x94, 0x26, 0xfa, 0x85, 0x51, 0xe8, 0x8e, 0xf7, 0xfe, 0xf8,
	0x70, 0xf1, 0xf6, 0xe8, 0x26, 0xe4, 0x7c, 0xcd, 0x86, 0x30, 0x94, 0x0b, 0xca, 0x1e, 0xd2, 0xba,
	0xf8, 0xf9, 0xd4, 0x36, 0x33, 0xd1, 0xb7, 0x21, 0x4e, 0xb4, 0xf7, 0xb7, 0x1d, 0x4e, 0xb2, 0x59,
	0x66, 0x37, 0xff, 0x8b, 0xfe, 0x46, 0x15, 0xf5, 0xc9, 0xbf, 0xfe, 0x15, 0x00, 0x00, 0xff, 0xff,
	0x46, 0xa2, 0xfc, 0xd1, 0x21, 0x06, 0x00, 0x00,
}
