// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package actor contains the code for all the actions executed
remotely on a tablet. These actions can be executed as:
- RPCs: called directly from vttablet
- ActionNodes: executed from within vtaction
*/
package actor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"

	log "github.com/golang/glog"
	"github.com/youtube/vitess/go/tb"
	"github.com/youtube/vitess/go/vt/concurrency"
	"github.com/youtube/vitess/go/vt/hook"
	"github.com/youtube/vitess/go/vt/key"
	"github.com/youtube/vitess/go/vt/mysqlctl"
	"github.com/youtube/vitess/go/vt/tabletmanager/actionnode"
	"github.com/youtube/vitess/go/vt/tabletmanager/initiator"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topotools"
)

// The actor applies individual commands to execute an action read
// from a node in topology server. Anything that modifies the state of the
// table should be applied by this code.
//
// The actor signals completion by removing the action node from topology server.
//
// Errors are written to the action node and must (currently) be resolved
// by hand using topo.Server tools.

type TabletActorError string

func (e TabletActorError) Error() string {
	return string(e)
}

// TabletActor is the main object for this package.
type TabletActor struct {
	mysqld      *mysqlctl.Mysqld
	mysqlDaemon mysqlctl.MysqlDaemon
	ts          topo.Server
	tabletAlias topo.TabletAlias
}

// NewTabletActor creates a new TabletActor object.
func NewTabletActor(mysqld *mysqlctl.Mysqld, mysqlDaemon mysqlctl.MysqlDaemon, topoServer topo.Server, tabletAlias topo.TabletAlias) *TabletActor {
	return &TabletActor{mysqld, mysqlDaemon, topoServer, tabletAlias}
}

// This function should be protected from unforseen panics, as
// dispatchAction will catch everything. The rest of the code in this
// function should not panic.
func (ta *TabletActor) HandleAction(actionPath, action, actionGuid string, forceRerun bool) error {
	tabletAlias, data, version, err := ta.ts.ReadTabletActionPath(actionPath)
	ta.tabletAlias = tabletAlias
	actionNode, err := actionnode.ActionNodeFromJson(data, actionPath)
	if err != nil {
		log.Errorf("HandleAction failed unmarshaling %v: %v", actionPath, err)
		return err
	}

	switch actionNode.State {
	case actionnode.ACTION_STATE_RUNNING:
		// see if the process is still running, and if so, wait for it
		proc, _ := os.FindProcess(actionNode.Pid)
		if proc.Signal(syscall.Signal(0)) == syscall.ESRCH {
			// process is dead, either clean up or re-run
			if !forceRerun {
				actionErr := fmt.Errorf("Previous vtaction process died")
				if err := StoreActionResponse(ta.ts, actionNode, actionPath, actionErr); err != nil {
					log.Errorf("Dead process detector failed to update actionNode: %v", err)
					return actionErr
				}
				if err := ta.ts.UnblockTabletAction(actionPath); err != nil {
					log.Errorf("Dead process detector failed unblocking: %v", err)
				}
				return actionErr
			}
		} else {
			log.Warningf("HandleAction waiting for running action: %v", actionPath)
			_, err := initiator.WaitForCompletion(ta.ts, actionPath, 0)
			return err
		}
	case actionnode.ACTION_STATE_FAILED:
		// this happens only in a couple cases:
		// - vtaction was killed by a signal and we caught it
		// - vtaction died unexpectedly, and the next vtaction run detected it
		return fmt.Errorf(actionNode.Error)
	case actionnode.ACTION_STATE_DONE:
		// this is bad
		return fmt.Errorf("Unexpected finished ActionNode in action queue: %v", actionPath)
	}

	// Claim the action by this process.
	actionNode.State = actionnode.ACTION_STATE_RUNNING
	actionNode.Pid = os.Getpid()
	newData := actionNode.ToJson()
	err = ta.ts.UpdateTabletAction(actionPath, newData, version)
	if err != nil {
		if err == topo.ErrBadVersion {
			// The action is schedule by another
			// actor. Most likely the tablet restarted
			// during an action. Just wait for completion.
			log.Warningf("HandleAction waiting for scheduled action: %v", actionPath)
			_, err = initiator.WaitForCompletion(ta.ts, actionPath, 0)
			return err
		} else {
			return err
		}
	}

	// signal handler after we've signed up for the action
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for sig := range c {
			err := StoreActionResponse(ta.ts, actionNode, actionPath, fmt.Errorf("vtaction interrupted by signal: %v", sig))
			if err != nil {
				log.Errorf("Signal handler failed to update actionNode: %v", err)
				os.Exit(-2)
			}
			os.Exit(-1)
		}
	}()

	log.Infof("HandleAction: %v %v", actionPath, data)
	// validate actions, but don't write this back into topo.Server
	if actionNode.Action != action || actionNode.ActionGuid != actionGuid {
		log.Errorf("HandleAction validation failed %v: (%v,%v) (%v,%v)",
			actionPath, actionNode.Action, action, actionNode.ActionGuid, actionGuid)
		return TabletActorError("invalid action initiation: " + action + " " + actionGuid)
	}
	actionErr := ta.dispatchAction(actionNode)
	if err := StoreActionResponse(ta.ts, actionNode, actionPath, actionErr); err != nil {
		return err
	}

	// unblock in topo.Server on completion
	if err := ta.ts.UnblockTabletAction(actionPath); err != nil {
		log.Errorf("HandleAction failed unblocking: %v", err)
		return err
	}
	return actionErr
}

