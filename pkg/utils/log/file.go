/*
 Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

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

// Package log output logged entries to respective logging hooks
package log

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	logFilePermission = 0640
	defaultFileSize   = 1024 * 1024 * 20 // 20M

	backupTimeFormat  = "20060102-150405"
	defaultMaxBackups = 9

	logFileRootDirPermission = 0750
	rotatedLogFilePermission = 0440

	decimalBase  = 10
	int64BitSize = 64
)

var (
	logFileSize = flag.String("log-file-size",
		strconv.Itoa(defaultFileSize),
		"Maximum file size before log rotation")
	maxBackups = flag.Uint("max-backups",
		defaultMaxBackups,
		"maximum number of backup log file")
)

// FileHook sends log entries to a file.
type FileHook struct {
	logFileHandle        *fileHandler
	logRotationThreshold int64
	formatter            logrus.Formatter
	logRotateMutex       *sync.Mutex
}

// ensure interface implementation
var _ flushable = &FileHook{}
var _ closable = &FileHook{}

// newFileHook creates a new log hook for writing to a file.
func newFileHook(logFilePath string, logFormat logrus.Formatter) (*FileHook, error) {
	logFileRootDir := filepath.Dir(logFilePath)
	dir, err := os.Lstat(logFileRootDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(logFileRootDir, logFileRootDirPermission); err != nil {
			return nil, fmt.Errorf("could not create log directory %v. %v", logFileRootDir, err)
		}
	}
	if dir != nil && !dir.IsDir() {
		return nil, fmt.Errorf("log path %v exists and is not a directory, please remove it", logFileRootDir)
	}

	filesizeThreshold, err := getNumInByte()
	if err != nil {
		return nil, fmt.Errorf("error in evaluating max log file size: %v. Check 'logFileSize' flag", err)
	}

	return &FileHook{
		logRotationThreshold: filesizeThreshold,
		formatter:            logFormat,
		logFileHandle:        newFileHandler(logFilePath),
		logRotateMutex:       &sync.Mutex{}}, nil
}

// Close file handler
func (hook *FileHook) close() {
	// All writes are synced and no file descriptor are left to close with current implementation
}

// Flush commits the current contents of the file
func (hook *FileHook) flush() {
	// All writes are synced and no file descriptor are left to close with current implementation
}

// Levels returns all supported levels
func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire ensure logging of respective log entries
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	// Get formatted entry
	lineBytes, err := hook.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("could not read log entry. %v", err)
	}

	// Write log entry to file
	_, err = hook.logFileHandle.writeString(string(lineBytes))
	if err != nil {
		// let logrus print error message
		return fmt.Errorf("write log message [%s] error. %v", lineBytes, err)
	}

	// Rotate the file as needed
	if err = hook.maybeDoLogfileRotation(); err != nil {
		return err
	}

	return nil
}

// logfileNeedsRotation checks to see if a file has grown too large
func (hook *FileHook) fileNeedsRotation() bool {
	fileInfo, err := hook.logFileHandle.stat()
	if err != nil {
		return false
	}

	return fileInfo.Size() >= hook.logRotationThreshold
}

// maybeDoLogfileRotation check and perform log rotation
func (hook *FileHook) maybeDoLogfileRotation() error {
	if hook.fileNeedsRotation() {
		hook.logRotateMutex.Lock()
		defer hook.logRotateMutex.Unlock()

		if hook.fileNeedsRotation() {
			// Do the rotation.
			err := hook.logFileHandle.rotate()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type fileHandler struct {
	rwLock   *sync.RWMutex
	filePath string
}

func newFileHandler(logFilePath string) *fileHandler {
	return &fileHandler{
		filePath: logFilePath,
	}
}

func (f *fileHandler) stat() (os.FileInfo, error) {
	return os.Stat(f.filePath)
}

func (f *fileHandler) writeString(s string) (int, error) {
	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, logFilePermission)
	if err != nil {
		return 0, fmt.Errorf("failed to open log file with error [%v]", err)
	}
	defer file.Close()
	return file.WriteString(s)
}

func (f *fileHandler) rotate() error {
	// Do the rotation.
	rotatedLogFileLocation := f.filePath + time.Now().Format(backupTimeFormat)
	if err := os.Rename(f.filePath, rotatedLogFileLocation); err != nil {
		return fmt.Errorf("failed to create backup file. %v", err)
	}
	if err := os.Chmod(rotatedLogFileLocation, rotatedLogFilePermission); err != nil {
		return fmt.Errorf("failed to chmod backup file. %s", err)
	}
	// try to remove old backup files
	backupFiles, err := f.sortedBackupLogFiles()
	if err != nil {
		return err
	}

	if *maxBackups < uint(len(backupFiles)) {
		oldBackupFiles := backupFiles[*maxBackups:]

		for _, file := range oldBackupFiles {
			err := os.Remove(filepath.Join(filepath.Dir(f.filePath), file.Name()))
			if err != nil {
				return fmt.Errorf("failed to remove old backup file [%s]. %v", file.Name(), err)
			}
		}
	}
	return nil
}

type logFileInfo struct {
	timestamp time.Time
	os.FileInfo
}

func (f *fileHandler) sortedBackupLogFiles() ([]logFileInfo, error) {
	files, err := ioutil.ReadDir(filepath.Dir(f.filePath))
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %v", err)
	}

	logFiles := make([]logFileInfo, 0)
	baseLogFileName := filepath.Base(f.filePath)

	// take out log files from directory
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// ignore files other than log file and current log file itself
		fileName := f.Name()
		if !strings.HasPrefix(fileName, baseLogFileName) || fileName == baseLogFileName {
			continue
		}

		timestamp, err := time.Parse(backupTimeFormat, fileName[len(baseLogFileName):])
		if err != nil {
			logrus.Warningf("Failed parsing log file suffix timestamp. %v", err)
			continue
		}

		logFiles = append(logFiles, logFileInfo{timestamp: timestamp, FileInfo: f})
	}

	sort.Sort(byTimeFormat(logFiles))

	return logFiles, nil
}

type byTimeFormat []logFileInfo

func (by byTimeFormat) isOutBounds(i, j int) bool {
	return i >= len(by) || j >= len(by)
}

func (by byTimeFormat) Less(i, j int) bool {
	if by.isOutBounds(i, j) {
		return false
	}

	return by[i].timestamp.After(by[j].timestamp)
}

func (by byTimeFormat) Swap(i, j int) {
	if by.isOutBounds(i, j) {
		return
	}

	by[i], by[j] = by[j], by[i]
}

func (by byTimeFormat) Len() int {
	return len(by)
}

func getNumInByte() (int64, error) {
	var sum int64 = 0
	var err error

	maxDataNum := strings.ToUpper(*logFileSize)
	lastLetter := maxDataNum[len(maxDataNum)-1:]

	if lastLetter >= "0" && lastLetter <= "9" {
		sum, err = strconv.ParseInt(maxDataNum, decimalBase, int64BitSize)
		if err != nil {
			return 0, err
		}
	} else {
		sum, err = strconv.ParseInt(maxDataNum[:len(maxDataNum)-1], decimalBase, int64BitSize)
		if err != nil {
			return 0, err
		}

		if lastLetter == "M" {
			sum *= 1024 * 1024
		} else if lastLetter == "K" {
			sum *= 1024
		}
	}

	return sum, nil
}
