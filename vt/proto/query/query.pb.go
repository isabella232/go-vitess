// Code generated by protoc-gen-go.
// source: query.proto
// DO NOT EDIT!

/*
Package query is a generated protocol buffer package.

It is generated from these files:
	query.proto

It has these top-level messages:
	Target
	VTGateCallerID
	BindVariable
	BoundQuery
	Field
	Row
	QueryResult
	GetSessionIdRequest
	GetSessionIdResponse
	ExecuteRequest
	ExecuteResponse
	ExecuteBatchRequest
	ExecuteBatchResponse
	StreamExecuteRequest
	StreamExecuteResponse
	BeginRequest
	BeginResponse
	CommitRequest
	CommitResponse
	RollbackRequest
	RollbackResponse
	SplitQueryRequest
	QuerySplit
	SplitQueryResponse
	StreamHealthRequest
	RealtimeStats
	StreamHealthResponse
*/
package query

import proto "github.com/golang/protobuf/proto"
import topodata "github.com/youtube/vitess/go/vt/proto/topodata"
import vtrpc "github.com/youtube/vitess/go/vt/proto/vtrpc"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type BindVariable_Type int32

const (
	BindVariable_TYPE_NULL       BindVariable_Type = 0
	BindVariable_TYPE_BYTES      BindVariable_Type = 1
	BindVariable_TYPE_INT        BindVariable_Type = 2
	BindVariable_TYPE_UINT       BindVariable_Type = 3
	BindVariable_TYPE_FLOAT      BindVariable_Type = 4
	BindVariable_TYPE_BYTES_LIST BindVariable_Type = 5
	BindVariable_TYPE_INT_LIST   BindVariable_Type = 6
	BindVariable_TYPE_UINT_LIST  BindVariable_Type = 7
	BindVariable_TYPE_FLOAT_LIST BindVariable_Type = 8
)

var BindVariable_Type_name = map[int32]string{
	0: "TYPE_NULL",
	1: "TYPE_BYTES",
	2: "TYPE_INT",
	3: "TYPE_UINT",
	4: "TYPE_FLOAT",
	5: "TYPE_BYTES_LIST",
	6: "TYPE_INT_LIST",
	7: "TYPE_UINT_LIST",
	8: "TYPE_FLOAT_LIST",
}
var BindVariable_Type_value = map[string]int32{
	"TYPE_NULL":       0,
	"TYPE_BYTES":      1,
	"TYPE_INT":        2,
	"TYPE_UINT":       3,
	"TYPE_FLOAT":      4,
	"TYPE_BYTES_LIST": 5,
	"TYPE_INT_LIST":   6,
	"TYPE_UINT_LIST":  7,
	"TYPE_FLOAT_LIST": 8,
}

func (x BindVariable_Type) String() string {
	return proto.EnumName(BindVariable_Type_name, int32(x))
}

// Type follows enum_field_types from mysql.h.
type Field_Type int32

const (
	Field_TYPE_DECIMAL     Field_Type = 0
	Field_TYPE_TINY        Field_Type = 1
	Field_TYPE_SHORT       Field_Type = 2
	Field_TYPE_LONG        Field_Type = 3
	Field_TYPE_FLOAT       Field_Type = 4
	Field_TYPE_DOUBLE      Field_Type = 5
	Field_TYPE_NULL        Field_Type = 6
	Field_TYPE_TIMESTAMP   Field_Type = 7
	Field_TYPE_LONGLONG    Field_Type = 8
	Field_TYPE_INT24       Field_Type = 9
	Field_TYPE_DATE        Field_Type = 10
	Field_TYPE_TIME        Field_Type = 11
	Field_TYPE_DATETIME    Field_Type = 12
	Field_TYPE_YEAR        Field_Type = 13
	Field_TYPE_NEWDATE     Field_Type = 14
	Field_TYPE_VARCHAR     Field_Type = 15
	Field_TYPE_BIT         Field_Type = 16
	Field_TYPE_NEWDECIMAL  Field_Type = 246
	Field_TYPE_ENUM        Field_Type = 247
	Field_TYPE_SET         Field_Type = 248
	Field_TYPE_TINY_BLOB   Field_Type = 249
	Field_TYPE_MEDIUM_BLOB Field_Type = 250
	Field_TYPE_LONG_BLOB   Field_Type = 251
	Field_TYPE_BLOB        Field_Type = 252
	Field_TYPE_VAR_STRING  Field_Type = 253
	Field_TYPE_STRING      Field_Type = 254
	Field_TYPE_GEOMETRY    Field_Type = 255
)