func (ta *TabletActor) dispatchAction(actionNode *actionnode.ActionNode) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = tb.Errorf("dispatchAction panic %v", x)
		}
	}()

	switch actionNode.Action {
	case actionnode.TABLET_ACTION_MULTI_SNAPSHOT:
		err = ta.multiSnapshot(actionNode)
	case actionnode.TABLET_ACTION_MULTI_RESTORE:
		err = ta.multiRestore(actionNode)
	case actionnode.TABLET_ACTION_PING:
		// Just an end-to-end verification that we got the message.
		err = nil
	case actionnode.TABLET_ACTION_RESERVE_FOR_RESTORE:
		err = ta.reserveForRestore(actionNode)
	case actionnode.TABLET_ACTION_RESTORE:
		err = ta.restore(actionNode)
	case actionnode.TABLET_ACTION_SNAPSHOT:
		err = ta.snapshot(actionNode)
	case actionnode.TABLET_ACTION_SNAPSHOT_SOURCE_END:
		err = ta.snapshotSourceEnd(actionNode)

	case actionnode.TABLET_ACTION_EXECUTE_HOOK,
		actionnode.TABLET_ACTION_SET_RDONLY,
		actionnode.TABLET_ACTION_SET_RDWR,
		actionnode.TABLET_ACTION_CHANGE_TYPE,
		actionnode.TABLET_ACTION_SCRAP,
		actionnode.TABLET_ACTION_SLEEP,
		actionnode.TABLET_ACTION_GET_SCHEMA,
		actionnode.TABLET_ACTION_RELOAD_SCHEMA,
		actionnode.TABLET_ACTION_PREFLIGHT_SCHEMA,
		actionnode.TABLET_ACTION_APPLY_SCHEMA,
		actionnode.TABLET_ACTION_GET_PERMISSIONS,
		actionnode.TABLET_ACTION_SLAVE_STATUS,
		actionnode.TABLET_ACTION_WAIT_SLAVE_POSITION,
		actionnode.TABLET_ACTION_MASTER_POSITION,
		actionnode.TABLET_ACTION_REPARENT_POSITION,
		actionnode.TABLET_ACTION_DEMOTE_MASTER,
		actionnode.TABLET_ACTION_PROMOTE_SLAVE,
		actionnode.TABLET_ACTION_SLAVE_WAS_PROMOTED,
		actionnode.TABLET_ACTION_RESTART_SLAVE,
		actionnode.TABLET_ACTION_SLAVE_WAS_RESTARTED,
		actionnode.TABLET_ACTION_BREAK_SLAVES,
		actionnode.TABLET_ACTION_STOP_SLAVE,
		actionnode.TABLET_ACTION_STOP_SLAVE_MINIMUM,
		actionnode.TABLET_ACTION_START_SLAVE,
		actionnode.TABLET_ACTION_EXTERNALLY_REPARENTED,
		actionnode.TABLET_ACTION_GET_SLAVES,
		actionnode.TABLET_ACTION_WAIT_BLP_POSITION,
		actionnode.TABLET_ACTION_STOP_BLP,
		actionnode.TABLET_ACTION_START_BLP,
		actionnode.TABLET_ACTION_RUN_BLP_UNTIL:
		err = TabletActorError("Operation " + actionNode.Action + "  only supported as RPC")
	default:
		err = TabletActorError("invalid action: " + actionNode.Action)
	}

	return
}

