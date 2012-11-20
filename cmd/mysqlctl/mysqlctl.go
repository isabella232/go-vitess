// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/vitess/go/relog"
	"code.google.com/p/vitess/go/vt/dbconfigs"
	"code.google.com/p/vitess/go/vt/key"
	"code.google.com/p/vitess/go/vt/mysqlctl"
	"flag"
	"fmt"
	"log"
	"os"
)

var port = flag.Int("port", 6612, "vtocc port")
var force = flag.Bool("force", false, "force action")
var mysqlPort = flag.Int("mysql-port", 3306, "mysql port")
var tabletUid = flag.Uint("tablet-uid", 41983, "tablet uid")
var logLevel = flag.String("log.level", "WARNING", "set log level")
var tabletAddr string

func initCmd(mysqld *mysqlctl.Mysqld, args []string) {
	if err := mysqlctl.Init(mysqld); err != nil {
		relog.Fatal("failed init mysql: %v", err)
	}
}

func partialRestoreCmd(mysqld *mysqlctl.Mysqld, args []string) {
	if len(args) != 1 {
		relog.Fatal("Command partialrestore requires <split snapshot manifest file>")
	}
	rs, err := mysqlctl.ReadSplitSnapshotManifest(args[0])
	if err == nil {
		err = mysqld.RestoreFromPartialSnapshot(rs)
	}
	if err != nil {
		relog.Fatal("partialrestore failed: %v", err)
	}
}

func partialSnapshotCmd(mysqld *mysqlctl.Mysqld, args []string) {
	subFlags := flag.NewFlagSet("partialsnapshot", flag.ExitOnError)
	start := subFlags.String("start", "", "start of the key range")
	end := subFlags.String("end", "", "end of the key range")
	subFlags.Parse(args)

	if len(subFlags.Args()) != 2 {
		relog.Fatal("action partialsnapshot requires <db name> <key name>")
	}

	filename, err := mysqld.CreateSplitSnapshot(subFlags.Arg(0), subFlags.Arg(1), key.HexKeyspaceId(*start), key.HexKeyspaceId(*end), tabletAddr, false)
	if err != nil {
		relog.Fatal("partialsnapshot failed: %v", err)
	} else {
		relog.Info("manifest location: %v", filename)
	}
}

func restoreCmd(mysqld *mysqlctl.Mysqld, args []string) {
	if len(args) != 1 {
		relog.Fatal("Command restore requires <snapshot manifest file>")
	}
	rs, err := mysqlctl.ReadSnapshotManifest(args[0])
	if err == nil {
		err = mysqld.RestoreFromSnapshot(rs)
	}
	if err != nil {
		relog.Fatal("restore failed: %v", err)
	}
}

func shutdownCmd(mysqld *mysqlctl.Mysqld, args []string) {
	if mysqlErr := mysqlctl.Shutdown(mysqld, true); mysqlErr != nil {
		relog.Fatal("failed shutdown mysql: %v", mysqlErr)
	}
}

func snapshotCmd(mysqld *mysqlctl.Mysqld, args []string) {
	if len(args) != 1 {
		relog.Fatal("Command snapshot requires <db name>")
	}
	filename, err := mysqld.CreateSnapshot(args[0], tabletAddr, false)
	if err != nil {
		relog.Fatal("snapshot failed: %v", err)
	} else {
		relog.Info("manifest location: %v", filename)
	}
}

func startCmd(mysqld *mysqlctl.Mysqld, args []string) {
	if err := mysqlctl.Start(mysqld); err != nil {
		relog.Fatal("failed start mysql: %v", err)
	}
}

func teardownCmd(mysqld *mysqlctl.Mysqld, args []string) {
	if err := mysqlctl.Teardown(mysqld, *force); err != nil {
		relog.Fatal("failed teardown mysql (forced? %v): %v", *force, err)
	}
}

type command struct {
	name   string
	method func(*mysqlctl.Mysqld, []string)
	params string
	help   string
}

var commands = []command{
	command{"init", initCmd, "",
		"Initalizes the directory structure and starts mysqld"},
	command{"teardown", teardownCmd, "",
		"Shuts mysqld down, and removes the directory"},

	command{"start", startCmd, "",
		"Starts mysqld on an already 'init'-ed directory"},
	command{"shutdown", shutdownCmd, "",
		"Shuts down mysqld, does not remove any file"},

	command{"snapshot", snapshotCmd,
		"<db name>",
		"Takes a full snapshot, copying the innodb data files"},
	command{"restore", restoreCmd,
		"<snapshot manifest file>",
		"Restores a full snapshot"},

	command{"partialsnapshot", partialSnapshotCmd,
		"[--start=<start key>] [--stop=<stop key>] <db name> <key name>",
		"Takes a partial snapshot using 'select * into' commands"},
	command{"partialrestore", partialRestoreCmd,
		"<split snapshot manifest file>",
		"Restores a database from a partial snapshot"},
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [global parameters] command [command parameters]\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "\nThe global optional parameters are:\n")
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\nThe commands are:\n")
		for _, cmd := range commands {
			fmt.Fprintf(os.Stderr, "  %s", cmd.name)
			if cmd.params != "" {
				fmt.Fprintf(os.Stderr, " %s", cmd.params)
			}
			fmt.Fprintf(os.Stderr, "\n    %s\n", cmd.help)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
	flag.Parse()
	logger := relog.New(os.Stderr, "",
		log.Ldate|log.Lmicroseconds|log.Lshortfile,
		relog.LogNameToLogLevel(*logLevel))
	relog.SetLogger(logger)

	tabletAddr = fmt.Sprintf("%v:%v", "localhost", *port)
	mycnf := mysqlctl.NewMycnf(uint32(*tabletUid), *mysqlPort, mysqlctl.VtReplParams{})
	dbcfgs, err := dbconfigs.Init(mycnf.SocketFile)
	if err != nil {
		relog.Fatal("%v", err)
	}
	mysqld := mysqlctl.NewMysqld(mycnf, dbcfgs.Dba, dbcfgs.Repl)

	action := flag.Arg(0)
	for _, cmd := range commands {
		if cmd.name == action {
			cmd.method(mysqld, flag.Args()[1:])
			return
		}
	}
	relog.Fatal("invalid action: %v", action)
}
