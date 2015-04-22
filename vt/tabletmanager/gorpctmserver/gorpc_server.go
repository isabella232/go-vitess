// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gorpctmserver

import (
	"sync"
	"time"

	mproto "github.com/youtube/vitess/go/mysql/proto"
	"github.com/youtube/vitess/go/rpcplus"
	blproto "github.com/youtube/vitess/go/vt/binlog/proto"
	"github.com/youtube/vitess/go/vt/callinfo"
	"github.com/youtube/vitess/go/vt/hook"
	"github.com/youtube/vitess/go/vt/logutil"
	myproto "github.com/youtube/vitess/go/vt/mysqlctl/proto"
	"github.com/youtube/vitess/go/vt/rpc"
	"github.com/youtube/vitess/go/vt/servenv"
	"github.com/youtube/vitess/go/vt/tabletmanager"
	"github.com/youtube/vitess/go/vt/tabletmanager/actionnode"
	"github.com/youtube/vitess/go/vt/tabletmanager/gorpcproto"
	"github.com/youtube/vitess/go/vt/topo"
	"golang.org/x/net/context"
)

// TabletManager is the Go RPC implementation of the RPC service
type TabletManager struct {
	// implementation of the agent to call
	agent tabletmanager.RPCAgent
}

//
// Various read-only methods
//

// Ping wraps RPCAgent.
func (tm *TabletManager) Ping(ctx context.Context, args, reply *string) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionPing, args, reply, func() error {
		*reply = tm.agent.Ping(ctx, *args)
		return nil
	})
}

// Sleep wraps RPCAgent.
func (tm *TabletManager) Sleep(ctx context.Context, args *time.Duration, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionSleep, args, reply, true, func() error {
		tm.agent.Sleep(ctx, *args)
		return nil
	})
}

// ExecuteHook wraps RPCAgent.
func (tm *TabletManager) ExecuteHook(ctx context.Context, args *hook.Hook, reply *hook.HookResult) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionExecuteHook, args, reply, true, func() error {
		*reply = *tm.agent.ExecuteHook(ctx, args)
		return nil
	})
}

// GetSchema wraps RPCAgent.
func (tm *TabletManager) GetSchema(ctx context.Context, args *gorpcproto.GetSchemaArgs, reply *myproto.SchemaDefinition) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionGetSchema, args, reply, func() error {
		sd, err := tm.agent.GetSchema(ctx, args.Tables, args.ExcludeTables, args.IncludeViews)
		if err == nil {
			*reply = *sd
		}
		return err
	})
}

// GetPermissions wraps RPCAgent.
func (tm *TabletManager) GetPermissions(ctx context.Context, args *rpc.Unused, reply *myproto.Permissions) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionGetPermissions, args, reply, func() error {
		p, err := tm.agent.GetPermissions(ctx)
		if err == nil {
			*reply = *p
		}
		return err
	})
}

//
// Various read-write methods
//

// SetReadOnly wraps RPCAgent.
func (tm *TabletManager) SetReadOnly(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionSetReadOnly, args, reply, true, func() error {
		return tm.agent.SetReadOnly(ctx, true)
	})
}

// SetReadWrite wraps RPCAgent.
func (tm *TabletManager) SetReadWrite(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionSetReadWrite, args, reply, true, func() error {
		return tm.agent.SetReadOnly(ctx, false)
	})
}

// ChangeType wraps RPCAgent.
func (tm *TabletManager) ChangeType(ctx context.Context, args *topo.TabletType, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionChangeType, args, reply, true, func() error {
		return tm.agent.ChangeType(ctx, *args)
	})
}

// Scrap wraps RPCAgent.
func (tm *TabletManager) Scrap(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionScrap, args, reply, true, func() error {
		return tm.agent.Scrap(ctx)
	})
}

// RefreshState wraps RPCAgent.
func (tm *TabletManager) RefreshState(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionRefreshState, args, reply, true, func() error {
		tm.agent.RefreshState(ctx)
		return nil
	})
}

// RunHealthCheck wraps RPCAgent.
func (tm *TabletManager) RunHealthCheck(ctx context.Context, args *topo.TabletType, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionRunHealthCheck, args, reply, func() error {
		tm.agent.RunHealthCheck(ctx, *args)
		return nil
	})
}

