// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grpcvtgateservice provides the gRPC glue for vtgate
package grpcvtgateservice

import (
	"google.golang.org/grpc"

	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/callerid"
	"github.com/youtube/vitess/go/vt/callinfo"
	"github.com/youtube/vitess/go/vt/servenv"
	tproto "github.com/youtube/vitess/go/vt/tabletserver/proto"
	"github.com/youtube/vitess/go/vt/vterrors"
	"github.com/youtube/vitess/go/vt/vtgate"
	"github.com/youtube/vitess/go/vt/vtgate/vtgateservice"
	"golang.org/x/net/context"

	pb "github.com/youtube/vitess/go/vt/proto/vtgate"
	pbs "github.com/youtube/vitess/go/vt/proto/vtgateservice"
)

// VTGate is the public structure that is exported via gRPC
type VTGate struct {
	server vtgateservice.VTGateService
}

// Execute is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) Execute(ctx context.Context, request *pb.ExecuteRequest) (response *pb.ExecuteResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return nil, vterrors.ToGRPCError(err)
	}
	result, err := vtg.server.Execute(ctx, string(request.Query.Sql), bv, request.TabletType, request.Session, request.NotInTransaction)
	return &pb.ExecuteResponse{
		Result:  sqltypes.ResultToProto3(result),
		Session: request.Session,
		Error:   vterrors.VtRPCErrorFromVtError(err),
	}, nil
}

// ExecuteShards is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) ExecuteShards(ctx context.Context, request *pb.ExecuteShardsRequest) (response *pb.ExecuteShardsResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return nil, vterrors.ToGRPCError(err)
	}
	result, err := vtg.server.ExecuteShards(ctx,
		string(request.Query.Sql),
		bv,
		request.Keyspace,
		request.Shards,
		request.TabletType,
		request.Session,
		request.NotInTransaction)
	return &pb.ExecuteShardsResponse{
		Result:  sqltypes.ResultToProto3(result),
		Session: request.Session,
		Error:   vterrors.VtRPCErrorFromVtError(err),
	}, nil
}

// ExecuteKeyspaceIds is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) ExecuteKeyspaceIds(ctx context.Context, request *pb.ExecuteKeyspaceIdsRequest) (response *pb.ExecuteKeyspaceIdsResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return nil, vterrors.ToGRPCError(err)
	}
	result, err := vtg.server.ExecuteKeyspaceIds(ctx,
		string(request.Query.Sql),
		bv,
		request.Keyspace,
		request.KeyspaceIds,
		request.TabletType,
		request.Session,
		request.NotInTransaction)
	return &pb.ExecuteKeyspaceIdsResponse{
		Result:  sqltypes.ResultToProto3(result),
		Session: request.Session,
		Error:   vterrors.VtRPCErrorFromVtError(err),
	}, nil
}

// ExecuteKeyRanges is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) ExecuteKeyRanges(ctx context.Context, request *pb.ExecuteKeyRangesRequest) (response *pb.ExecuteKeyRangesResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return nil, vterrors.ToGRPCError(err)
	}
	result, err := vtg.server.ExecuteKeyRanges(ctx,
		string(request.Query.Sql),
		bv,
		request.Keyspace,
		request.KeyRanges,
		request.TabletType,
		request.Session,
		request.NotInTransaction)
	return &pb.ExecuteKeyRangesResponse{
		Result:  sqltypes.ResultToProto3(result),
		Session: request.Session,
		Error:   vterrors.VtRPCErrorFromVtError(err),
	}, nil
}

// ExecuteEntityIds is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) ExecuteEntityIds(ctx context.Context, request *pb.ExecuteEntityIdsRequest) (response *pb.ExecuteEntityIdsResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return nil, vterrors.ToGRPCError(err)
	}
	result, err := vtg.server.ExecuteEntityIds(ctx,
		string(request.Query.Sql),
		bv,
		request.Keyspace,
		request.EntityColumnName,
		request.EntityKeyspaceIds,
		request.TabletType,
		request.Session,
		request.NotInTransaction)
	return &pb.ExecuteEntityIdsResponse{
		Result:  sqltypes.ResultToProto3(result),
		Session: request.Session,
		Error:   vterrors.VtRPCErrorFromVtError(err),
	}, nil
}

