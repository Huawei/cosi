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
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
)

type fakeFileInfo struct {
	dir      bool
	basename string
	modtime  time.Time
	ents     []*fakeFileInfo
	contents string
	err      error
}

func (f *fakeFileInfo) Name() string       { return f.basename }
func (f *fakeFileInfo) Sys() any           { return nil }
func (f *fakeFileInfo) ModTime() time.Time { return f.modtime }
func (f *fakeFileInfo) IsDir() bool        { return f.dir }
func (f *fakeFileInfo) Size() int64        { return int64(len(f.contents)) }
func (f *fakeFileInfo) Mode() fs.FileMode  { return 0644 }
func (f *fakeFileInfo) String() string     { return fs.FormatFileInfo(f) }

func Test_FileHandler_SortedBackupLogFiles_Success(t *testing.T) {
	// arrange
	fh := fileHandler{filePath: "/var/log/huawei-cosi/cosi-driver/cosi-driver", rwLock: &sync.RWMutex{}}
	fakeFileOne := &fakeFileInfo{basename: "cosi-driver", dir: false}
	fakeFileTwo := &fakeFileInfo{basename: "cosi-driver20241112-082324", dir: false}
	fakeFileThree := &fakeFileInfo{basename: "cosi-driver20241110-082324", dir: false}
	fakeFileFour := &fakeFileInfo{basename: "cosi-driver-wrong-time", dir: false}
	fakeFileDir := &fakeFileInfo{basename: "/var/log/cosi/fake-dir", dir: true}
	fakeFiles := []fs.FileInfo{fakeFileOne, fakeFileTwo, fakeFileThree, fakeFileFour, fakeFileDir}
	wantLogFiles := []logFileInfo{
		{FileInfo: fakeFileTwo},
		{FileInfo: fakeFileThree},
	}

	// mock
	patches := gomonkey.ApplyFuncReturn(ioutil.ReadDir, fakeFiles, nil)

	// act
	gotLogFiles, gotErr := fh.sortedBackupLogFiles()

	// assert
	if len(gotLogFiles) != 2 || gotLogFiles[0].Name() != wantLogFiles[0].Name() ||
		gotLogFiles[1].Name() != wantLogFiles[1].Name() {
		t.Errorf("Test_fileHandler_sortedBackupLogFiles failed, "+
			"wantLogFiles= [%v], gotLogFiles= [%v]", wantLogFiles, gotLogFiles)
		return
	}

	if gotErr != nil {
		t.Errorf("Test_fileHandler_sortedBackupLogFiles failed, wantErr= nil, gotErr= [%v]", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_FileHandler_Rotate_Success(t *testing.T) {
	// arrange
	fh := fileHandler{filePath: "/var/log/huawei-cosi/cosi-driver/cosi-driver", rwLock: &sync.RWMutex{}}
	fakeFileTwo := &fakeFileInfo{basename: "cosi-driver20241112-082324", dir: false}

	var wantLogFiles []logFileInfo
	for i := 0; i < int(*maxBackups)+1; i++ {
		wantLogFiles = append(wantLogFiles, logFileInfo{FileInfo: fakeFileTwo})
	}

	// mock
	patches := gomonkey.ApplyFuncReturn(os.Rename, nil).
		ApplyFuncReturn(os.Chmod, nil).
		ApplyFuncReturn(filepath.Join, "path-demo").
		ApplyFunc((*fileHandler).sortedBackupLogFiles, func(_ *fileHandler) ([]logFileInfo, error) {
			return wantLogFiles, nil
		}).
		ApplyFuncReturn(os.Remove, nil)

	// act
	gotErr := fh.rotate()

	// assert
	if gotErr != nil {
		t.Errorf("Test_FileHandler_Rotate_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_GetNumInByte(t *testing.T) {
	// arrange
	var testData = []struct {
		input    string
		expected int64
		err      error
	}{
		{"100", 100, nil},
		{"100M", 104857600, nil},
		{"100K", 102400, nil},
		{"100.5", 0, strconv.ErrSyntax},
	}

	for _, data := range testData {
		// act
		logFileSize = &data.input
		result, gotErr := getNumInByte()

		// assert
		if result != data.expected || (gotErr != nil && data.err == nil) || (gotErr == nil && data.err != nil) {
			t.Errorf("Test_GetNumInByte failed for input= [%s], expected= [%d], got= [%d], "+
				"wantErr= [%v], gotErr= [%v]", data.input, data.expected, result, data.err, gotErr)
		}
	}
}