// HealthStream wraps RPCAgent.
func (tm *TabletManager) HealthStream(ctx context.Context, args *rpc.Unused, sendReply func(interface{}) error) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionHealthStream, args, nil, func() error {
		c := make(chan *actionnode.HealthStreamReply, 10)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for hsr := range c {
				// we send until the client disconnects
				if err := sendReply(hsr); err != nil {
					return
				}
			}
		}()

		id, err := tm.agent.RegisterHealthStream(c)
		if err != nil {
			close(c)
			wg.Wait()
			return err
		}
		wg.Wait()
		return tm.agent.UnregisterHealthStream(id)
	})
}

// ReloadSchema wraps RPCAgent.
func (tm *TabletManager) ReloadSchema(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionReloadSchema, args, reply, true, func() error {
		tm.agent.ReloadSchema(ctx)
		return nil
	})
}

// PreflightSchema wraps RPCAgent.
func (tm *TabletManager) PreflightSchema(ctx context.Context, args *string, reply *myproto.SchemaChangeResult) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionPreflightSchema, args, reply, true, func() error {
		scr, err := tm.agent.PreflightSchema(ctx, *args)
		if err == nil {
			*reply = *scr
		}
		return err
	})
}

// ApplySchema wraps RPCAgent.
func (tm *TabletManager) ApplySchema(ctx context.Context, args *myproto.SchemaChange, reply *myproto.SchemaChangeResult) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionApplySchema, args, reply, true, func() error {
		scr, err := tm.agent.ApplySchema(ctx, args)
		if err == nil {
			*reply = *scr
		}
		return err
	})
}

// ExecuteFetch wraps RPCAgent.
func (tm *TabletManager) ExecuteFetch(ctx context.Context, args *gorpcproto.ExecuteFetchArgs, reply *mproto.QueryResult) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionExecuteFetch, args, reply, func() error {
		qr, err := tm.agent.ExecuteFetch(ctx, args.Query, args.MaxRows, args.WantFields, args.DisableBinlogs, args.DBConfigName)
		if err == nil {
			*reply = *qr
		}
		return err
	})
}

//
// Replication related methods
//

// SlaveStatus wraps RPCAgent.
func (tm *TabletManager) SlaveStatus(ctx context.Context, args *rpc.Unused, reply *myproto.ReplicationStatus) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionSlaveStatus, args, reply, func() error {
		status, err := tm.agent.SlaveStatus(ctx)
		if err == nil {
			*reply = *status
		}
		return err
	})
}

// WaitSlavePosition wraps RPCAgent.
func (tm *TabletManager) WaitSlavePosition(ctx context.Context, args *gorpcproto.WaitSlavePositionArgs, reply *myproto.ReplicationStatus) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionWaitSlavePosition, args, reply, true, func() error {
		status, err := tm.agent.WaitSlavePosition(ctx, args.Position, args.WaitTimeout)
		if err == nil {
			*reply = *status
		}
		return err
	})
}

// MasterPosition wraps RPCAgent.
func (tm *TabletManager) MasterPosition(ctx context.Context, args *rpc.Unused, reply *myproto.ReplicationPosition) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionMasterPosition, args, reply, func() error {
		position, err := tm.agent.MasterPosition(ctx)
		if err == nil {
			*reply = position
		}
		return err
	})
}

// ReparentPosition wraps RPCAgent.
func (tm *TabletManager) ReparentPosition(ctx context.Context, args *myproto.ReplicationPosition, reply *actionnode.RestartSlaveData) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionReparentPosition, args, reply, func() error {
		rsd, err := tm.agent.ReparentPosition(ctx, args)
		if err == nil {
			*reply = *rsd
		}
		return err
	})
}

// StopSlave wraps RPCAgent.
func (tm *TabletManager) StopSlave(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionStopSlave, args, reply, true, func() error {
		return tm.agent.StopSlave(ctx)
	})
}

// StopSlaveMinimum wraps RPCAgent.
func (tm *TabletManager) StopSlaveMinimum(ctx context.Context, args *gorpcproto.StopSlaveMinimumArgs, reply *myproto.ReplicationStatus) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionStopSlaveMinimum, args, reply, true, func() error {
		status, err := tm.agent.StopSlaveMinimum(ctx, args.Position, args.WaitTime)
		if err == nil {
			*reply = *status
		}
		return err
	})
}

