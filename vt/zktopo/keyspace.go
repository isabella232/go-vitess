// Copyright 2013, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zktopo

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"

	"github.com/youtube/vitess/go/event"
	"github.com/youtube/vitess/go/jscfg"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topo/events"
	"github.com/youtube/vitess/go/zk"
	"launchpad.net/gozk/zookeeper"
)

/*
This file contains the Keyspace management code for zktopo.Server
*/

const (
	globalKeyspacesPath = "/zk/global/vt/keyspaces"
)

func (zkts *Server) CreateKeyspace(keyspace string, value *topo.Keyspace) error {
	keyspacePath := path.Join(globalKeyspacesPath, keyspace)
	pathList := []string{
		keyspacePath,
		path.Join(keyspacePath, "action"),
		path.Join(keyspacePath, "actionlog"),
		path.Join(keyspacePath, "shards"),
	}
	data := jscfg.ToJson(value)

	alreadyExists := false
	for i, zkPath := range pathList {
		c := ""
		if i == 0 {
			c = data
		}
		_, err := zk.CreateRecursive(zkts.zconn, zkPath, c, 0, zookeeper.WorldACL(zookeeper.PERM_ALL))
		if err != nil {
			if zookeeper.IsError(err, zookeeper.ZNODEEXISTS) {
				alreadyExists = true
			} else {
				return fmt.Errorf("error creating keyspace: %v %v", zkPath, err)
			}
		}
	}
	if alreadyExists {
		return topo.ErrNodeExists
	}

	event.Dispatch(&events.KeyspaceChange{
		Keyspace: keyspace,
		Status:   "created",
		Data:     data,
	})
	return nil
}

func (zkts *Server) UpdateKeyspace(ki *topo.KeyspaceInfo) error {
	keyspacePath := path.Join(globalKeyspacesPath, ki.KeyspaceName())
	data := jscfg.ToJson(ki.Keyspace)
	_, err := zkts.zconn.Set(keyspacePath, data, -1)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			// The code should be:
			//   err = topo.ErrNoNode
			// Temporary code until we have Keyspace object
			// everywhere:
			_, err = zkts.zconn.Create(keyspacePath, data, 0, zookeeper.WorldACL(zookeeper.PERM_ALL))
			if err != nil {
				if zookeeper.IsError(err, zookeeper.ZNONODE) {
					// the directory doesn't even exist
					err = topo.ErrNoNode
				}
				return err
			}

			event.Dispatch(&events.KeyspaceChange{
				Keyspace: ki.KeyspaceName(),
				Status:   "updated (had to create Keyspace object)",
				Data:     data,
			})
			return nil
		}
		return err
	}

	event.Dispatch(&events.KeyspaceChange{
		Keyspace: ki.KeyspaceName(),
		Status:   "updated",
		Data:     data,
	})
	return nil
}

func (zkts *Server) GetKeyspace(keyspace string) (*topo.KeyspaceInfo, error) {
	keyspacePath := path.Join(globalKeyspacesPath, keyspace)
	data, _, err := zkts.zconn.Get(keyspacePath)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			err = topo.ErrNoNode
		}
		return nil, err
	}

	k := &topo.Keyspace{}
	if err = json.Unmarshal([]byte(data), k); err != nil {
		return nil, fmt.Errorf("bad keyspace data %v", err)
	}

	return topo.NewKeyspaceInfo(keyspace, k), nil
}

func (zkts *Server) GetKeyspaces() ([]string, error) {
	children, _, err := zkts.zconn.Children(globalKeyspacesPath)
	if err != nil {
		if zookeeper.IsError(err, zookeeper.ZNONODE) {
			return nil, nil
		}
		return nil, err
	}

	sort.Strings(children)
	return children, nil
}

func (zkts *Server) DeleteKeyspaceShards(keyspace string) error {
	shardsPath := path.Join(globalKeyspacesPath, keyspace, "shards")
	if err := zk.DeleteRecursive(zkts.zconn, shardsPath, -1); err != nil && !zookeeper.IsError(err, zookeeper.ZNONODE) {
		return err
	}

	event.Dispatch(&events.KeyspaceChange{
		Keyspace: keyspace,
		Status:   "deleted all shards",
	})
	return nil
}
