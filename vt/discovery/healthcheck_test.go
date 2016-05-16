package discovery

import (
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/tabletserver/querytypes"
	"github.com/youtube/vitess/go/vt/tabletserver/tabletconn"
	"github.com/youtube/vitess/go/vt/topo"
	"golang.org/x/net/context"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

var connMap map[string]*fakeConn

func init() {
	tabletconn.RegisterDialer("fake_discovery", discoveryDialer)
	flag.Set("tablet_protocol", "fake_discovery")
	connMap = make(map[string]*fakeConn)
}

func TestHealthCheck(t *testing.T) {
	ep := topo.NewTablet(0, "cell", "a")
	ep.PortMap["vt"] = 1
	input := make(chan *querypb.StreamHealthResponse)
	fakeConn := createFakeConn(ep, input)
	t.Logf(`createFakeConn({Host: "a", PortMap: {"vt": 1}}, c)`)
	l := newListener()
	hc := NewHealthCheck(1*time.Millisecond, 1*time.Millisecond, time.Hour, "" /* statsSuffix */).(*HealthCheckImpl)
	hc.SetListener(l)
	hc.AddTablet("cell", "", ep)
	t.Logf(`hc = HealthCheck(); hc.AddTablet("cell", "", {Host: "a", PortMap: {"vt": 1}})`)

	// no tablet before getting first StreamHealthResponse
	tsList := hc.GetTabletStatsFromKeyspaceShard("k", "s")
	if len(tsList) != 0 {
		t.Errorf(`hc.GetTabletStatsFromKeyspaceShard("k", "s") = %+v; want empty`, tsList)
	}

	// one tablet after receiving a StreamHealthResponse
	shr := &querypb.StreamHealthResponse{
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Serving: true,
		TabletExternallyReparentedTimestamp: 10,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
	}
	want := &TabletStats{
		Tablet:  ep,
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Up:      true,
		Serving: true,
		Stats:   &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
		TabletExternallyReparentedTimestamp: 10,
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: MASTER}, Serving: true, TabletExternallyReparentedTimestamp: 10, {SecondsBehindMaster: 1, CpuUsage: 0.2}}`)
	res := <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	tsList = hc.GetTabletStatsFromKeyspaceShard("k", "s")
	if len(tsList) != 1 || !reflect.DeepEqual(tsList[0], want) {
		t.Errorf(`hc.GetTabletStatsFromKeyspaceShard("k", "s") = %+v; want %+v`, tsList, want)
	}
	epcsl := hc.CacheStatus()
	epcslWant := TabletsCacheStatusList{{
		Cell:   "cell",
		Target: &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		TabletsStats: TabletStatsList{{
			Tablet:  ep,
			Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
			Up:      true,
			Serving: true,
			Stats:   &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
			TabletExternallyReparentedTimestamp: 10,
		}},
	}}
	if !reflect.DeepEqual(epcsl, epcslWant) {
		t.Errorf(`hc.CacheStatus() = %+v; want %+v`, epcsl, epcslWant)
	}

	// TabletType changed
	shr = &querypb.StreamHealthResponse{
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Serving: true,
		TabletExternallyReparentedTimestamp: 0,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.5},
	}
	want = &TabletStats{
		Tablet:  ep,
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:      true,
		Serving: true,
		Stats:   &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.5},
		TabletExternallyReparentedTimestamp: 0,
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: REPLICA}, Serving: true, TabletExternallyReparentedTimestamp: 0, {SecondsBehindMaster: 1, CpuUsage: 0.5}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	tsList = hc.GetTabletStatsFromTarget("k", "s", topodatapb.TabletType_REPLICA)
	if len(tsList) != 1 || !reflect.DeepEqual(tsList[0], want) {
		t.Errorf(`hc.GetTabletStatsFromTarget("k", "s", REPLICA) = %+v; want %+v`, tsList, want)
	}

	// Serving & RealtimeStats changed
	shr = &querypb.StreamHealthResponse{
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Serving: false,
		TabletExternallyReparentedTimestamp: 0,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.3},
	}
	want = &TabletStats{
		Tablet:  ep,
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:      true,
		Serving: false,
		Stats:   &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.3},
		TabletExternallyReparentedTimestamp: 0,
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: REPLICA}, TabletExternallyReparentedTimestamp: 0, {SecondsBehindMaster: 1, CpuUsage: 0.3}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}

	// HealthError
	shr = &querypb.StreamHealthResponse{
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Serving: true,
		TabletExternallyReparentedTimestamp: 0,
		RealtimeStats:                       &querypb.RealtimeStats{HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3},
	}
	want = &TabletStats{
		Tablet:  ep,
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:      true,
		Serving: false,
		Stats:   &querypb.RealtimeStats{HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3},
		TabletExternallyReparentedTimestamp: 0,
		LastError:                           fmt.Errorf("vttablet error: some error"),
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: REPLICA}, Serving: true, TabletExternallyReparentedTimestamp: 0, {HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}

	// remove tablet
	hc.deleteConn(ep)
	close(fakeConn.hcChan)
	t.Logf(`hc.RemoveTablet({Host: "a", PortMap: {"vt": 1}})`)
	want = &TabletStats{
		Tablet:  ep,
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:      false,
		Serving: false,
		Stats:   &querypb.RealtimeStats{HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3},
		TabletExternallyReparentedTimestamp: 0,
		LastError:                           fmt.Errorf("recv error"),
	}
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	tsList = hc.GetTabletStatsFromKeyspaceShard("k", "s")
	if len(tsList) != 0 {
		t.Errorf(`hc.GetTabletStatsFromKeyspaceShard("k", "s") = %+v; want empty`, tsList)
	}
	// close healthcheck
	hc.Close()
}

func TestHealthCheckTimeout(t *testing.T) {
	timeout := 500 * time.Millisecond
	ep := topo.NewTablet(0, "cell", "a")
	ep.PortMap["vt"] = 1
	input := make(chan *querypb.StreamHealthResponse)
	createFakeConn(ep, input)
	t.Logf(`createFakeConn({Host: "a", PortMap: {"vt": 1}}, c)`)
	l := newListener()
	hc := NewHealthCheck(1*time.Millisecond, 1*time.Millisecond, timeout, "" /* statsSuffix */).(*HealthCheckImpl)
	hc.SetListener(l)
	hc.AddTablet("cell", "", ep)
	t.Logf(`hc = HealthCheck(); hc.AddTablet("cell", "", {Host: "a", PortMap: {"vt": 1}})`)

	// one tablet after receiving a StreamHealthResponse
	shr := &querypb.StreamHealthResponse{
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Serving: true,
		TabletExternallyReparentedTimestamp: 10,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
	}
	want := &TabletStats{
		Tablet:  ep,
		Target:  &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Up:      true,
		Serving: true,
		Stats:   &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
		TabletExternallyReparentedTimestamp: 10,
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: MASTER}, Serving: true, TabletExternallyReparentedTimestamp: 10, {SecondsBehindMaster: 1, CpuUsage: 0.2}}`)
	res := <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	tsList := hc.GetTabletStatsFromKeyspaceShard("k", "s")
	if len(tsList) != 1 || !reflect.DeepEqual(tsList[0], want) {
		t.Errorf(`hc.GetTabletStatsFromKeyspaceShard("k", "s") = %+v; want %+v`, tsList, want)
	}
	// wait for timeout period
	time.Sleep(2 * timeout)
	t.Logf(`Sleep(2 * timeout)`)
	res = <-l.output
	if res.Serving {
		t.Errorf(`<-l.output: %+v; want not serving`, res)
	}
	tsList = hc.GetTabletStatsFromKeyspaceShard("k", "s")
	if len(tsList) != 1 || tsList[0].Serving {
		t.Errorf(`hc.GetTabletStatsFromKeyspaceShard("k", "s") = %+v; want not serving`, tsList)
	}
	// send a healthcheck response, it should be serving again
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: MASTER}, Serving: true, TabletExternallyReparentedTimestamp: 10, {SecondsBehindMaster: 1, CpuUsage: 0.2}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	tsList = hc.GetTabletStatsFromKeyspaceShard("k", "s")
	if len(tsList) != 1 || !reflect.DeepEqual(tsList[0], want) {
		t.Errorf(`hc.GetTabletStatsFromKeyspaceShard("k", "s") = %+v; want %+v`, tsList, want)
	}
	// close healthcheck
	hc.Close()
}