// StartSlave wraps RPCAgent.
func (tm *TabletManager) StartSlave(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionStartSlave, args, reply, true, func() error {
		return tm.agent.StartSlave(ctx)
	})
}

// TabletExternallyReparented wraps RPCAgent.
func (tm *TabletManager) TabletExternallyReparented(ctx context.Context, args *gorpcproto.TabletExternallyReparentedArgs, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	// TODO(alainjobart) we should forward the RPC deadline from
	// the original gorpc call. Until we support that, use a
	// reasonable hard-coded value.
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionExternallyReparented, args, reply, false, func() error {
		return tm.agent.TabletExternallyReparented(ctx, args.ExternalID)
	})
}

// GetSlaves wraps RPCAgent.
func (tm *TabletManager) GetSlaves(ctx context.Context, args *rpc.Unused, reply *gorpcproto.GetSlavesReply) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionGetSlaves, args, reply, func() error {
		var err error
		reply.Addrs, err = tm.agent.GetSlaves(ctx)
		return err
	})
}

// WaitBlpPosition wraps RPCAgent.
func (tm *TabletManager) WaitBlpPosition(ctx context.Context, args *gorpcproto.WaitBlpPositionArgs, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionWaitBLPPosition, args, reply, true, func() error {
		return tm.agent.WaitBlpPosition(ctx, &args.BlpPosition, args.WaitTimeout)
	})
}

// StopBlp wraps RPCAgent.
func (tm *TabletManager) StopBlp(ctx context.Context, args *rpc.Unused, reply *blproto.BlpPositionList) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionStopBLP, args, reply, true, func() error {
		positions, err := tm.agent.StopBlp(ctx)
		if err == nil {
			*reply = *positions
		}
		return err
	})
}

// StartBlp wraps RPCAgent.
func (tm *TabletManager) StartBlp(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionStartBLP, args, reply, true, func() error {
		return tm.agent.StartBlp(ctx)
	})
}

// RunBlpUntil wraps RPCAgent.
func (tm *TabletManager) RunBlpUntil(ctx context.Context, args *gorpcproto.RunBlpUntilArgs, reply *myproto.ReplicationPosition) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLock(ctx, actionnode.TabletActionRunBLPUntil, args, reply, true, func() error {
		position, err := tm.agent.RunBlpUntil(ctx, args.BlpPositionList, args.WaitTimeout)
		if err == nil {
			*reply = *position
		}
		return err
	})
}

//
// Reparenting related functions
//

// InitMaster wraps RPCAgent.InitMaster
func (tm *TabletManager) InitMaster(ctx context.Context, args *rpc.Unused, reply *myproto.ReplicationPosition) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionInitMaster, args, reply, true, func() error {
		rp, err := tm.agent.InitMaster(ctx)
		if err == nil {
			*reply = rp
		}
		return err
	})
}

// PopulateReparentJournal wraps RPCAgent.PopulateReparentJournal
func (tm *TabletManager) PopulateReparentJournal(ctx context.Context, args *gorpcproto.PopulateReparentJournalArgs, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrap(ctx, actionnode.TabletActionPopulateReparentJournal, args, reply, func() error {
		return tm.agent.PopulateReparentJournal(ctx, args.TimeCreatedNS, args.ActionName, args.MasterAlias, args.ReplicationPosition)
	})
}

// InitSlave wraps RPCAgent.InitSlave
func (tm *TabletManager) InitSlave(ctx context.Context, args *gorpcproto.InitSlaveArgs, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionInitSlave, args, reply, true, func() error {
		if args.WaitTimeout != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, args.WaitTimeout)
			defer cancel()
		}
		return tm.agent.InitSlave(ctx, args.Parent, args.ReplicationPosition, args.TimeCreatedNS)
	})
}

// DemoteMaster wraps RPCAgent.
func (tm *TabletManager) DemoteMaster(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionDemoteMaster, args, reply, true, func() error {
		return tm.agent.DemoteMaster(ctx)
	})
}

