/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package testenv supplies test functions for testing vstreamer.
package testenv

import (
	"context"
	"fmt"
	"os"
	"path"

	"gopkg.in/src-d/go-vitess.v1/json2"
	"gopkg.in/src-d/go-vitess.v1/vt/dbconfigs"
	"gopkg.in/src-d/go-vitess.v1/vt/logutil"
	"gopkg.in/src-d/go-vitess.v1/vt/mysqlctl"
	"gopkg.in/src-d/go-vitess.v1/vt/srvtopo"
	"gopkg.in/src-d/go-vitess.v1/vt/topo"
	"gopkg.in/src-d/go-vitess.v1/vt/topo/memorytopo"
	"gopkg.in/src-d/go-vitess.v1/vt/topotools"
	"gopkg.in/src-d/go-vitess.v1/vt/vttablet/tabletserver/connpool"
	"gopkg.in/src-d/go-vitess.v1/vt/vttablet/tabletserver/schema"
	"gopkg.in/src-d/go-vitess.v1/vt/vttablet/tabletserver/tabletenv"
	"gopkg.in/src-d/go-vitess.v1/vt/vttest"

	topodatapb "gopkg.in/src-d/go-vitess.v1/vt/proto/topodata"
	vschemapb "gopkg.in/src-d/go-vitess.v1/vt/proto/vschema"
	vttestpb "gopkg.in/src-d/go-vitess.v1/vt/proto/vttest"
)

// Env contains all the env vars for a test against a mysql instance.
type Env struct {
	cluster *vttest.LocalCluster

	KeyspaceName string
	ShardName    string
	Cells        []string

	TopoServ     *topo.Server
	SrvTopo      srvtopo.Server
	Dbcfgs       *dbconfigs.DBConfigs
	Mysqld       *mysqlctl.Mysqld
	SchemaEngine *schema.Engine
}

type checker struct{}

var _ = connpool.MySQLChecker(checker{})

func (checker) CheckMySQL() {}

// Init initializes an Env.
func Init() (*Env, error) {
	te := &Env{
		KeyspaceName: "vttest",
		ShardName:    "0",
		Cells:        []string{"cell1"},
	}

	ctx := context.Background()
	te.TopoServ = memorytopo.NewServer(te.Cells...)
	if err := te.TopoServ.CreateKeyspace(ctx, te.KeyspaceName, &topodatapb.Keyspace{}); err != nil {
		return nil, err
	}
	if err := te.TopoServ.CreateShard(ctx, te.KeyspaceName, te.ShardName); err != nil {
		panic(err)
	}
	te.SrvTopo = srvtopo.NewResilientServer(te.TopoServ, "TestTopo")

	cfg := vttest.Config{
		Topology: &vttestpb.VTTestTopology{
			Keyspaces: []*vttestpb.Keyspace{
				{
					Name: te.KeyspaceName,
					Shards: []*vttestpb.Shard{
						{
							Name:           "0",
							DbNameOverride: "vttest",
						},
					},
				},
			},
		},
		ExtraMyCnf: []string{path.Join(os.Getenv("VTTOP"), "config/mycnf/rbr.cnf")},
		OnlyMySQL:  true,
	}
	te.cluster = &vttest.LocalCluster{
		Config: cfg,
	}
	if err := te.cluster.Setup(); err != nil {
		os.RemoveAll(te.cluster.Config.SchemaDir)
		return nil, fmt.Errorf("could not launch mysql: %v", err)
	}

	te.Dbcfgs = dbconfigs.NewTestDBConfigs(te.cluster.MySQLConnParams(), te.cluster.MySQLAppDebugConnParams(), te.cluster.DbName())
	te.Mysqld = mysqlctl.NewMysqld(te.Dbcfgs)
	te.SchemaEngine = schema.NewEngine(checker{}, tabletenv.DefaultQsConfig)
	te.SchemaEngine.InitDBConfig(te.Dbcfgs)

	// The first vschema should not be empty. Leads to Node not found error.
	// TODO(sougou): need to fix the bug.
	if err := te.SetVSchema(`{"sharded": true}`); err != nil {
		te.Close()
		return nil, err
	}

	return te, nil
}

// Close tears down TestEnv.
func (te *Env) Close() {
	te.SchemaEngine.Close()
	te.Mysqld.Close()
	te.cluster.TearDown()
	os.RemoveAll(te.cluster.Config.SchemaDir)
}

// SetVSchema sets the vschema for the test keyspace.
func (te *Env) SetVSchema(vs string) error {
	ctx := context.Background()
	logger := logutil.NewConsoleLogger()
	var kspb vschemapb.Keyspace
	if err := json2.Unmarshal([]byte(vs), &kspb); err != nil {
		return err
	}
	if err := te.TopoServ.SaveVSchema(ctx, te.KeyspaceName, &kspb); err != nil {
		return err
	}
	return topotools.RebuildVSchema(ctx, logger, te.TopoServ, te.Cells)
}