var Field_Type_name = map[int32]string{
	0:   "TYPE_DECIMAL",
	1:   "TYPE_TINY",
	2:   "TYPE_SHORT",
	3:   "TYPE_LONG",
	4:   "TYPE_FLOAT",
	5:   "TYPE_DOUBLE",
	6:   "TYPE_NULL",
	7:   "TYPE_TIMESTAMP",
	8:   "TYPE_LONGLONG",
	9:   "TYPE_INT24",
	10:  "TYPE_DATE",
	11:  "TYPE_TIME",
	12:  "TYPE_DATETIME",
	13:  "TYPE_YEAR",
	14:  "TYPE_NEWDATE",
	15:  "TYPE_VARCHAR",
	16:  "TYPE_BIT",
	246: "TYPE_NEWDECIMAL",
	247: "TYPE_ENUM",
	248: "TYPE_SET",
	249: "TYPE_TINY_BLOB",
	250: "TYPE_MEDIUM_BLOB",
	251: "TYPE_LONG_BLOB",
	252: "TYPE_BLOB",
	253: "TYPE_VAR_STRING",
	254: "TYPE_STRING",
	255: "TYPE_GEOMETRY",
}
var Field_Type_value = map[string]int32{
	"TYPE_DECIMAL":     0,
	"TYPE_TINY":        1,
	"TYPE_SHORT":       2,
	"TYPE_LONG":        3,
	"TYPE_FLOAT":       4,
	"TYPE_DOUBLE":      5,
	"TYPE_NULL":        6,
	"TYPE_TIMESTAMP":   7,
	"TYPE_LONGLONG":    8,
	"TYPE_INT24":       9,
	"TYPE_DATE":        10,
	"TYPE_TIME":        11,
	"TYPE_DATETIME":    12,
	"TYPE_YEAR":        13,
	"TYPE_NEWDATE":     14,
	"TYPE_VARCHAR":     15,
	"TYPE_BIT":         16,
	"TYPE_NEWDECIMAL":  246,
	"TYPE_ENUM":        247,
	"TYPE_SET":         248,
	"TYPE_TINY_BLOB":   249,
	"TYPE_MEDIUM_BLOB": 250,
	"TYPE_LONG_BLOB":   251,
	"TYPE_BLOB":        252,
	"TYPE_VAR_STRING":  253,
	"TYPE_STRING":      254,
	"TYPE_GEOMETRY":    255,
}

func (x Field_Type) String() string {
	return proto.EnumName(Field_Type_name, int32(x))
}

// Flag contains the MySQL field flags bitset values e.g. to
// distinguish between signed and unsigned integer.  These numbers
// should exactly match values defined in
// dist/mysql-5.1.52/include/mysql_com.h
type Field_Flag int32

const (
	// ZEROVALUE_FLAG is not part of the MySQL specification and only
	// used in unit tests.
	Field_VT_ZEROVALUE_FLAG        Field_Flag = 0
	Field_VT_NOT_NULL_FLAG         Field_Flag = 1
	Field_VT_PRI_KEY_FLAG          Field_Flag = 2
	Field_VT_UNIQUE_KEY_FLAG       Field_Flag = 4
	Field_VT_MULTIPLE_KEY_FLAG     Field_Flag = 8
	Field_VT_BLOB_FLAG             Field_Flag = 16
	Field_VT_UNSIGNED_FLAG         Field_Flag = 32
	Field_VT_ZEROFILL_FLAG         Field_Flag = 64
	Field_VT_BINARY_FLAG           Field_Flag = 128
	Field_VT_ENUM_FLAG             Field_Flag = 256
	Field_VT_AUTO_INCREMENT_FLAG   Field_Flag = 512
	Field_VT_TIMESTAMP_FLAG        Field_Flag = 1024
	Field_VT_SET_FLAG              Field_Flag = 2048
	Field_VT_NO_DEFAULT_VALUE_FLAG Field_Flag = 4096
	Field_VT_ON_UPDATE_NOW_FLAG    Field_Flag = 8192
	Field_VT_NUM_FLAG              Field_Flag = 32768
)

