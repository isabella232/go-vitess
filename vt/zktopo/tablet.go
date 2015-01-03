// Copyright 2013, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zktopo

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/youtube/vitess/go/event"
	"github.com/youtube/vitess/go/jscfg"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topo/events"
	"github.com/youtube/vitess/go/zk"
	"launchpad.net/gozk/zookeeper"
)

/*
This file contains the tablet management parts of zktopo.Server
*/

// TabletPathForAlias converts a tablet alias to the zk path
func TabletPathForAlias(alias topo.TabletAlias) string {
	return fmt.Sprintf("/zk/%v/vt/tablets/%v", alias.Cell, alias.TabletUIDStr())
}

func tabletDirectoryForCell(cell string) string {
	return fmt.Sprintf("/zk/%v/vt/tablets", cell)
}

func tabletFromJSON(data string) (*topo.Tablet, error) {
	t := &topo.Tablet{}
	err := json.Unmarshal([]byte(data), t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func tabletInfoFromJSON(data string, version int64) (*topo.TabletInfo, error) {
	tablet, err := tabletFromJSON(data)
	if err != nil {
		return nil, err
	}
	return topo.NewTabletInfo(tablet, version), nil
}

// CreateTablet is part of the topo.Server interface
func (zkts *Server) CreateTablet(tablet *topo.Tablet) error {
	zkTabletPath := TabletPathForAlias(tablet.Alias)

	// Create /zk/<cell>/vt/tablets/<uid>
	_, err := zk.CreateRecursive(zkts.zconn, zkTabletPath, tablet.JSON(), 0, zookeeper.WorldACL(zookeeper.PERM_ALL))
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNODEEXISTS) {
			err = topo.ErrNodeExists
		}
		return err
	}

	event.Dispatch(&events.TabletChange{
		Tablet: *tablet,
		Status: "created",
	})
	return nil
}

// UpdateTablet is part of the topo.Server interface
func (zkts *Server) UpdateTablet(tablet *topo.TabletInfo, existingVersion int64) (int64, error) {
	zkTabletPath := TabletPathForAlias(tablet.Alias)
	stat, err := zkts.zconn.Set(zkTabletPath, tablet.JSON(), int(existingVersion))
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZBADVERSION) {
			err = topo.ErrBadVersion
		} else if zookeeper.IsError(err, zookeeper.ZNONODE) {
			err = topo.ErrNoNode
		}

		return 0, err
	}

	event.Dispatch(&events.TabletChange{
		Tablet: *tablet.Tablet,
		Status: "updated",
	})
	return int64(stat.Version()), nil
}

// UpdateTabletFields is part of the topo.Server interface
func (zkts *Server) UpdateTabletFields(tabletAlias topo.TabletAlias, update func(*topo.Tablet) error) error {
	// Store the last tablet value so we can log it if the change succeeds.
	var lastTablet *topo.Tablet

	zkTabletPath := TabletPathForAlias(tabletAlias)
	f := func(oldValue string, oldStat zk.Stat) (string, error) {
		if oldValue == "" {
			return "", fmt.Errorf("no data for tablet addr update: %v", tabletAlias)
		}

		tablet, err := tabletFromJSON(oldValue)
		if err != nil {
			return "", err
		}
		if err := update(tablet); err != nil {
			return "", err
		}
		lastTablet = tablet
		return jscfg.ToJson(tablet), nil
	}
	err := zkts.zconn.RetryChange(zkTabletPath, 0, zookeeper.WorldACL(zookeeper.PERM_ALL), f)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			err = topo.ErrNoNode
		}
		return err
	}

	if lastTablet != nil {
		event.Dispatch(&events.TabletChange{
			Tablet: *lastTablet,
			Status: "updated",
		})
	}
	return nil
}

// DeleteTablet is part of the topo.Server interface
func (zkts *Server) DeleteTablet(alias topo.TabletAlias) error {
	// We need to find out the keyspace and shard names because those are required
	// in the TabletChange event.
	ti, tiErr := zkts.GetTablet(alias)

	zkTabletPath := TabletPathForAlias(alias)
	err := zk.DeleteRecursive(zkts.zconn, zkTabletPath, -1)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			err = topo.ErrNoNode
		}
		return err
	}

	// Only try to log if we have the required information.
	if tiErr == nil {
		// We only want to copy the identity info for the tablet (alias, etc.).
		// The rest has just been deleted, so it should be blank.
		event.Dispatch(&events.TabletChange{
			Tablet: topo.Tablet{
				Alias:    ti.Tablet.Alias,
				Keyspace: ti.Tablet.Keyspace,
				Shard:    ti.Tablet.Shard,
			},
			Status: "deleted",
		})
	}
	return nil
}

// GetTablet is part of the topo.Server interface
func (zkts *Server) GetTablet(alias topo.TabletAlias) (*topo.TabletInfo, error) {
	zkTabletPath := TabletPathForAlias(alias)
	data, stat, err := zkts.zconn.Get(zkTabletPath)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			err = topo.ErrNoNode
		}
		return nil, err
	}
	return tabletInfoFromJSON(data, int64(stat.Version()))
}

// GetTabletsByCell is part of the topo.Server interface
func (zkts *Server) GetTabletsByCell(cell string) ([]topo.TabletAlias, error) {
	zkTabletsPath := tabletDirectoryForCell(cell)
	children, _, err := zkts.zconn.Children(zkTabletsPath)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			err = topo.ErrNoNode
		}
		return nil, err
	}

	sort.Strings(children)
	result := make([]topo.TabletAlias, len(children))
	for i, child := range children {
		result[i].Cell = cell
		result[i].Uid, err = topo.ParseUID(child)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