// StoreActionResponse writes the result of an action into topology server
func StoreActionResponse(ts topo.Server, actionNode *actionnode.ActionNode, actionPath string, actionErr error) error {
	// change our state
	if actionErr != nil {
		// on failure, set an error field on the node
		actionNode.Error = actionErr.Error()
		actionNode.State = actionnode.ACTION_STATE_FAILED
	} else {
		actionNode.Error = ""
		actionNode.State = actionnode.ACTION_STATE_DONE
	}
	actionNode.Pid = 0

	// Write the data first to our action node, then to the log.
	// In the error case, this node will be left behind to debug.
	data := actionNode.ToJson()
	return ts.StoreTabletActionResponse(actionPath, data)
}

func (ta *TabletActor) hookExtraEnv() map[string]string {
	return map[string]string{"TABLET_ALIAS": ta.tabletAlias.String()}
}

// Operate on a backup tablet. Shutdown mysqld and copy the data files aside.
func (ta *TabletActor) snapshot(actionNode *actionnode.ActionNode) error {
	args := actionNode.Args.(*actionnode.SnapshotArgs)

	tablet, err := ta.ts.GetTablet(ta.tabletAlias)
	if err != nil {
		return err
	}

	if tablet.Type != topo.TYPE_BACKUP {
		return fmt.Errorf("expected backup type, not %v: %v", tablet.Type, ta.tabletAlias)
	}

	filename, slaveStartRequired, readOnly, err := ta.mysqld.CreateSnapshot(tablet.DbName(), tablet.Addr(), false, args.Concurrency, args.ServerMode, ta.hookExtraEnv())
	if err != nil {
		return err
	}

	sr := &actionnode.SnapshotReply{ManifestPath: filename, SlaveStartRequired: slaveStartRequired, ReadOnly: readOnly}
	if tablet.Parent.Uid == topo.NO_TABLET {
		// If this is a master, this will be the new parent.
		// FIXME(msolomon) this doesn't work in hierarchical replication.
		sr.ParentAlias = tablet.Alias
	} else {
		sr.ParentAlias = tablet.Parent
	}
	actionNode.Reply = sr
	return nil
}

func (ta *TabletActor) snapshotSourceEnd(actionNode *actionnode.ActionNode) error {
	args := actionNode.Args.(*actionnode.SnapshotSourceEndArgs)

	tablet, err := ta.ts.GetTablet(ta.tabletAlias)
	if err != nil {
		return err
	}

	if tablet.Type != topo.TYPE_SNAPSHOT_SOURCE {
		return fmt.Errorf("expected snapshot_source type, not %v: %v", tablet.Type, ta.tabletAlias)
	}

	return ta.mysqld.SnapshotSourceEnd(args.SlaveStartRequired, args.ReadOnly, true, ta.hookExtraEnv())
}

// fetch a json file and parses it
func fetchAndParseJsonFile(addr, filename string, result interface{}) error {
	// read the manifest
	murl := "http://" + addr + filename
	resp, err := http.Get(murl)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error fetching url %v: %v", murl, resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	// unpack it
	return json.Unmarshal(data, result)
}

