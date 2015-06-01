// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testlib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/youtube/vitess/go/vt/logutil"
	"github.com/youtube/vitess/go/vt/tabletmanager/tmclient"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/wrangler"
	"github.com/youtube/vitess/go/vt/zktopo"
)

func expvarHandler(gitRev *string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		var vars struct {
			BuildHost      string
			BuildUser      string
			BuildTimestamp int64
			BuildGitRev    string
		}
		vars.BuildHost = "fake host"
		vars.BuildUser = "fake user"
		vars.BuildTimestamp = 123
		vars.BuildGitRev = *gitRev
		result, err := json.Marshal(&vars)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot marshal json: %s", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(result)+"\n")
	}
}

func TestVersion(t *testing.T) {
	// Initialize our environment
	ts := zktopo.NewTestServer(t, []string{"cell1", "cell2"})
	wr := wrangler.New(logutil.NewConsoleLogger(), ts, tmclient.NewTabletManagerClient(), time.Second)
	vp := NewVtctlPipe(t, ts)
	defer vp.Close()

	// couple tablets is enough
	sourceMaster := NewFakeTablet(t, wr, "cell1", 10, topo.TYPE_MASTER,
		TabletKeyspaceShard(t, "source", "0"))
	sourceReplica := NewFakeTablet(t, wr, "cell1", 11, topo.TYPE_REPLICA,
		TabletKeyspaceShard(t, "source", "0"))

	// sourceMaster loop
	sourceMasterGitRev := "fake git rev"
	sourceMaster.StartActionLoop(t, wr)
	sourceMaster.HTTPServer.Handler.(*http.ServeMux).HandleFunc("/debug/vars", expvarHandler(&sourceMasterGitRev))
	defer sourceMaster.StopActionLoop(t)

	// sourceReplica loop
	sourceReplicaGitRev := "fake git rev"
	sourceReplica.StartActionLoop(t, wr)
	sourceReplica.HTTPServer.Handler.(*http.ServeMux).HandleFunc("/debug/vars", expvarHandler(&sourceReplicaGitRev))
	defer sourceReplica.StopActionLoop(t)

	// test when versions are the same
	if err := vp.Run([]string{"ValidateVersionKeyspace", sourceMaster.Tablet.Keyspace}); err != nil {
		t.Fatalf("ValidateVersionKeyspace(same) failed: %v", err)
	}

	// test when versions are different
	sourceReplicaGitRev = "different fake git rev"
	if err := vp.Run([]string{"ValidateVersionKeyspace", sourceMaster.Tablet.Keyspace}); err == nil || !strings.Contains(err.Error(), "is different than slave") {
		t.Fatalf("ValidateVersionKeyspace(different) returned an unexpected error: %v", err)
	}
}
