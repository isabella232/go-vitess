// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package backupstorage contains the interface and file system implementation
// of the backup system.
package backupstorage

import (
	"flag"
	"fmt"
	"io"
)

var (
	// BackupStorageImplementation is the implementation to use
	// for BackupStorage. Exported for test purposes.
	BackupStorageImplementation = flag.String("backup_storage_implementation", "", "which implementation to use for the backup storage feature")
)

// BackupHandle describes an individual backup.
type BackupHandle interface {
	// Bucket is the location of the backup. Will contain keyspace/shard.
	Bucket() string

	// Name is the individual name of the backup. Will contain
	// tabletAlias-timestamp.
	Name() string

	// AddFile opens a new file to be added to the backup.
	// Only works for read-write backups (created by StartBackup).
	// filename is guaranteed to only contain alphanumerical
	// characters and hyphens.
	// It should be thread safe, it is possible to call AddFile in
	// multiple go routines once a backup has been started.
	AddFile(filename string) (io.WriteCloser, error)

	// EndBackup stops and closes a backup. The contents should be kept.
	// Only works for read-write backups (created by StartBackup).
	EndBackup() error

	// AbortBackup stops a backup, and removes the contents that
	// have been copied already. It is called if an error occurs
	// while the backup is being taken, and the backup cannot be finished.
	// Only works for read-write backups (created by StartBackup).
	AbortBackup() error

	// ReadFile starts reading a file from a backup.
	// Only works for read-only backups (created by ListBackups).
	ReadFile(filename string) (io.ReadCloser, error)
}

// BackupStorage is the interface to the storage system
type BackupStorage interface {
	// ListBackups returns all the backups in a bucket.  The
	// returned backups are read-only (ReadFile can be called, but
	// AddFile/EndBackup/AbortBackup cannot)
	ListBackups(bucket string) ([]BackupHandle, error)

	// StartBackup creates a new backup with the given name.  If a
	// backup with the same name already exists, it's an error.
	// The returned backup is read-write
	// (AddFile/EndBackup/AbortBackup cann all be called, not
	// ReadFile)
	StartBackup(bucket, name string) (BackupHandle, error)

	// RemoveBackup removes all the data associated with a backup.
	// It will not appear in ListBackups after RemoveBackup succeeds.
	RemoveBackup(bucket, name string) error
}

// BackupStorageMap contains the registered implementations for BackupStorage
var BackupStorageMap = make(map[string]BackupStorage)

// GetBackupStorage returns the current BackupStorage implementation.
// Should be called after flags have been initialized.
func GetBackupStorage() (BackupStorage, error) {
	bs, ok := BackupStorageMap[*BackupStorageImplementation]
	if !ok {
		return nil, fmt.Errorf("no registered implementation of BackupStorage")
	}
	return bs, nil
}