var Field_Flag_name = map[int32]string{
	0:     "VT_ZEROVALUE_FLAG",
	1:     "VT_NOT_NULL_FLAG",
	2:     "VT_PRI_KEY_FLAG",
	4:     "VT_UNIQUE_KEY_FLAG",
	8:     "VT_MULTIPLE_KEY_FLAG",
	16:    "VT_BLOB_FLAG",
	32:    "VT_UNSIGNED_FLAG",
	64:    "VT_ZEROFILL_FLAG",
	128:   "VT_BINARY_FLAG",
	256:   "VT_ENUM_FLAG",
	512:   "VT_AUTO_INCREMENT_FLAG",
	1024:  "VT_TIMESTAMP_FLAG",
	2048:  "VT_SET_FLAG",
	4096:  "VT_NO_DEFAULT_VALUE_FLAG",
	8192:  "VT_ON_UPDATE_NOW_FLAG",
	32768: "VT_NUM_FLAG",
}
var Field_Flag_value = map[string]int32{
	"VT_ZEROVALUE_FLAG":        0,
	"VT_NOT_NULL_FLAG":         1,
	"VT_PRI_KEY_FLAG":          2,
	"VT_UNIQUE_KEY_FLAG":       4,
	"VT_MULTIPLE_KEY_FLAG":     8,
	"VT_BLOB_FLAG":             16,
	"VT_UNSIGNED_FLAG":         32,
	"VT_ZEROFILL_FLAG":         64,
	"VT_BINARY_FLAG":           128,
	"VT_ENUM_FLAG":             256,
	"VT_AUTO_INCREMENT_FLAG":   512,
	"VT_TIMESTAMP_FLAG":        1024,
	"VT_SET_FLAG":              2048,
	"VT_NO_DEFAULT_VALUE_FLAG": 4096,
	"VT_ON_UPDATE_NOW_FLAG":    8192,
	"VT_NUM_FLAG":              32768,
}

func (x Field_Flag) String() string {
	return proto.EnumName(Field_Flag_name, int32(x))
}

// Target describes what the client expects the tablet is.
// If the tablet does not match, an error is returned.
type Target struct {
	Keyspace   string              `protobuf:"bytes,1,opt,name=keyspace" json:"keyspace,omitempty"`
	Shard      string              `protobuf:"bytes,2,opt,name=shard" json:"shard,omitempty"`
	TabletType topodata.TabletType `protobuf:"varint,3,opt,name=tablet_type,enum=topodata.TabletType" json:"tablet_type,omitempty"`
}

func (m *Target) Reset()         { *m = Target{} }
func (m *Target) String() string { return proto.CompactTextString(m) }
func (*Target) ProtoMessage()    {}

// VTGateCallerID is sent by VTGate to VTTablet to describe the
// caller. If possible, this information is secure. For instance,
// if using unique certificates that guarantee that VTGate->VTTablet
// traffic cannot be spoofed, then VTTablet can trust this information,
// and VTTablet will use it for tablet ACLs, for instance.
// Because of this security guarantee, this is different than the CallerID
// structure, which is not secure at all, because it is provided
// by the Vitess client.
type VTGateCallerID struct {
	Username string `protobuf:"bytes,1,opt,name=username" json:"username,omitempty"`
}

func (m *VTGateCallerID) Reset()         { *m = VTGateCallerID{} }
func (m *VTGateCallerID) String() string { return proto.CompactTextString(m) }
func (*VTGateCallerID) ProtoMessage()    {}

// BindVariable represents a single bind variable in a Query
type BindVariable struct {
	Type BindVariable_Type `protobuf:"varint,1,opt,name=type,enum=query.BindVariable_Type" json:"type,omitempty"`
	// Depending on type, only one value below is set.
	ValueBytes     []byte    `protobuf:"bytes,2,opt,name=value_bytes,proto3" json:"value_bytes,omitempty"`
	ValueInt       int64     `protobuf:"varint,3,opt,name=value_int" json:"value_int,omitempty"`
	ValueUint      uint64    `protobuf:"varint,4,opt,name=value_uint" json:"value_uint,omitempty"`
	ValueFloat     float64   `protobuf:"fixed64,5,opt,name=value_float" json:"value_float,omitempty"`
	ValueBytesList [][]byte  `protobuf:"bytes,6,rep,name=value_bytes_list,proto3" json:"value_bytes_list,omitempty"`
	ValueIntList   []int64   `protobuf:"varint,7,rep,name=value_int_list" json:"value_int_list,omitempty"`
	ValueUintList  []uint64  `protobuf:"varint,8,rep,name=value_uint_list" json:"value_uint_list,omitempty"`
	ValueFloatList []float64 `protobuf:"fixed64,9,rep,name=value_float_list" json:"value_float_list,omitempty"`
}

