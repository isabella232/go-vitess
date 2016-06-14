// Package gatewaytest contains a test suite to run against a Gateway object.
// We re-use the tabletconn test suite, as it tests all queries and parameters
// go through. There are two exceptions:
// - the health check: we just make that one work, so the gateway knows the
//   tablet is healthy.
// - the error type returned: it's not a TabletError any more, but a ShardError.
//   We still check the error code is correct though which is really all we care
//   about.
package gatewaytest

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/tabletserver/querytypes"
	"github.com/youtube/vitess/go/vt/tabletserver/tabletconn"
	"github.com/youtube/vitess/go/vt/tabletserver/tabletconntest"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/vtgate/gateway"
	"github.com/youtube/vitess/go/vt/zktopo/zktestserver"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// gatewayAdapter implements the TabletConn interface, but sends the
// queries to the Gateway.
type gatewayAdapter struct {
	g          gateway.Gateway
	keyspace   string
	shard      string
	tabletType topodatapb.TabletType
}

func (ga *gatewayAdapter) Execute(ctx context.Context, query string, bindVars map[string]interface{}, transactionID int64) (*sqltypes.Result, error) {
	return ga.g.Execute(ctx, ga.keyspace, ga.shard, ga.tabletType, query, bindVars, transactionID)
}

func (ga *gatewayAdapter) ExecuteBatch(ctx context.Context, queries []querytypes.BoundQuery, asTransaction bool, transactionID int64) ([]sqltypes.Result, error) {
	return ga.g.ExecuteBatch(ctx, ga.keyspace, ga.shard, ga.tabletType, queries, asTransaction, transactionID)
}

func (ga *gatewayAdapter) StreamExecute(ctx context.Context, query string, bindVars map[string]interface{}) (sqltypes.ResultStream, error) {
	return ga.g.StreamExecute(ctx, ga.keyspace, ga.shard, ga.tabletType, query, bindVars)
}

func (ga *gatewayAdapter) Begin(ctx context.Context) (transactionID int64, err error) {
	return ga.g.Begin(ctx, ga.keyspace, ga.shard, ga.tabletType)
}

func (ga *gatewayAdapter) Commit(ctx context.Context, transactionID int64) error {
	return ga.g.Commit(ctx, ga.keyspace, ga.shard, ga.tabletType, transactionID)
}

func (ga *gatewayAdapter) Rollback(ctx context.Context, transactionID int64) error {
	return ga.g.Rollback(ctx, ga.keyspace, ga.shard, ga.tabletType, transactionID)
}

func (ga *gatewayAdapter) BeginExecute(ctx context.Context, query string, bindVars map[string]interface{}) (result *sqltypes.Result, transactionID int64, err error) {
	return ga.g.BeginExecute(ctx, ga.keyspace, ga.shard, ga.tabletType, query, bindVars)
}

func (ga *gatewayAdapter) BeginExecuteBatch(ctx context.Context, queries []querytypes.BoundQuery, asTransaction bool) (results []sqltypes.Result, transactionID int64, err error) {
	return ga.g.BeginExecuteBatch(ctx, ga.keyspace, ga.shard, ga.tabletType, queries, asTransaction)
}

func (ga *gatewayAdapter) Close() {
}

func (ga *gatewayAdapter) SetTarget(keyspace, shard string, tabletType topodatapb.TabletType) error {
	return nil
}

func (ga *gatewayAdapter) Tablet() *topodatapb.Tablet {
	return &topodatapb.Tablet{}
}

func (ga *gatewayAdapter) SplitQuery(ctx context.Context, query querytypes.BoundQuery, splitColumn string, splitCount int64) ([]querytypes.QuerySplit, error) {
	return ga.g.SplitQuery(ctx, ga.keyspace, ga.shard, ga.tabletType, query.Sql, query.BindVariables, splitColumn, splitCount)
}

func (ga *gatewayAdapter) SplitQueryV2(ctx context.Context, query querytypes.BoundQuery, splitColumns []string, splitCount int64, numRowsPerQueryPart int64, algorithm querypb.SplitQueryRequest_Algorithm) (queries []querytypes.QuerySplit, err error) {
	return ga.g.SplitQueryV2(ctx, ga.keyspace, ga.shard, ga.tabletType, query.Sql, query.BindVariables, splitColumns, splitCount, numRowsPerQueryPart, algorithm)
}

func (ga *gatewayAdapter) StreamHealth(ctx context.Context) (tabletconn.StreamHealthReader, error) {
	return nil, fmt.Errorf("Not Implemented")
}

// CreateFakeServers returns the servers to use for these tests
func CreateFakeServers(t *testing.T) (*tabletconntest.FakeQueryService, topo.Server, string) {
	cell := "local"

	// the FakeServer is just slightly modified
	f := tabletconntest.CreateFakeServer(t)
	f.TestingGateway = true
	f.StreamHealthResponse = &querypb.StreamHealthResponse{
		Target:  tabletconntest.TestTarget,
		Serving: true,
		TabletExternallyReparentedTimestamp: 1234589,
		RealtimeStats: &querypb.RealtimeStats{
			SecondsBehindMaster: 1,
		},
	}

	// The topo server has a single SrvKeyspace
	ts := zktestserver.New(t, []string{cell})
	if err := ts.UpdateSrvKeyspace(context.Background(), cell, tabletconntest.TestTarget.Keyspace, &topodatapb.SrvKeyspace{
		Partitions: []*topodatapb.SrvKeyspace_KeyspacePartition{
			{
				ServedType: topodatapb.TabletType_MASTER,
				ShardReferences: []*topodatapb.ShardReference{
					{
						Name: tabletconntest.TestTarget.Shard,
					},
				},
			},
		},
	}); err != nil {
		t.Fatalf("can't add srvKeyspace: %v", err)
	}

	return f, ts, cell
}

// TestSuite executes a set of tests on the provided gateway. The provided
// gateway needs to be configured with one established connection for
// tabletconntest.TestTarget.{Keyspace, Shard, TabletType} to the
// provided tabletconntest.FakeQueryService.
func TestSuite(t *testing.T, name string, g gateway.Gateway, f *tabletconntest.FakeQueryService) {

	protocolName := "gateway-test-" + name

	tabletconn.RegisterDialer(protocolName, func(ctx context.Context, tablet *topodatapb.Tablet, keyspace, shard string, tabletType topodatapb.TabletType, timeout time.Duration) (tabletconn.TabletConn, error) {
		return &gatewayAdapter{
			g:          g,
			keyspace:   keyspace,
			shard:      shard,
			tabletType: tabletType,
		}, nil
	})

	tabletconntest.TestSuite(t, protocolName, &topodatapb.Tablet{}, f)
}