// change a tablet type to RESTORE and set all the other arguments.
// from now on, we can go to:
// - back to IDLE if we don't use the tablet at all (after for instance
//   a successful ReserveForRestore but a failed Snapshot)
// - to SCRAP if something in the process on the target host fails
// - to SPARE if the clone works
func (ta *TabletActor) changeTypeToRestore(tablet, sourceTablet *topo.TabletInfo, parentAlias topo.TabletAlias, keyRange key.KeyRange) error {
	// run the optional preflight_assigned hook
	hk := hook.NewSimpleHook("preflight_assigned")
	topotools.ConfigureTabletHook(hk, ta.tabletAlias)
	if err := hk.ExecuteOptional(); err != nil {
		return err
	}

	// change the type
	tablet.Parent = parentAlias
	tablet.Keyspace = sourceTablet.Keyspace
	tablet.Shard = sourceTablet.Shard
	tablet.Type = topo.TYPE_RESTORE
	tablet.KeyRange = keyRange
	tablet.DbNameOverride = sourceTablet.DbNameOverride
	if err := topo.UpdateTablet(ta.ts, tablet); err != nil {
		return err
	}

	// and create the replication graph items
	return topo.CreateTabletReplicationData(ta.ts, tablet.Tablet)
}

// Reserve a tablet for restore.
// Can be called remotely
func (ta *TabletActor) reserveForRestore(actionNode *actionnode.ActionNode) error {
	// first check mysql, no need to go further if we can't restore
	if err := ta.mysqld.ValidateCloneTarget(ta.hookExtraEnv()); err != nil {
		return err
	}
	args := actionNode.Args.(*actionnode.ReserveForRestoreArgs)

	// read our current tablet, verify its state
	tablet, err := ta.ts.GetTablet(ta.tabletAlias)
	if err != nil {
		return err
	}
	if tablet.Type != topo.TYPE_IDLE {
		return fmt.Errorf("expected idle type, not %v: %v", tablet.Type, ta.tabletAlias)
	}

	// read the source tablet
	sourceTablet, err := ta.ts.GetTablet(args.SrcTabletAlias)
	if err != nil {
		return err
	}

	// find the parent tablet alias we will be using
	var parentAlias topo.TabletAlias
	if sourceTablet.Parent.Uid == topo.NO_TABLET {
		// If this is a master, this will be the new parent.
		// FIXME(msolomon) this doesn't work in hierarchical replication.
		parentAlias = sourceTablet.Alias
	} else {
		parentAlias = sourceTablet.Parent
	}

	return ta.changeTypeToRestore(tablet, sourceTablet, parentAlias, sourceTablet.KeyRange)
}

// Operate on restore tablet.
// Check that the SnapshotManifest is valid and the master has not changed.
// Shutdown mysqld.
// Load the snapshot from source tablet.
// Restart mysqld and replication.
// Put tablet into the replication graph as a spare.
func (ta *TabletActor) restore(actionNode *actionnode.ActionNode) error {
	args := actionNode.Args.(*actionnode.RestoreArgs)

	// read our current tablet, verify its state
	tablet, err := ta.ts.GetTablet(ta.tabletAlias)
	if err != nil {
		return err
	}
	if args.WasReserved {
		if tablet.Type != topo.TYPE_RESTORE {
			return fmt.Errorf("expected restore type, not %v: %v", tablet.Type, ta.tabletAlias)
		}
	} else {
		if tablet.Type != topo.TYPE_IDLE {
			return fmt.Errorf("expected idle type, not %v: %v", tablet.Type, ta.tabletAlias)
		}
	}

	// read the source tablet, compute args.SrcFilePath if default
	sourceTablet, err := ta.ts.GetTablet(args.SrcTabletAlias)
	if err != nil {
		return err
	}
	if strings.ToLower(args.SrcFilePath) == "default" {
		args.SrcFilePath = path.Join(mysqlctl.SnapshotURLPath, mysqlctl.SnapshotManifestFile)
	}

	// read the parent tablet, verify its state
	parentTablet, err := ta.ts.GetTablet(args.ParentAlias)
	if err != nil {
		return err
	}
	if parentTablet.Type != topo.TYPE_MASTER && parentTablet.Type != topo.TYPE_SNAPSHOT_SOURCE {
		return fmt.Errorf("restore expected master or snapshot_source parent: %v %v", parentTablet.Type, args.ParentAlias)
	}

	// read & unpack the manifest
	sm := new(mysqlctl.SnapshotManifest)
	if err := fetchAndParseJsonFile(sourceTablet.Addr(), args.SrcFilePath, sm); err != nil {
		return err
	}

	if !args.WasReserved {
		if err := ta.changeTypeToRestore(tablet, sourceTablet, parentTablet.Alias, sourceTablet.KeyRange); err != nil {
			return err
		}
	}

	// do the work
	if err := ta.mysqld.RestoreFromSnapshot(sm, args.FetchConcurrency, args.FetchRetryCount, args.DontWaitForSlaveStart, ta.hookExtraEnv()); err != nil {
		log.Errorf("RestoreFromSnapshot failed (%v), scrapping", err)
		if err := topotools.Scrap(ta.ts, ta.tabletAlias, false); err != nil {
			log.Errorf("Failed to Scrap after failed RestoreFromSnapshot: %v", err)
		}

		return err
	}

	// change to TYPE_SPARE, we're done!
	return topotools.ChangeType(ta.ts, ta.tabletAlias, topo.TYPE_SPARE, nil, true)
}

