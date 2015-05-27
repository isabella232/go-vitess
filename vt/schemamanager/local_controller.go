// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schemamanager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/golang/glog"
)

// LocalController listens to the specified schema change dir and applies schema changes.
// schema change dir lay out
//            |
//            |----keyspace_01
//                 |----input
//                      |---- create_test_table.sql
//                      |---- alter_test_table_02.sql
//                      |---- ...
//                 |----complete // contains completed schema changes in yyyy/MM/dd
//                      |----2015
//                           |----01
//                                |----01
//                                     |--- create_table_table_02.sql
//                 |----log // contains detailed execution information about schema changes
//                      |----2015
//                           |----01
//                                |----01
//                                     |--- create_table_table_02.sql
//                 |----error // contains failed schema changes
//                      |----2015
//                           |----01
//                                |----01
//                                     |--- create_table_table_03.sql
// Schema Change Files: ${keyspace}/input/*.sql
// Error Files:         ${keysapce}/error/${YYYY}/${MM}/${DD}/*.sql
// Log Files:           ${keysapce}/log/${YYYY}/${MM}/${DD}/*.sql
// Complete Files:      ${keysapce}/compelte/${YYYY}/${MM}/${DD}/*.sql
type LocalController struct {
	schemaChangeDir string
	keyspace        string
	sqlPath         string
	sqlFilename     string
	errorDir        string
	logDir          string
	completeDir     string
}

// NewLocalController creates a new LocalController instance.
func NewLocalController(schemaChangeDir string) *LocalController {
	return &LocalController{
		schemaChangeDir: schemaChangeDir,
	}
}

// Open goes through the schema change dir and find a keyspace with a pending
// schema change.
func (controller *LocalController) Open() error {
	// find all keyspace directories.
	fileInfos, err := ioutil.ReadDir(controller.schemaChangeDir)
	if err != nil {
		return err
	}
	for _, fileinfo := range fileInfos {
		if !fileinfo.IsDir() {
			continue
		}
		dirpath := path.Join(controller.schemaChangeDir, fileinfo.Name())
		schemaChanges, err := ioutil.ReadDir(path.Join(dirpath, "input"))
		if err != nil {
			log.Warningf("there is no input dir in %s", dirpath)
			continue
		}
		// found a schema change
		if len(schemaChanges) > 0 {
			controller.keyspace = fileinfo.Name()
			controller.sqlFilename = schemaChanges[0].Name()
			controller.sqlPath = path.Join(dirpath, "input", schemaChanges[0].Name())

			currentTime := time.Now()
			datePart := fmt.Sprintf(
				"%d/%d/%d",
				currentTime.Year(),
				currentTime.Month(),
				currentTime.Day())

			controller.errorDir = path.Join(dirpath, "error", datePart)
			controller.completeDir = path.Join(dirpath, "complete", datePart)
			controller.logDir = path.Join(dirpath, "log", datePart)
			// the remaining schema changes will be picked by the next runs
			break
		}
	}
	return nil
}

// Read reads schema changes.
func (controller *LocalController) Read() ([]string, error) {
	if controller.keyspace == "" || controller.sqlPath == "" {
		return []string{}, nil
	}
	data, err := ioutil.ReadFile(controller.sqlPath)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), ";"), nil
}

// Keyspace returns current keyspace that is ready for applying schema change.
func (controller *LocalController) Keyspace() string {
	return controller.keyspace
}

// Close reset keyspace, sqlPath, errorDir, logDir and completeDir.
func (controller *LocalController) Close() {
	controller.keyspace = ""
	controller.sqlPath = ""
	controller.sqlFilename = ""
	controller.errorDir = ""
	controller.logDir = ""
	controller.completeDir = ""
}

// OnReadSuccess is no-op
func (controller *LocalController) OnReadSuccess() error {
	return nil
}

// OnReadFail is no-op
func (controller *LocalController) OnReadFail(err error) error {
	log.Errorf("failed to read file: %s, error: %v", controller.sqlPath, err)
	return nil
}

// OnValidationSuccess is no-op
func (controller *LocalController) OnValidationSuccess() error {
	return nil
}

// OnValidationFail is no-op
func (controller *LocalController) OnValidationFail(err error) error {
	return controller.moveToErrorDir()
}

// OnExecutorComplete is no-op
func (controller *LocalController) OnExecutorComplete(result *ExecuteResult) error {
	if len(result.FailedShards) > 0 || result.ExecutorErr != "" {
		return controller.moveToErrorDir()
	}
	if err := os.MkdirAll(controller.completeDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(controller.logDir, os.ModePerm); err != nil {
		return err
	}

	if err := controller.writeToLogDir(result); err != nil {
		return err
	}

	return os.Rename(
		controller.sqlPath,
		path.Join(controller.completeDir, controller.sqlFilename))
}

func (controller *LocalController) moveToErrorDir() error {
	if err := os.MkdirAll(controller.errorDir, os.ModePerm); err != nil {
		return err
	}
	return os.Rename(
		controller.sqlPath,
		path.Join(controller.errorDir, controller.sqlFilename))
}

func (controller *LocalController) writeToLogDir(result *ExecuteResult) error {
	logFile, err := os.Create(path.Join(controller.logDir, controller.sqlFilename))
	if err != nil {
		return err
	}
	defer logFile.Close()

	logFile.WriteString(fmt.Sprintf("-- new file: %s\n", controller.sqlPath))
	for _, sql := range result.Sqls {
		logFile.WriteString(sql)
		logFile.WriteString(";\n")
	}
	rowsReturned := uint64(0)
	rowsAffected := uint64(0)
	for _, queryResult := range result.SuccessShards {
		rowsReturned += uint64(len(queryResult.Result.Rows))
		rowsAffected += queryResult.Result.RowsAffected
	}
	logFile.WriteString(fmt.Sprintf("-- Rows returned: %d\n", rowsReturned))
	logFile.WriteString(fmt.Sprintf("-- Rows affected: %d\n", rowsAffected))
	logFile.WriteString("-- \n")
	logFile.WriteString(fmt.Sprintf("-- ran in %fs\n", result.TotalTimeSpent.Seconds()))
	logFile.WriteString("-- Execution succeeded\n")
	return nil
}

var _ Controller = (*LocalController)(nil)

func init() {
	RegisterControllerFactory(
		"local",
		func(params map[string]string) (Controller, error) {
			schemaChangeDir, ok := params[SchemaChangeDirName]
			if !ok {
				return nil, fmt.Errorf("unable to construct a LocalController instance because param: %s is missing in params: %v", SchemaChangeDirName, params)
			}
			return NewLocalController(schemaChangeDir), nil
		},
	)
}
