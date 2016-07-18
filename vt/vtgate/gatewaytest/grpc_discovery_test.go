package gatewaytest

import (
	"flag"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/youtube/vitess/go/vt/discovery"
	"github.com/youtube/vitess/go/vt/tabletserver/grpcqueryservice"
	"github.com/youtube/vitess/go/vt/tabletserver/tabletconntest"
	"github.com/youtube/vitess/go/vt/vtgate/gateway"
	"github.com/youtube/vitess/go/vt/vtgate/l2vtgate"

	// We will use gRPC to connect, register the dialer
	_ "github.com/youtube/vitess/go/vt/tabletserver/grpctabletconn"

	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// TestGRPCDiscovery tests the discovery gateway with a gRPC
// connection from the gateway to the fake tablet.
func TestGRPCDiscovery(t *testing.T) {
	flag.Set("tablet_protocol", "grpc")
	flag.Set("gateway_implementation", "discoverygateway")

	// Fake services for the tablet, topo server.
	service, ts, cell := CreateFakeServers(t)

	// Tablet: listen on a random port.
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Cannot listen: %v", err)
	}
	host := listener.Addr().(*net.TCPAddr).IP.String()
	port := listener.Addr().(*net.TCPAddr).Port
	defer listener.Close()

	// Tablet: create a gRPC server and listen on the port.
	server := grpc.NewServer()
	grpcqueryservice.Register(server, service)
	go server.Serve(listener)
	defer server.Stop()

	// VTGate: create the discovery healthcheck, and the gateway.
	// Wait for the right tablets to be present.
	hc := discovery.NewHealthCheck(30*time.Second, 10*time.Second, 2*time.Minute)
	hc.AddTablet(&topodatapb.Tablet{
		Alias: &topodatapb.TabletAlias{
			Cell: cell,
		},
		Keyspace: tabletconntest.TestTarget.Keyspace,
		Shard:    tabletconntest.TestTarget.Shard,
		Type:     tabletconntest.TestTarget.TabletType,
		Hostname: host,
		PortMap: map[string]int32{
			"grpc": int32(port),
		},
	}, "test_tablet")
	dg := gateway.GetCreator()(hc, ts, ts, cell, 2, []topodatapb.TabletType{tabletconntest.TestTarget.TabletType})

	// and run the test suite.
	TestSuite(t, "discovery-grpc", dg, service)
}

// TestL2VTGateDiscovery tests the l2vtgate gateway with a gRPC
// connection from the gateway to a l2vtgate in-process object.
func TestL2VTGateDiscovery(t *testing.T) {
	flag.Set("tablet_protocol", "grpc")
	flag.Set("gateway_implementation", "discoverygateway")

	// Fake services for the tablet, topo server.
	service, ts, cell := CreateFakeServers(t)

	// Tablet: listen on a random port.
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Cannot listen: %v", err)
	}
	host := listener.Addr().(*net.TCPAddr).IP.String()
	port := listener.Addr().(*net.TCPAddr).Port
	defer listener.Close()

	// Tablet: create a gRPC server and listen on the port.
	server := grpc.NewServer()
	grpcqueryservice.Register(server, service)
	go server.Serve(listener)
	defer server.Stop()

	// L2VTGate: Create the discovery healthcheck, and the gateway.
	// Wait for the right tablets to be present.
	hc := discovery.NewHealthCheck(30*time.Second, 10*time.Second, 2*time.Minute)
	hc.AddTablet(&topodatapb.Tablet{
		Alias: &topodatapb.TabletAlias{
			Cell: cell,
		},
		Keyspace: tabletconntest.TestTarget.Keyspace,
		Shard:    tabletconntest.TestTarget.Shard,
		Type:     tabletconntest.TestTarget.TabletType,
		Hostname: host,
		PortMap: map[string]int32{
			"grpc": int32(port),
		},
	}, "test_tablet")
	l2vtgate := l2vtgate.Init(hc, ts, ts, cell, 2, []topodatapb.TabletType{tabletconntest.TestTarget.TabletType})

	// L2VTGate: listen on a random port.
	listener, err = net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Cannot listen: %v", err)
	}
	defer listener.Close()

	// L2VTGate: create a gRPC server and listen on the port.
	server = grpc.NewServer()
	grpcqueryservice.Register(server, l2vtgate)
	go server.Serve(listener)
	defer server.Stop()

	// VTGate: create the l2vtgate gateway
	flag.Set("gateway_implementation", "l2vtgategateway")
	lg := gateway.GetCreator()(nil, ts, nil, "", 2, nil)
	lg.(*gateway.L2VTGateGateway).AddL2VTGateConn(listener.Addr().String(), tabletconntest.TestTarget.Keyspace, tabletconntest.TestTarget.Shard)

	// and run the test suite.
	TestSuite(t, "l2vtgate-grpc", lg, service)
}