func (ta *TabletActor) multiSnapshot(actionNode *actionnode.ActionNode) error {
	args := actionNode.Args.(*actionnode.MultiSnapshotArgs)

	tablet, err := ta.ts.GetTablet(ta.tabletAlias)
	if err != nil {
		return err
	}
	ki, err := ta.ts.GetKeyspace(tablet.Keyspace)
	if err != nil {
		return err
	}

	if tablet.Type != topo.TYPE_BACKUP {
		return fmt.Errorf("expected backup type, not %v: %v", tablet.Type, ta.tabletAlias)
	}

	filenames, err := ta.mysqld.CreateMultiSnapshot(args.KeyRanges, tablet.DbName(), ki.ShardingColumnName, ki.ShardingColumnType, tablet.Addr(), false, args.Concurrency, args.Tables, args.ExcludeTables, args.SkipSlaveRestart, args.MaximumFilesize, ta.hookExtraEnv())
	if err != nil {
		return err
	}

	sr := &actionnode.MultiSnapshotReply{ManifestPaths: filenames}
	if tablet.Parent.Uid == topo.NO_TABLET {
		// If this is a master, this will be the new parent.
		// FIXME(msolomon) this doens't work in hierarchical replication.
		sr.ParentAlias = tablet.Alias
	} else {
		sr.ParentAlias = tablet.Parent
	}
	actionNode.Reply = sr
	return nil
}