type listener struct {
	output chan *TabletStats
}

func newListener() *listener {
	return &listener{output: make(chan *TabletStats, 1)}
}

func (l *listener) StatsUpdate(ts *TabletStats) {
	l.output <- ts
}

func createFakeConn(tablet *topodatapb.Tablet, c chan *querypb.StreamHealthResponse) *fakeConn {
	key := TabletToMapKey(tablet)
	conn := &fakeConn{tablet: tablet, hcChan: c}
	connMap[key] = conn
	return conn
}

func discoveryDialer(ctx context.Context, tablet *topodatapb.Tablet, keyspace, shard string, tabletType topodatapb.TabletType, timeout time.Duration) (tabletconn.TabletConn, error) {
	key := TabletToMapKey(tablet)
	return connMap[key], nil
}

type fakeConn struct {
	tablet *topodatapb.Tablet
	hcChan chan *querypb.StreamHealthResponse
}

type streamHealthReader struct {
	c <-chan *querypb.StreamHealthResponse
}

// Recv implements tabletconn.StreamHealthReader.
// It returns one response from the chan.
func (r *streamHealthReader) Recv() (*querypb.StreamHealthResponse, error) {
	resp, ok := <-r.c
	if !ok {
		return nil, fmt.Errorf("recv error")
	}
	return resp, nil
}