// PromoteSlave wraps RPCAgent.
func (tm *TabletManager) PromoteSlave(ctx context.Context, args *rpc.Unused, reply *actionnode.RestartSlaveData) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionPromoteSlave, args, reply, true, func() error {
		rsd, err := tm.agent.PromoteSlave(ctx)
		if err == nil {
			*reply = *rsd
		}
		return err
	})
}

// SlaveWasPromoted wraps RPCAgent.
func (tm *TabletManager) SlaveWasPromoted(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionSlaveWasPromoted, args, reply, true, func() error {
		return tm.agent.SlaveWasPromoted(ctx)
	})
}

// RestartSlave wraps RPCAgent.
func (tm *TabletManager) RestartSlave(ctx context.Context, args *actionnode.RestartSlaveData, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionRestartSlave, args, reply, true, func() error {
		return tm.agent.RestartSlave(ctx, args)
	})
}

// SlaveWasRestarted wraps RPCAgent.
func (tm *TabletManager) SlaveWasRestarted(ctx context.Context, args *actionnode.SlaveWasRestartedArgs, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionSlaveWasRestarted, args, reply, true, func() error {
		return tm.agent.SlaveWasRestarted(ctx, args)
	})
}

// BreakSlaves wraps RPCAgent.
func (tm *TabletManager) BreakSlaves(ctx context.Context, args *rpc.Unused, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionBreakSlaves, args, reply, true, func() error {
		return tm.agent.BreakSlaves(ctx)
	})
}

// backup related methods

// Snapshot wraps RPCAgent.
func (tm *TabletManager) Snapshot(ctx context.Context, args *actionnode.SnapshotArgs, sendReply func(interface{}) error) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionSnapshot, args, nil, true, func() error {
		// create a logger, send the result back to the caller
		logger := logutil.NewChannelLogger(10)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for e := range logger {
				ssr := &gorpcproto.SnapshotStreamingReply{
					Log: &e,
				}
				// Note we don't interrupt the loop here, as
				// we still need to flush and finish the
				// command, even if the channel to the client
				// has been broken. We'll just keep trying to send.
				sendReply(ssr)
			}
			wg.Done()
		}()

		sr, err := tm.agent.Snapshot(ctx, args, logger)
		close(logger)
		wg.Wait()
		if err != nil {
			return err
		}
		ssr := &gorpcproto.SnapshotStreamingReply{
			Result: sr,
		}
		sendReply(ssr)
		return nil
	})
}

// SnapshotSourceEnd wraps RPCAgent.
func (tm *TabletManager) SnapshotSourceEnd(ctx context.Context, args *actionnode.SnapshotSourceEndArgs, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionSnapshotSourceEnd, args, reply, true, func() error {
		return tm.agent.SnapshotSourceEnd(ctx, args)
	})
}

// ReserveForRestore wraps RPCAgent.
func (tm *TabletManager) ReserveForRestore(ctx context.Context, args *actionnode.ReserveForRestoreArgs, reply *rpc.Unused) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionReserveForRestore, args, reply, true, func() error {
		return tm.agent.ReserveForRestore(ctx, args)
	})
}

// Restore wraps RPCAgent.
func (tm *TabletManager) Restore(ctx context.Context, args *actionnode.RestoreArgs, sendReply func(interface{}) error) error {
	ctx = callinfo.RPCWrapCallInfo(ctx)
	return tm.agent.RPCWrapLockAction(ctx, actionnode.TabletActionRestore, args, nil, true, func() error {
		// create a logger, send the result back to the caller
		logger := logutil.NewChannelLogger(10)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for e := range logger {
				// Note we don't interrupt the loop here, as
				// we still need to flush and finish the
				// command, even if the channel to the client
				// has been broken. We'll just keep trying to send.
				sendReply(&e)
			}
			wg.Done()
		}()

		err := tm.agent.Restore(ctx, args, logger)
		close(logger)
		wg.Wait()
		return err
	})
}

// registration glue

func init() {
	tabletmanager.RegisterQueryServices = append(tabletmanager.RegisterQueryServices, func(agent *tabletmanager.ActionAgent) {
		servenv.Register("tabletmanager", &TabletManager{agent})
	})
}

// RegisterForTest will register the RPC, to be used by test instances only
func RegisterForTest(server *rpcplus.Server, agent *tabletmanager.ActionAgent) {
	server.Register(&TabletManager{agent})
}