func (ta *TabletActor) multiRestore(actionNode *actionnode.ActionNode) (err error) {
	args := actionNode.Args.(*actionnode.MultiRestoreArgs)

	// read our current tablet, verify its state
	// we only support restoring to the master or active replicas
	tablet, err := ta.ts.GetTablet(ta.tabletAlias)
	if err != nil {
		return err
	}
	if tablet.Type != topo.TYPE_MASTER && !topo.IsSlaveType(tablet.Type) {
		return fmt.Errorf("expected master, or slave type, not %v: %v", tablet.Type, ta.tabletAlias)
	}

	// get source tablets addresses
	sourceAddrs := make([]*url.URL, len(args.SrcTabletAliases))
	keyRanges := make([]key.KeyRange, len(args.SrcTabletAliases))
	fromStoragePaths := make([]string, len(args.SrcTabletAliases))
	for i, alias := range args.SrcTabletAliases {
		t, e := ta.ts.GetTablet(alias)
		if e != nil {
			return e
		}
		sourceAddrs[i] = &url.URL{
			Host: t.Addr(),
			Path: "/" + t.DbName(),
		}
		keyRanges[i], e = key.KeyRangesOverlap(tablet.KeyRange, t.KeyRange)
		if e != nil {
			return e
		}
		fromStoragePaths[i] = path.Join(ta.mysqld.SnapshotDir, "from-storage", fmt.Sprintf("from-%v-%v", keyRanges[i].Start.Hex(), keyRanges[i].End.Hex()))
	}

	// change type to restore, no change to replication graph
	originalType := tablet.Type
	tablet.Type = topo.TYPE_RESTORE
	err = topo.UpdateTablet(ta.ts, tablet)
	if err != nil {
		return err
	}

	// first try to get the data from a remote storage
	wg := sync.WaitGroup{}
	rec := concurrency.AllErrorRecorder{}
	for i, alias := range args.SrcTabletAliases {
		wg.Add(1)
		go func(i int, alias topo.TabletAlias) {
			defer wg.Done()
			h := hook.NewSimpleHook("copy_snapshot_from_storage")
			h.ExtraEnv = make(map[string]string)
			for k, v := range ta.hookExtraEnv() {
				h.ExtraEnv[k] = v
			}
			h.ExtraEnv["KEYRANGE"] = fmt.Sprintf("%v-%v", keyRanges[i].Start.Hex(), keyRanges[i].End.Hex())
			h.ExtraEnv["SNAPSHOT_PATH"] = fromStoragePaths[i]
			h.ExtraEnv["SOURCE_TABLET_ALIAS"] = alias.String()
			hr := h.Execute()
			if hr.ExitStatus != hook.HOOK_SUCCESS {
				rec.RecordError(fmt.Errorf("%v hook failed(%v): %v", h.Name, hr.ExitStatus, hr.Stderr))
			}
		}(i, alias)
	}
	wg.Wait()

	// stop replication for slaves, so it doesn't interfere
	if topo.IsSlaveType(originalType) {
		if err := ta.mysqld.StopSlave(map[string]string{"TABLET_ALIAS": tablet.Alias.String()}); err != nil {
			return err
		}
	}

	// run the action, scrap if it fails
	if rec.HasErrors() {
		log.Infof("Got errors trying to get snapshots from storage, trying to get them from original tablets: %v", rec.Error())
		err = ta.mysqld.MultiRestore(tablet.DbName(), keyRanges, sourceAddrs, nil, args.Concurrency, args.FetchConcurrency, args.InsertTableConcurrency, args.FetchRetryCount, args.Strategy)
	} else {
		log.Infof("Got snapshots from storage, reading them from disk directly")
		err = ta.mysqld.MultiRestore(tablet.DbName(), keyRanges, nil, fromStoragePaths, args.Concurrency, args.FetchConcurrency, args.InsertTableConcurrency, args.FetchRetryCount, args.Strategy)
	}
	if err != nil {
		if e := topotools.Scrap(ta.ts, ta.tabletAlias, false); e != nil {
			log.Errorf("Failed to Scrap after failed RestoreFromMultiSnapshot: %v", e)
		}
		return err
	}

	// restart replication
	if topo.IsSlaveType(originalType) {
		if err := ta.mysqld.StartSlave(map[string]string{"TABLET_ALIAS": tablet.Alias.String()}); err != nil {
			return err
		}
	}

	// restore type back
	tablet.Type = originalType
	return topo.UpdateTablet(ta.ts, tablet)
}

// ChecktabletMysqlPort will check the mysql port for the tablet is good,
// and if not will try to update it
func CheckTabletMysqlPort(ts topo.Server, mysqlDaemon mysqlctl.MysqlDaemon, tablet *topo.TabletInfo) *topo.TabletInfo {
	mport, err := mysqlDaemon.GetMysqlPort()
	if err != nil {
		log.Warningf("Cannot get current mysql port, not checking it: %v", err)
		return nil
	}

	if mport == tablet.Portmap["mysql"] {
		return nil
	}

	log.Warningf("MySQL port has changed from %v to %v, updating it in tablet record", tablet.Portmap["mysql"], mport)
	tablet.Portmap["mysql"] = mport
	if err := topo.UpdateTablet(ts, tablet); err != nil {
		log.Warningf("Failed to update tablet record, may use old mysql port")
		return nil
	}

	return tablet
}
