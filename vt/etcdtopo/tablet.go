// Copyright 2014, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etcdtopo

import (
	"encoding/json"
	"fmt"

	"github.com/youtube/vitess/go/event"
	"github.com/youtube/vitess/go/jscfg"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topo/events"
	"golang.org/x/net/context"

	pb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// CreateTablet implements topo.Server.
func (s *Server) CreateTablet(ctx context.Context, tablet *pb.Tablet) error {
	cell, err := s.getCell(tablet.Alias.Cell)
	if err != nil {
		return err
	}

	data := jscfg.ToJSON(tablet)
	_, err = cell.Create(tabletFilePath(tablet.Alias), data, 0 /* ttl */)
	if err != nil {
		return convertError(err)
	}

	event.Dispatch(&events.TabletChange{
		Tablet: *tablet,
		Status: "created",
	})
	return nil
}

// UpdateTablet implements topo.Server.
func (s *Server) UpdateTablet(ctx context.Context, ti *topo.TabletInfo, existingVersion int64) (int64, error) {
	cell, err := s.getCell(ti.Alias.Cell)
	if err != nil {
		return -1, err
	}

	data := jscfg.ToJSON(ti.Tablet)
	resp, err := cell.CompareAndSwap(tabletFilePath(ti.Alias),
		data, 0 /* ttl */, "" /* prevValue */, uint64(existingVersion))
	if err != nil {
		return -1, convertError(err)
	}
	if resp.Node == nil {
		return -1, ErrBadResponse
	}

	event.Dispatch(&events.TabletChange{
		Tablet: *ti.Tablet,
		Status: "updated",
	})
	return int64(resp.Node.ModifiedIndex), nil
}

// UpdateTabletFields implements topo.Server.
func (s *Server) UpdateTabletFields(ctx context.Context, tabletAlias *pb.TabletAlias, updateFunc func(*pb.Tablet) error) error {
	var ti *topo.TabletInfo
	var err error

	for {
		if ti, err = s.GetTablet(ctx, tabletAlias); err != nil {
			return err
		}
		if err = updateFunc(ti.Tablet); err != nil {
			return err
		}
		if _, err = s.UpdateTablet(ctx, ti, ti.Version()); err != topo.ErrBadVersion {
			break
		}
	}
	if err != nil {
		return err
	}

	event.Dispatch(&events.TabletChange{
		Tablet: *ti.Tablet,
		Status: "updated",
	})
	return nil
}

// DeleteTablet implements topo.Server.
func (s *Server) DeleteTablet(ctx context.Context, tabletAlias *pb.TabletAlias) error {
	cell, err := s.getCell(tabletAlias.Cell)
	if err != nil {
		return err
	}

	// Get the keyspace and shard names for the TabletChange event.
	ti, tiErr := s.GetTablet(ctx, tabletAlias)

	_, err = cell.Delete(tabletDirPath(tabletAlias), true /* recursive */)
	if err != nil {
		return convertError(err)
	}

	// Only try to log if we have the required info.
	if tiErr == nil {
		// Only copy the identity info for the tablet. The rest has been deleted.
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

// ValidateTablet implements topo.Server.
func (s *Server) ValidateTablet(ctx context.Context, tabletAlias *pb.TabletAlias) error {
	_, err := s.GetTablet(ctx, tabletAlias)
	return err
}

// GetTablet implements topo.Server.
func (s *Server) GetTablet(ctx context.Context, tabletAlias *pb.TabletAlias) (*topo.TabletInfo, error) {
	cell, err := s.getCell(tabletAlias.Cell)
	if err != nil {
		return nil, err
	}

	resp, err := cell.Get(tabletFilePath(tabletAlias), false /* sort */, false /* recursive */)
	if err != nil {
		return nil, convertError(err)
	}
	if resp.Node == nil {
		return nil, ErrBadResponse
	}

	value := &pb.Tablet{}
	if err := json.Unmarshal([]byte(resp.Node.Value), value); err != nil {
		return nil, fmt.Errorf("bad tablet data (%v): %q", err, resp.Node.Value)
	}

	return topo.NewTabletInfo(value, int64(resp.Node.ModifiedIndex)), nil
}

// GetTabletsByCell implements topo.Server.
func (s *Server) GetTabletsByCell(ctx context.Context, cellName string) ([]*pb.TabletAlias, error) {
	cell, err := s.getCell(cellName)
	if err != nil {
		return nil, err
	}

	resp, err := cell.Get(tabletsDirPath, false /* sort */, false /* recursive */)
	if err != nil {
		return nil, convertError(err)
	}

	nodes, err := getNodeNames(resp)
	if err != nil {
		return nil, err
	}

	tablets := make([]*pb.TabletAlias, 0, len(nodes))
	for _, node := range nodes {
		tabletAlias, err := topo.ParseTabletAliasString(node)
		if err != nil {
			return nil, err
		}
		tablets = append(tablets, tabletAlias)
	}
	return tablets, nil
}
