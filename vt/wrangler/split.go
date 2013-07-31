// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wrangler

import (
	"github.com/youtube/vitess/go/relog"
	"github.com/youtube/vitess/go/vt/key"
	tm "github.com/youtube/vitess/go/vt/tabletmanager"
	"github.com/youtube/vitess/go/vt/topo"
)

// replaceError replaces original with recent if recent is not nil,
// logging original if it wasn't nil. This should be used in deferred
// cleanup functions if they change the returned error.
func replaceError(original, recent error) error {
	if recent == nil {
		return original
	}
	if original != nil {
		relog.Error("One of multiple error: %v", original)
	}
	return recent
}

// prepareToSnapshot changes the type of the tablet to backup (when
// the original type is master, it will proceed only if
// forceMasterSnapshot is true). It returns a function that will
// restore the original state.
func (wr *Wrangler) prepareToSnapshot(tabletAlias topo.TabletAlias, forceMasterSnapshot bool) (restoreAfterSnapshot func() error, err error) {
	ti, err := wr.ts.GetTablet(tabletAlias)
	if err != nil {
		return
	}

	originalType := ti.Tablet.Type

	if ti.Tablet.Type == topo.TYPE_MASTER && forceMasterSnapshot {
		// In this case, we don't bother recomputing the serving graph.
		// All queries will have to fail anyway.
		relog.Info("force change type master -> backup: %v", tabletAlias)
		// There is a legitimate reason to force in the case of a single
		// master.
		ti.Tablet.Type = topo.TYPE_BACKUP
		err = topo.UpdateTablet(wr.ts, ti)
	} else {
		err = wr.ChangeType(ti.Alias(), topo.TYPE_BACKUP, false)
	}

	if err != nil {
		return
	}

	restoreAfterSnapshot = func() (err error) {
		relog.Info("change type after snapshot: %v %v", tabletAlias, originalType)

		if ti.Tablet.Parent.Uid == topo.NO_TABLET && forceMasterSnapshot {
			relog.Info("force change type backup -> master: %v", tabletAlias)
			ti.Tablet.Type = topo.TYPE_MASTER
			return topo.UpdateTablet(wr.ts, ti)
		}

		return wr.ChangeType(ti.Alias(), originalType, false)
	}

	return

}

func (wr *Wrangler) RestoreFromMultiSnapshot(dstTabletAlias topo.TabletAlias, sources []topo.TabletAlias, concurrency, fetchConcurrency, insertTableConcurrency, fetchRetryCount int, strategy string) error {
	actionPath, err := wr.ai.RestoreFromMultiSnapshot(dstTabletAlias, &tm.MultiRestoreArgs{
		SrcTabletAliases:       sources,
		Concurrency:            concurrency,
		FetchConcurrency:       fetchConcurrency,
		InsertTableConcurrency: insertTableConcurrency,
		FetchRetryCount:        fetchRetryCount,
		Strategy:               strategy})
	if err != nil {
		return err
	}

	return wr.ai.WaitForCompletion(actionPath, wr.actionTimeout())
}

func (wr *Wrangler) MultiSnapshot(keyRanges []key.KeyRange, tabletAlias topo.TabletAlias, keyName string, concurrency int, tables []string, forceMasterSnapshot, skipSlaveRestart bool, maximumFilesize uint64) (manifests []string, parent topo.TabletAlias, err error) {
	restoreAfterSnapshot, err := wr.prepareToSnapshot(tabletAlias, forceMasterSnapshot)
	if err != nil {
		return
	}
	defer func() {
		err = replaceError(err, restoreAfterSnapshot())
	}()

	actionPath, err := wr.ai.MultiSnapshot(tabletAlias, &tm.MultiSnapshotArgs{KeyName: keyName, KeyRanges: keyRanges, Concurrency: concurrency, Tables: tables, SkipSlaveRestart: skipSlaveRestart, MaximumFilesize: maximumFilesize})
	if err != nil {
		return
	}

	results, err := wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout())
	if err != nil {
		return
	}

	reply := results.(*tm.MultiSnapshotReply)

	return reply.ManifestPaths, reply.ParentAlias, nil
}