// StreamHealth implements tabletconn.TabletConn.
func (fc *fakeConn) StreamHealth(ctx context.Context) (tabletconn.StreamHealthReader, error) {
	return &streamHealthReader{
		c: fc.hcChan,
	}, nil
}

// Execute implements tabletconn.TabletConn.
func (fc *fakeConn) Execute(ctx context.Context, query string, bindVars map[string]interface{}, transactionID int64) (*sqltypes.Result, error) {
	return nil, fmt.Errorf("not implemented")
}

// ExecuteBatch implements tabletconn.TabletConn.
func (fc *fakeConn) ExecuteBatch(ctx context.Context, queries []querytypes.BoundQuery, asTransaction bool, transactionID int64) ([]sqltypes.Result, error) {
	return nil, fmt.Errorf("not implemented")
}

// StreamExecute implements tabletconn.TabletConn.
func (fc *fakeConn) StreamExecute(ctx context.Context, query string, bindVars map[string]interface{}) (sqltypes.ResultStream, error) {
	return nil, fmt.Errorf("not implemented")
}

// Begin implements tabletconn.TabletConn.
func (fc *fakeConn) Begin(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

// Commit implements tabletconn.TabletConn.
func (fc *fakeConn) Commit(ctx context.Context, transactionID int64) error {
	return fmt.Errorf("not implemented")
}

// Rollback implements tabletconn.TabletConn.
func (fc *fakeConn) Rollback(ctx context.Context, transactionID int64) error {
	return fmt.Errorf("not implemented")
}

// BeginExecute implements tabletconn.TabletConn.
func (fc *fakeConn) BeginExecute(ctx context.Context, query string, bindVars map[string]interface{}) (*sqltypes.Result, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// BeginExecuteBatch implements tabletconn.TabletConn.
func (fc *fakeConn) BeginExecuteBatch(ctx context.Context, queries []querytypes.BoundQuery, asTransaction bool) ([]sqltypes.Result, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// SplitQuery implements tabletconn.TabletConn.
func (fc *fakeConn) SplitQuery(ctx context.Context, query querytypes.BoundQuery, splitColumn string, splitCount int64) ([]querytypes.QuerySplit, error) {
	return nil, fmt.Errorf("not implemented")
}

// SplitQueryV2 implements tabletconn.TabletConn.
func (fc *fakeConn) SplitQueryV2(
	ctx context.Context,
	query querytypes.BoundQuery,
	splitColumn []string,
	splitCount int64,
	numRowsPerQueryPart int64,
	algorithm querypb.SplitQueryRequest_Algorithm,
) ([]querytypes.QuerySplit, error) {
	return nil, fmt.Errorf("not implemented")
}

// SetTarget implements tabletconn.TabletConn.
func (fc *fakeConn) SetTarget(keyspace, shard string, tabletType topodatapb.TabletType) error {
	return fmt.Errorf("not implemented")
}

// Tablet returns the tablet associated with the connection.
func (fc *fakeConn) Tablet() *topodatapb.Tablet {
	return fc.tablet
}

// Close closes the connection.
func (fc *fakeConn) Close() {
}