// ExecuteBatchShards is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) ExecuteBatchShards(ctx context.Context, request *pb.ExecuteBatchShardsRequest) (response *pb.ExecuteBatchShardsResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	result, err := vtg.server.ExecuteBatchShards(ctx,
		request.Queries,
		request.TabletType,
		request.AsTransaction,
		request.Session)
	return &pb.ExecuteBatchShardsResponse{
		Results: sqltypes.ResultsToProto3(result),
		Session: request.Session,
		Error:   vterrors.VtRPCErrorFromVtError(err),
	}, nil
}

// ExecuteBatchKeyspaceIds is the RPC version of
// vtgateservice.VTGateService method
func (vtg *VTGate) ExecuteBatchKeyspaceIds(ctx context.Context, request *pb.ExecuteBatchKeyspaceIdsRequest) (response *pb.ExecuteBatchKeyspaceIdsResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	result, err := vtg.server.ExecuteBatchKeyspaceIds(ctx,
		request.Queries,
		request.TabletType,
		request.AsTransaction,
		request.Session)
	return &pb.ExecuteBatchKeyspaceIdsResponse{
		Results: sqltypes.ResultsToProto3(result),
		Session: request.Session,
		Error:   vterrors.VtRPCErrorFromVtError(err),
	}, nil
}

// StreamExecute is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) StreamExecute(request *pb.StreamExecuteRequest, stream pbs.Vitess_StreamExecuteServer) (err error) {
	defer vtg.server.HandlePanic(&err)
	ctx := callerid.NewContext(callinfo.GRPCCallInfo(stream.Context()),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return vterrors.ToGRPCError(err)
	}
	vtgErr := vtg.server.StreamExecute(ctx,
		string(request.Query.Sql),
		bv,
		request.TabletType,
		func(value *sqltypes.Result) error {
			return stream.Send(&pb.StreamExecuteResponse{
				Result: sqltypes.ResultToProto3(value),
			})
		})
	return vterrors.ToGRPCError(vtgErr)
}

// StreamExecuteShards is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) StreamExecuteShards(request *pb.StreamExecuteShardsRequest, stream pbs.Vitess_StreamExecuteShardsServer) (err error) {
	defer vtg.server.HandlePanic(&err)
	ctx := callerid.NewContext(callinfo.GRPCCallInfo(stream.Context()),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return vterrors.ToGRPCError(err)
	}
	vtgErr := vtg.server.StreamExecuteShards(ctx,
		string(request.Query.Sql),
		bv,
		request.Keyspace,
		request.Shards,
		request.TabletType,
		func(value *sqltypes.Result) error {
			return stream.Send(&pb.StreamExecuteShardsResponse{
				Result: sqltypes.ResultToProto3(value),
			})
		})
	return vterrors.ToGRPCError(vtgErr)
}

// StreamExecuteKeyspaceIds is the RPC version of
// vtgateservice.VTGateService method
func (vtg *VTGate) StreamExecuteKeyspaceIds(request *pb.StreamExecuteKeyspaceIdsRequest, stream pbs.Vitess_StreamExecuteKeyspaceIdsServer) (err error) {
	defer vtg.server.HandlePanic(&err)
	ctx := callerid.NewContext(callinfo.GRPCCallInfo(stream.Context()),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return vterrors.ToGRPCError(err)
	}
	vtgErr := vtg.server.StreamExecuteKeyspaceIds(ctx,
		string(request.Query.Sql),
		bv,
		request.Keyspace,
		request.KeyspaceIds,
		request.TabletType,
		func(value *sqltypes.Result) error {
			return stream.Send(&pb.StreamExecuteKeyspaceIdsResponse{
				Result: sqltypes.ResultToProto3(value),
			})
		})
	return vterrors.ToGRPCError(vtgErr)
}

