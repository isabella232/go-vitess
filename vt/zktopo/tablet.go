// Copyright 2013, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zktopo

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/youtube/vitess/go/event"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topo/events"
	"github.com/youtube/vitess/go/zk"
	"golang.org/x/net/context"
	"launchpad.net/gozk/zookeeper"

	pb "github.com/youtube/vitess/go/vt/proto/topodata"
)

/*
This file contains the tablet management parts of zktopo.Server
*/

// TabletPathForAlias converts a tablet alias to the zk path
func TabletPathForAlias(alias *pb.TabletAlias) string {
	return fmt.Sprintf("/zk/%v/vt/tablets/%v", alias.Cell, topo.TabletAliasUIDStr(alias))
}

func tabletDirectoryForCell(cell string) string {
	return fmt.Sprintf("/zk/%v/vt/tablets", cell)
}

func tabletFromJSON(data string) (*pb.Tablet, error) {
	t := &pb.Tablet{}
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
func (zkts *Server) CreateTablet(ctx context.Context, tablet *pb.Tablet) error {
	zkTabletPath := TabletPathForAlias(tablet.Alias)

	data, err := json.MarshalIndent(tablet, "  ", "  ")
	if err != nil {
		return err
	}

	// Create /zk/<cell>/vt/tablets/<uid>
	_, err = zk.CreateRecursive(zkts.zconn, zkTabletPath, string(data), 0, zookeeper.WorldACL(zookeeper.PERM_ALL))
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
func (zkts *Server) UpdateTablet(ctx context.Context, tablet *topo.TabletInfo, existingVersion int64) (int64, error) {
	zkTabletPath := TabletPathForAlias(tablet.Alias)
	data, err := json.MarshalIndent(tablet.Tablet, "  ", "  ")
	if err != nil {
		return 0, err
	}

	stat, err := zkts.zconn.Set(zkTabletPath, string(data), int(existingVersion))
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
func (zkts *Server) UpdateTabletFields(ctx context.Context, tabletAlias *pb.TabletAlias, update func(*pb.Tablet) error) error {
	// Store the last tablet value so we can log it if the change succeeds.
	var lastTablet *pb.Tablet

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
		data, err := json.MarshalIndent(tablet, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
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
func (zkts *Server) DeleteTablet(ctx context.Context, alias *pb.TabletAlias) error {
	// We need to find out the keyspace and shard names because
	// those are required in the TabletChange event.
	ti, tiErr := zkts.GetTablet(ctx, alias)

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
			Tablet: pb.Tablet{
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
func (zkts *Server) GetTablet(ctx context.Context, alias *pb.TabletAlias) (*topo.TabletInfo, error) {
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
func (zkts *Server) GetTabletsByCell(ctx context.Context, cell string) ([]*pb.TabletAlias, error) {
	zkTabletsPath := tabletDirectoryForCell(cell)
	children, _, err := zkts.zconn.Children(zkTabletsPath)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			err = topo.ErrNoNode
		}
		return nil, err
	}

	sort.Strings(children)
	result := make([]*pb.TabletAlias, len(children))
	for i, child := range children {
		uid, err := topo.ParseUID(child)
		if err != nil {
			return nil, err
		}
		result[i] = &pb.TabletAlias{
			Cell: cell,
			Uid:  uid,
		}
	}
	return result, nil
}