func (m *BindVariable) Reset()         { *m = BindVariable{} }
func (m *BindVariable) String() string { return proto.CompactTextString(m) }
func (*BindVariable) ProtoMessage()    {}

// BoundQuery is a query with its bind variables
type BoundQuery struct {
	Sql           []byte                   `protobuf:"bytes,1,opt,name=sql,proto3" json:"sql,omitempty"`
	BindVariables map[string]*BindVariable `protobuf:"bytes,2,rep,name=bind_variables" json:"bind_variables,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *BoundQuery) Reset()         { *m = BoundQuery{} }
func (m *BoundQuery) String() string { return proto.CompactTextString(m) }
func (*BoundQuery) ProtoMessage()    {}

func (m *BoundQuery) GetBindVariables() map[string]*BindVariable {
	if m != nil {
		return m.BindVariables
	}
	return nil
}

// Field describes a single column returned by a query
type Field struct {
	// name of the field as returned by mysql C API
	Name string     `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Type Field_Type `protobuf:"varint,2,opt,name=type,enum=query.Field_Type" json:"type,omitempty"`
	// flags is essentially a bitset<Flag>.
	Flags int64 `protobuf:"varint,3,opt,name=flags" json:"flags,omitempty"`
}

func (m *Field) Reset()         { *m = Field{} }
func (m *Field) String() string { return proto.CompactTextString(m) }
func (*Field) ProtoMessage()    {}

// Row is a database row.
type Row struct {
	Values [][]byte `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
}

func (m *Row) Reset()         { *m = Row{} }
func (m *Row) String() string { return proto.CompactTextString(m) }
func (*Row) ProtoMessage()    {}

// QueryResult is returned by Execute and ExecuteStream.
//
// As returned by Execute, len(fields) is always equal to len(row)
// (for each row in rows).
//
// As returned by StreamExecute, the first QueryResult has the fields
// set, and subsequent QueryResult have rows set. And as Execute,
// len(QueryResult[0].fields) is always equal to len(row) (for each
// row in rows for each QueryResult in QueryResult[1:]).
type QueryResult struct {
	Fields       []*Field `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	RowsAffected uint64   `protobuf:"varint,2,opt,name=rows_affected" json:"rows_affected,omitempty"`
	InsertId     uint64   `protobuf:"varint,3,opt,name=insert_id" json:"insert_id,omitempty"`
	Rows         []*Row   `protobuf:"bytes,4,rep,name=rows" json:"rows,omitempty"`
}

func (m *QueryResult) Reset()         { *m = QueryResult{} }
func (m *QueryResult) String() string { return proto.CompactTextString(m) }
func (*QueryResult) ProtoMessage()    {}

func (m *QueryResult) GetFields() []*Field {
	if m != nil {
		return m.Fields
	}
	return nil
}

func (m *QueryResult) GetRows() []*Row {
	if m != nil {
		return m.Rows
	}
	return nil
}

// GetSessionIdRequest is the payload to GetSessionId
type GetSessionIdRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Keyspace          string          `protobuf:"bytes,3,opt,name=keyspace" json:"keyspace,omitempty"`
	Shard             string          `protobuf:"bytes,4,opt,name=shard" json:"shard,omitempty"`
}

func (m *GetSessionIdRequest) Reset()         { *m = GetSessionIdRequest{} }
func (m *GetSessionIdRequest) String() string { return proto.CompactTextString(m) }
func (*GetSessionIdRequest) ProtoMessage()    {}