// StreamExecuteKeyRanges is the RPC version of
// vtgateservice.VTGateService method
func (vtg *VTGate) StreamExecuteKeyRanges(request *pb.StreamExecuteKeyRangesRequest, stream pbs.Vitess_StreamExecuteKeyRangesServer) (err error) {
	defer vtg.server.HandlePanic(&err)
	ctx := callerid.NewContext(callinfo.GRPCCallInfo(stream.Context()),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return vterrors.ToGRPCError(err)
	}
	vtgErr := vtg.server.StreamExecuteKeyRanges(ctx,
		string(request.Query.Sql),
		bv,
		request.Keyspace,
		request.KeyRanges,
		request.TabletType,
		func(value *sqltypes.Result) error {
			return stream.Send(&pb.StreamExecuteKeyRangesResponse{
				Result: sqltypes.ResultToProto3(value),
			})
		})
	return vterrors.ToGRPCError(vtgErr)
}

// Begin is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) Begin(ctx context.Context, request *pb.BeginRequest) (response *pb.BeginResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	outSession := new(pb.Session)
	vtgErr := vtg.server.Begin(ctx, outSession)
	response = &pb.BeginResponse{}
	if vtgErr == nil {
		response.Session = outSession
		return response, nil
	}
	return nil, vterrors.ToGRPCError(vtgErr)
}

// Commit is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) Commit(ctx context.Context, request *pb.CommitRequest) (response *pb.CommitResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	vtgErr := vtg.server.Commit(ctx, request.Session)
	response = &pb.CommitResponse{}
	if vtgErr == nil {
		return response, nil
	}
	return nil, vterrors.ToGRPCError(vtgErr)
}

// Rollback is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) Rollback(ctx context.Context, request *pb.RollbackRequest) (response *pb.RollbackResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	vtgErr := vtg.server.Rollback(ctx, request.Session)
	response = &pb.RollbackResponse{}
	if vtgErr == nil {
		return response, nil
	}
	return nil, vterrors.ToGRPCError(vtgErr)
}

// SplitQuery is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) SplitQuery(ctx context.Context, request *pb.SplitQueryRequest) (response *pb.SplitQueryResponse, err error) {

	defer vtg.server.HandlePanic(&err)
	ctx = callerid.NewContext(callinfo.GRPCCallInfo(ctx),
		request.CallerId,
		callerid.NewImmediateCallerID("grpc client"))
	bv, err := tproto.Proto3ToBindVariables(request.Query.BindVariables)
	if err != nil {
		return nil, vterrors.ToGRPCError(err)
	}
	splits, vtgErr := vtg.server.SplitQuery(ctx,
		request.Keyspace,
		string(request.Query.Sql),
		bv,
		request.SplitColumn,
		int(request.SplitCount))
	if vtgErr != nil {
		return nil, vterrors.ToGRPCError(vtgErr)
	}
	return &pb.SplitQueryResponse{
		Splits: splits,
	}, nil
}

// GetSrvKeyspace is the RPC version of vtgateservice.VTGateService method
func (vtg *VTGate) GetSrvKeyspace(ctx context.Context, request *pb.GetSrvKeyspaceRequest) (response *pb.GetSrvKeyspaceResponse, err error) {
	defer vtg.server.HandlePanic(&err)
	sk, vtgErr := vtg.server.GetSrvKeyspace(ctx, request.Keyspace)
	if vtgErr != nil {
		return nil, vterrors.ToGRPCError(vtgErr)
	}
	return &pb.GetSrvKeyspaceResponse{
		SrvKeyspace: sk,
	}, nil
}

func init() {
	vtgate.RegisterVTGates = append(vtgate.RegisterVTGates, func(vtGate vtgateservice.VTGateService) {
		if servenv.GRPCCheckServiceMap("vtgateservice") {
			pbs.RegisterVitessServer(servenv.GRPCServer, &VTGate{vtGate})
		}
	})
}

// RegisterForTest registers the gRPC implementation on the gRPC
// server.  Useful for unit tests only, for real use, the init()
// function does the registration.
func RegisterForTest(s *grpc.Server, service vtgateservice.VTGateService) {
	pbs.RegisterVitessServer(s, &VTGate{service})
}
