// Copyright 2013, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmclient

import (
	"flag"
	"time"

	log "github.com/golang/glog"
	mproto "github.com/youtube/vitess/go/mysql/proto"
	blproto "github.com/youtube/vitess/go/vt/binlog/proto"
	"github.com/youtube/vitess/go/vt/hook"
	"github.com/youtube/vitess/go/vt/logutil"
	myproto "github.com/youtube/vitess/go/vt/mysqlctl/proto"
	"github.com/youtube/vitess/go/vt/tabletmanager/actionnode"
	"github.com/youtube/vitess/go/vt/topo"
	"golang.org/x/net/context"
)

var tabletManagerProtocol = flag.String("tablet_manager_protocol", "bson", "the protocol to use to talk to vttablet")

// ErrFunc is used by streaming RPCs that don't return a specific result
type ErrFunc func() error

// SnapshotReplyFunc is used by Snapshot to return result and error
type SnapshotReplyFunc func() (*actionnode.SnapshotReply, error)

// TabletManagerClient defines the interface used to talk to a remote tablet
type TabletManagerClient interface {
	//
	// Various read-only methods
	//

	// Ping will try to ping the remote tablet
	Ping(ctx context.Context, tablet *topo.TabletInfo) error

	// GetSchema asks the remote tablet for its database schema
	GetSchema(ctx context.Context, tablet *topo.TabletInfo, tables, excludeTables []string, includeViews bool) (*myproto.SchemaDefinition, error)

	// GetPermissions asks the remote tablet for its permissions list
	GetPermissions(ctx context.Context, tablet *topo.TabletInfo) (*myproto.Permissions, error)

	//
	// Various read-write methods
	//

	// SetReadOnly makes the mysql instance read-only
	SetReadOnly(ctx context.Context, tablet *topo.TabletInfo) error

	// SetReadWrite makes the mysql instance read-write
	SetReadWrite(ctx context.Context, tablet *topo.TabletInfo) error

	// ChangeType asks the remote tablet to change its type
	ChangeType(ctx context.Context, tablet *topo.TabletInfo, dbType topo.TabletType) error

	// Scrap scraps the live running tablet
	Scrap(ctx context.Context, tablet *topo.TabletInfo) error

	// Sleep will sleep for a duration (used for tests)
	Sleep(ctx context.Context, tablet *topo.TabletInfo, duration time.Duration) error

	// ExecuteHook executes the provided hook remotely
	ExecuteHook(ctx context.Context, tablet *topo.TabletInfo, hk *hook.Hook) (*hook.HookResult, error)

	// RefreshState asks the remote tablet to reload its tablet record
	RefreshState(ctx context.Context, tablet *topo.TabletInfo) error

	// RunHealthCheck asks the remote tablet to run a health check cycle
	RunHealthCheck(ctx context.Context, tablet *topo.TabletInfo, targetTabletType topo.TabletType) error

	// HealthStream asks the tablet to stream its health status on
	// a regular basis
	HealthStream(ctx context.Context, tablet *topo.TabletInfo) (<-chan *actionnode.HealthStreamReply, ErrFunc, error)

	// ReloadSchema asks the remote tablet to reload its schema
	ReloadSchema(ctx context.Context, tablet *topo.TabletInfo) error

	// PreflightSchema will test a schema change
	PreflightSchema(ctx context.Context, tablet *topo.TabletInfo, change string) (*myproto.SchemaChangeResult, error)

	// ApplySchema will apply a schema change
	ApplySchema(ctx context.Context, tablet *topo.TabletInfo, change *myproto.SchemaChange) (*myproto.SchemaChangeResult, error)

	// ExecuteFetchAsDba executes a query remotely using the DBA pool
	ExecuteFetchAsDba(ctx context.Context, tablet *topo.TabletInfo, query string, maxRows int, wantFields, disableBinlogs bool) (*mproto.QueryResult, error)

	// ExecuteFetchAsApp executes a query remotely using the App pool
	ExecuteFetchAsApp(ctx context.Context, tablet *topo.TabletInfo, query string, maxRows int, wantFields bool) (*mproto.QueryResult, error)

	//
	// Replication related methods
	//

	// SlaveStatus returns the tablet's mysql slave status.
	SlaveStatus(ctx context.Context, tablet *topo.TabletInfo) (*myproto.ReplicationStatus, error)

	// MasterPosition returns the tablet's master position
	MasterPosition(ctx context.Context, tablet *topo.TabletInfo) (myproto.ReplicationPosition, error)

	// StopSlave stops the mysql replication
	StopSlave(ctx context.Context, tablet *topo.TabletInfo) error

	// StopSlaveMinimum stops the mysql replication after it reaches
	// the provided minimum point
	StopSlaveMinimum(ctx context.Context, tablet *topo.TabletInfo, stopPos myproto.ReplicationPosition, waitTime time.Duration) (*myproto.ReplicationStatus, error)

	// StartSlave starts the mysql replication
	StartSlave(ctx context.Context, tablet *topo.TabletInfo) error

	// TabletExternallyReparented tells a tablet it is now the master, after an
	// external tool has already promoted the underlying mysqld to master and
	// reparented the other mysqld servers to it.
	//
	// externalID is an optional string provided by the external tool that
	// vttablet will emit in logs to facilitate cross-referencing.
	TabletExternallyReparented(ctx context.Context, tablet *topo.TabletInfo, externalID string) error

	// GetSlaves returns the addresses of the slaves
	GetSlaves(ctx context.Context, tablet *topo.TabletInfo) ([]string, error)

	// WaitBlpPosition asks the tablet to wait until it reaches that
	// position in replication
	WaitBlpPosition(ctx context.Context, tablet *topo.TabletInfo, blpPosition blproto.BlpPosition, waitTime time.Duration) error

	// StopBlp asks the tablet to stop all its binlog players,
	// and returns the current position for all of them
	StopBlp(ctx context.Context, tablet *topo.TabletInfo) (*blproto.BlpPositionList, error)

	// StartBlp asks the tablet to restart its binlog players
	StartBlp(ctx context.Context, tablet *topo.TabletInfo) error

	// RunBlpUntil asks the tablet to restart its binlog players until
	// it reaches the given positions, if not there yet.
	RunBlpUntil(ctx context.Context, tablet *topo.TabletInfo, positions *blproto.BlpPositionList, waitTime time.Duration) (myproto.ReplicationPosition, error)

	//
	// Reparenting related functions
	//

	// ResetReplication tells a tablet to completely reset its
	// replication.  All binary and relay logs are flushed. All
	// replication positions are reset.
	ResetReplication(ctx context.Context, tablet *topo.TabletInfo) error

	// InitMaster tells a tablet to make itself the new master,
	// and return the replication position the slaves should use to
	// reparent to it.
	InitMaster(ctx context.Context, tablet *topo.TabletInfo) (myproto.ReplicationPosition, error)

	// PopulateReparentJournal asks the master to insert a row in
	// its reparent_journal table.
	PopulateReparentJournal(ctx context.Context, tablet *topo.TabletInfo, timeCreatedNS int64, actionName string, masterAlias topo.TabletAlias, pos myproto.ReplicationPosition) error

	// InitSlave tells a tablet to make itself a slave to the
	// passed in master tablet alias, and wait for the row in the
	// reparent_journal table.
	InitSlave(ctx context.Context, tablet *topo.TabletInfo, parent topo.TabletAlias, replicationPosition myproto.ReplicationPosition, timeCreatedNS int64) error

	// DemoteMaster tells the soon-to-be-former master it's gonna change,
	// and it should go read-only and return its current position.
	DemoteMaster(ctx context.Context, tablet *topo.TabletInfo) (myproto.ReplicationPosition, error)

	// PromoteSlaveWhenCaughtUp transforms the tablet from a slave to a master.
	PromoteSlaveWhenCaughtUp(ctx context.Context, tablet *topo.TabletInfo, pos myproto.ReplicationPosition) (myproto.ReplicationPosition, error)

	// SlaveWasPromoted tells the remote tablet it is now the master
	SlaveWasPromoted(ctx context.Context, tablet *topo.TabletInfo) error

	// SetMaster tells a tablet to make itself a slave to the
	// passed in master tablet alias, and wait for the row in the
	// reparent_journal table.
	SetMaster(ctx context.Context, tablet *topo.TabletInfo, parent topo.TabletAlias, timeCreatedNS int64) error

	// SlaveWasRestarted tells the remote tablet its master has changed
	SlaveWasRestarted(ctx context.Context, tablet *topo.TabletInfo, args *actionnode.SlaveWasRestartedArgs) error

	// StopReplicationAndGetPosition stops replication and returns the
	// current position.
	StopReplicationAndGetPosition(ctx context.Context, tablet *topo.TabletInfo) (myproto.ReplicationPosition, error)

	// PromoteSlave makes the tablet the new master
	PromoteSlave(ctx context.Context, tablet *topo.TabletInfo) (myproto.ReplicationPosition, error)

	//
	// Backup / restore related methods
	//

	// Snapshot takes a database snapshot
	Snapshot(ctx context.Context, tablet *topo.TabletInfo, sa *actionnode.SnapshotArgs) (<-chan *logutil.LoggerEvent, SnapshotReplyFunc, error)

	// SnapshotSourceEnd restarts the mysql server
	SnapshotSourceEnd(ctx context.Context, tablet *topo.TabletInfo, ssea *actionnode.SnapshotSourceEndArgs) error

	// ReserveForRestore will prepare a server for restore
	ReserveForRestore(ctx context.Context, tablet *topo.TabletInfo, rfra *actionnode.ReserveForRestoreArgs) error

	// Restore restores a database snapshot
	Restore(ctx context.Context, tablet *topo.TabletInfo, sa *actionnode.RestoreArgs) (<-chan *logutil.LoggerEvent, ErrFunc, error)

	//
	// RPC related methods
	//

	// IsTimeoutError checks if an error was caused by an RPC layer timeout vs an application-specific one
	IsTimeoutError(err error) bool
}

// TabletManagerClientFactory is the factory method to create
// TabletManagerClient objects.
type TabletManagerClientFactory func() TabletManagerClient

var tabletManagerClientFactories = make(map[string]TabletManagerClientFactory)

// RegisterTabletManagerClientFactory allows modules to register
// TabletManagerClient implementations. Should be called on init().
func RegisterTabletManagerClientFactory(name string, factory TabletManagerClientFactory) {
	if _, ok := tabletManagerClientFactories[name]; ok {
		log.Fatalf("RegisterTabletManagerClient %s already exists", name)
	}
	tabletManagerClientFactories[name] = factory
}

// NewTabletManagerClient creates a new TabletManagerClient. Should be
// called after flags are parsed.
func NewTabletManagerClient() TabletManagerClient {
	f, ok := tabletManagerClientFactories[*tabletManagerProtocol]
	if !ok {
		log.Fatalf("No TabletManagerProtocol registered with name %s", *tabletManagerProtocol)
	}

	return f()
}