func (m *GetSessionIdRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *GetSessionIdRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

// GetSessionIdResponse is the returned value from GetSessionId
type GetSessionIdResponse struct {
	SessionId int64 `protobuf:"varint,1,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *GetSessionIdResponse) Reset()         { *m = GetSessionIdResponse{} }
func (m *GetSessionIdResponse) String() string { return proto.CompactTextString(m) }
func (*GetSessionIdResponse) ProtoMessage()    {}

// ExecuteRequest is the payload to Execute
type ExecuteRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Target            *Target         `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	Query             *BoundQuery     `protobuf:"bytes,4,opt,name=query" json:"query,omitempty"`
	TransactionId     int64           `protobuf:"varint,5,opt,name=transaction_id" json:"transaction_id,omitempty"`
	SessionId         int64           `protobuf:"varint,6,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *ExecuteRequest) Reset()         { *m = ExecuteRequest{} }
func (m *ExecuteRequest) String() string { return proto.CompactTextString(m) }
func (*ExecuteRequest) ProtoMessage()    {}

func (m *ExecuteRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *ExecuteRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

func (m *ExecuteRequest) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (m *ExecuteRequest) GetQuery() *BoundQuery {
	if m != nil {
		return m.Query
	}
	return nil
}

// ExecuteResponse is the returned value from Execute
type ExecuteResponse struct {
	Result *QueryResult `protobuf:"bytes,1,opt,name=result" json:"result,omitempty"`
}

func (m *ExecuteResponse) Reset()         { *m = ExecuteResponse{} }
func (m *ExecuteResponse) String() string { return proto.CompactTextString(m) }
func (*ExecuteResponse) ProtoMessage()    {}

func (m *ExecuteResponse) GetResult() *QueryResult {
	if m != nil {
		return m.Result
	}
	return nil
}

// ExecuteBatchRequest is the payload to ExecuteBatch
type ExecuteBatchRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Target            *Target         `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	Queries           []*BoundQuery   `protobuf:"bytes,4,rep,name=queries" json:"queries,omitempty"`
	AsTransaction     bool            `protobuf:"varint,5,opt,name=as_transaction" json:"as_transaction,omitempty"`
	TransactionId     int64           `protobuf:"varint,6,opt,name=transaction_id" json:"transaction_id,omitempty"`
	SessionId         int64           `protobuf:"varint,7,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *ExecuteBatchRequest) Reset()         { *m = ExecuteBatchRequest{} }
func (m *ExecuteBatchRequest) String() string { return proto.CompactTextString(m) }
func (*ExecuteBatchRequest) ProtoMessage()    {}

func (m *ExecuteBatchRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *ExecuteBatchRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

func (m *ExecuteBatchRequest) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (m *ExecuteBatchRequest) GetQueries() []*BoundQuery {
	if m != nil {
		return m.Queries
	}
	return nil
}

// ExecuteBatchResponse is the returned value from ExecuteBatch
type ExecuteBatchResponse struct {
	Results []*QueryResult `protobuf:"bytes,1,rep,name=results" json:"results,omitempty"`
}

func (m *ExecuteBatchResponse) Reset()         { *m = ExecuteBatchResponse{} }
func (m *ExecuteBatchResponse) String() string { return proto.CompactTextString(m) }
func (*ExecuteBatchResponse) ProtoMessage()    {}

func (m *ExecuteBatchResponse) GetResults() []*QueryResult {
	if m != nil {
		return m.Results
	}
	return nil
}

// StreamExecuteRequest is the payload to StreamExecute
type StreamExecuteRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Target            *Target         `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	Query             *BoundQuery     `protobuf:"bytes,4,opt,name=query" json:"query,omitempty"`
	SessionId         int64           `protobuf:"varint,5,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *StreamExecuteRequest) Reset()         { *m = StreamExecuteRequest{} }
func (m *StreamExecuteRequest) String() string { return proto.CompactTextString(m) }
func (*StreamExecuteRequest) ProtoMessage()    {}

func (m *StreamExecuteRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *StreamExecuteRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

func (m *StreamExecuteRequest) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (m *StreamExecuteRequest) GetQuery() *BoundQuery {
	if m != nil {
		return m.Query
	}
	return nil
}

// StreamExecuteResponse is the returned value from StreamExecute
type StreamExecuteResponse struct {
	Result *QueryResult `protobuf:"bytes,1,opt,name=result" json:"result,omitempty"`
}

func (m *StreamExecuteResponse) Reset()         { *m = StreamExecuteResponse{} }
func (m *StreamExecuteResponse) String() string { return proto.CompactTextString(m) }
func (*StreamExecuteResponse) ProtoMessage()    {}

func (m *StreamExecuteResponse) GetResult() *QueryResult {
	if m != nil {
		return m.Result
	}
	return nil
}

// BeginRequest is the payload to Begin
type BeginRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Target            *Target         `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	SessionId         int64           `protobuf:"varint,4,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *BeginRequest) Reset()         { *m = BeginRequest{} }
func (m *BeginRequest) String() string { return proto.CompactTextString(m) }
func (*BeginRequest) ProtoMessage()    {}

func (m *BeginRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *BeginRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

func (m *BeginRequest) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

// BeginResponse is the returned value from Begin
type BeginResponse struct {
	TransactionId int64 `protobuf:"varint,1,opt,name=transaction_id" json:"transaction_id,omitempty"`
}

func (m *BeginResponse) Reset()         { *m = BeginResponse{} }
func (m *BeginResponse) String() string { return proto.CompactTextString(m) }
func (*BeginResponse) ProtoMessage()    {}

// CommitRequest is the payload to Commit
type CommitRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Target            *Target         `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	TransactionId     int64           `protobuf:"varint,4,opt,name=transaction_id" json:"transaction_id,omitempty"`
	SessionId         int64           `protobuf:"varint,5,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *CommitRequest) Reset()         { *m = CommitRequest{} }
func (m *CommitRequest) String() string { return proto.CompactTextString(m) }
func (*CommitRequest) ProtoMessage()    {}

func (m *CommitRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *CommitRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

func (m *CommitRequest) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

// CommitResponse is the returned value from Commit
type CommitResponse struct {
}

func (m *CommitResponse) Reset()         { *m = CommitResponse{} }
func (m *CommitResponse) String() string { return proto.CompactTextString(m) }
func (*CommitResponse) ProtoMessage()    {}

// RollbackRequest is the payload to Rollback
type RollbackRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Target            *Target         `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	TransactionId     int64           `protobuf:"varint,4,opt,name=transaction_id" json:"transaction_id,omitempty"`
	SessionId         int64           `protobuf:"varint,5,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *RollbackRequest) Reset()         { *m = RollbackRequest{} }
func (m *RollbackRequest) String() string { return proto.CompactTextString(m) }
func (*RollbackRequest) ProtoMessage()    {}

func (m *RollbackRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *RollbackRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

func (m *RollbackRequest) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

// RollbackResponse is the returned value from Rollback
type RollbackResponse struct {
}

func (m *RollbackResponse) Reset()         { *m = RollbackResponse{} }
func (m *RollbackResponse) String() string { return proto.CompactTextString(m) }
func (*RollbackResponse) ProtoMessage()    {}

// SplitQueryRequest is the payload for SplitQuery
type SplitQueryRequest struct {
	EffectiveCallerId *vtrpc.CallerID `protobuf:"bytes,1,opt,name=effective_caller_id" json:"effective_caller_id,omitempty"`
	ImmediateCallerId *VTGateCallerID `protobuf:"bytes,2,opt,name=immediate_caller_id" json:"immediate_caller_id,omitempty"`
	Target            *Target         `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	Query             *BoundQuery     `protobuf:"bytes,4,opt,name=query" json:"query,omitempty"`
	SplitColumn       string          `protobuf:"bytes,5,opt,name=split_column" json:"split_column,omitempty"`
	SplitCount        int64           `protobuf:"varint,6,opt,name=split_count" json:"split_count,omitempty"`
	SessionId         int64           `protobuf:"varint,7,opt,name=session_id" json:"session_id,omitempty"`
}

func (m *SplitQueryRequest) Reset()         { *m = SplitQueryRequest{} }
func (m *SplitQueryRequest) String() string { return proto.CompactTextString(m) }
func (*SplitQueryRequest) ProtoMessage()    {}

func (m *SplitQueryRequest) GetEffectiveCallerId() *vtrpc.CallerID {
	if m != nil {
		return m.EffectiveCallerId
	}
	return nil
}

func (m *SplitQueryRequest) GetImmediateCallerId() *VTGateCallerID {
	if m != nil {
		return m.ImmediateCallerId
	}
	return nil
}

func (m *SplitQueryRequest) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (m *SplitQueryRequest) GetQuery() *BoundQuery {
	if m != nil {
		return m.Query
	}
	return nil
}

// QuerySplit represents one query to execute on the tablet
type QuerySplit struct {
	// query is the query to execute
	Query *BoundQuery `protobuf:"bytes,1,opt,name=query" json:"query,omitempty"`
	// row_count is the approximate row count the query will return
	RowCount int64 `protobuf:"varint,2,opt,name=row_count" json:"row_count,omitempty"`
}

func (m *QuerySplit) Reset()         { *m = QuerySplit{} }
func (m *QuerySplit) String() string { return proto.CompactTextString(m) }
func (*QuerySplit) ProtoMessage()    {}

func (m *QuerySplit) GetQuery() *BoundQuery {
	if m != nil {
		return m.Query
	}
	return nil
}

// SplitQueryResponse is returned by SplitQuery and represents all the queries
// to execute in order to get the entire data set.
type SplitQueryResponse struct {
	Queries []*QuerySplit `protobuf:"bytes,1,rep,name=queries" json:"queries,omitempty"`
}

func (m *SplitQueryResponse) Reset()         { *m = SplitQueryResponse{} }
func (m *SplitQueryResponse) String() string { return proto.CompactTextString(m) }
func (*SplitQueryResponse) ProtoMessage()    {}

func (m *SplitQueryResponse) GetQueries() []*QuerySplit {
	if m != nil {
		return m.Queries
	}
	return nil
}

// StreamHealthRequest is the payload for StreamHealth
type StreamHealthRequest struct {
}

func (m *StreamHealthRequest) Reset()         { *m = StreamHealthRequest{} }
func (m *StreamHealthRequest) String() string { return proto.CompactTextString(m) }
func (*StreamHealthRequest) ProtoMessage()    {}

// RealtimeStats contains information about the tablet status
type RealtimeStats struct {
	// health_error is the last error we got from health check,
	// or empty is the server is healthy. This is used for subset selection,
	// we do not send queries to servers that are not healthy.
	HealthError string `protobuf:"bytes,1,opt,name=health_error" json:"health_error,omitempty"`
	// seconds_behind_master is populated for slaves only. It indicates
	// how far behind on (MySQL) replication a slave currently is.  It is used
	// by clients for subset selection (so we don't try to send traffic
	// to tablets that are too far behind).
	// TODO(mberlin): Let's switch it to int64 instead?
	SecondsBehindMaster uint32 `protobuf:"varint,2,opt,name=seconds_behind_master" json:"seconds_behind_master,omitempty"`
	// filtered_replication_synced_until_timestamp is populated for the receiving
	// master of an ongoing filtered replication only.
	// It is used to find out how far the receiving master lags behind the
	// source shard.
	FilteredReplicationSyncedUntilTimestamp int64 `protobuf:"varint,4,opt,name=filtered_replication_synced_until_timestamp" json:"filtered_replication_synced_until_timestamp,omitempty"`
	// cpu_usage is used for load-based balancing
	CpuUsage float64 `protobuf:"fixed64,3,opt,name=cpu_usage" json:"cpu_usage,omitempty"`
}

func (m *RealtimeStats) Reset()         { *m = RealtimeStats{} }
func (m *RealtimeStats) String() string { return proto.CompactTextString(m) }
func (*RealtimeStats) ProtoMessage()    {}

// StreamHealthResponse is streamed by StreamHealth on a regular basis
type StreamHealthResponse struct {
	// target is the current server type. Only queries with that exact Target
	// record will be accepted.
	Target *Target `protobuf:"bytes,1,opt,name=target" json:"target,omitempty"`
	// tablet_externally_reparented_timestamp contains the last time
	// tabletmanager.TabletExternallyReparented was called on this tablet,
	// or 0 if it was never called. This is meant to differentiate two tablets
	// that report a target.TabletType of MASTER, only the one with the latest
	// timestamp should be trusted.
	TabletExternallyReparentedTimestamp int64 `protobuf:"varint,2,opt,name=tablet_externally_reparented_timestamp" json:"tablet_externally_reparented_timestamp,omitempty"`
	// realtime_stats contains information about the tablet status
	RealtimeStats *RealtimeStats `protobuf:"bytes,3,opt,name=realtime_stats" json:"realtime_stats,omitempty"`
}

func (m *StreamHealthResponse) Reset()         { *m = StreamHealthResponse{} }
func (m *StreamHealthResponse) String() string { return proto.CompactTextString(m) }
func (*StreamHealthResponse) ProtoMessage()    {}

func (m *StreamHealthResponse) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (m *StreamHealthResponse) GetRealtimeStats() *RealtimeStats {
	if m != nil {
		return m.RealtimeStats
	}
	return nil
}

func init() {
	proto.RegisterEnum("query.BindVariable_Type", BindVariable_Type_name, BindVariable_Type_value)
	proto.RegisterEnum("query.Field_Type", Field_Type_name, Field_Type_value)
	proto.RegisterEnum("query.Field_Flag", Field_Flag_name, Field_Flag_value)
}
