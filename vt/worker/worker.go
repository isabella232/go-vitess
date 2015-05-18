// Copyright 2013, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package worker contains the framework, utility methods and core
functions for long running actions. 'vtworker' binary will use these.
*/
package worker

import (
	"html/template"
	"time"

	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/stats"
	"github.com/youtube/vitess/go/vt/topo"
)

// Worker is the base interface for all long running workers.
type Worker interface {
	// StatusAsHTML returns the current worker status in HTML
	StatusAsHTML() template.HTML

	// StatusAsText returns the current worker status in plain text
	StatusAsText() string

	// Run is the main entry point for the worker. It will be
	// called in a go routine.  When the passed in context is
	// cancelled, Run should exit as soon as possible.
	Run(context.Context) error
}

// Resolver is an interface that should be implemented by any workers that need to
// resolve the topology.
type Resolver interface {
	// ResolveDestinationMasters forces the worker to (re)resolve the topology and update
	// the destination masters that it knows about.
	ResolveDestinationMasters(ctx context.Context) error

	// GetDestinationMaster returns the most recently resolved destination master for a particular shard.
	GetDestinationMaster(shardName string) (*topo.TabletInfo, error)
}

// Resolvers should attempt to keep the previous topo resolution cached for at
// least this long.
const resolveTTL = 15 * time.Second

var (
	statsState = stats.NewString("WorkerState")
	// the number of times that the worker attempst to reresolve the masters
	statsDestinationAttemptedResolves = stats.NewInt("WorkerDestinationAttemptedResolves")
	// the number of times that the worker actually hits the topo server, i.e., they don't
	// use a cached topology
	statsDestinationActualResolves = stats.NewInt("WorkerDestinationActualResolves")
	statsRetryCounters             = stats.NewCounters("WorkerRetryCount")
)

// resetVars resets the debug variables that are meant to provide information on a
// per-run basis. This should be called at the beginning of each worker run.
func resetVars() {
	statsState.Set("")
	statsDestinationAttemptedResolves.Set(0)
	statsDestinationActualResolves.Set(0)
	statsRetryCounters.Reset()
}
