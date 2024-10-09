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

// Package provider providers cosi standard interface
package provider

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_assembleResourceId(t *testing.T) {
	// arrange
	acSecretNameSpace := "as-name-1"
	acSecretName := "as-namespace-1"
	bucketName := "bucket-name-1"
	wantRes := acSecretNameSpace + "/" + acSecretName + "/" + bucketName

	// act
	gotRes := assembleResourceId(acSecretNameSpace, acSecretName, bucketName)

	// assert
	if wantRes != gotRes {
		t.Errorf("Test_assembleResourceId failed, gotRes= [%v], wantRes= [%v]", gotRes, wantRes)
	}
}

func Test_disassembleResourceId(t *testing.T) {
	// arrange
	normalBucketId := "ns/name/id"
	invalidFormatBucketId := "ns/name"
	invalidValueBucketId := "ns//id"

	normalBucketData := &resourceIdInfo{acSecretNameSpace: "ns", acSecretName: "name", resourceName: "id"}
	tests := []struct {
		name         string
		bucketId     string
		bucketIdData *resourceIdInfo
		wantErr      error
	}{
		{"normal-case", normalBucketId, normalBucketData, nil},
		{"invalid-format-case", invalidFormatBucketId, nil, fmt.Errorf("invalid format of input [%s]", invalidFormatBucketId)},
		{"invalid-value-case", invalidValueBucketId, nil, fmt.Errorf("invalid value of input [%s]", invalidValueBucketId)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// act
			gotBucketIdData, gotErr := disassembleResourceId(tt.bucketId)

			// assert
			if gotErr != nil && !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("Test_disassembleResourceId failed, gotErr= [%v], wantErr= [%v]", gotErr, tt.wantErr)
				return
			}

			if (gotErr == nil) && (gotBucketIdData.acSecretNameSpace != tt.bucketIdData.acSecretNameSpace ||
				gotBucketIdData.acSecretName != tt.bucketIdData.acSecretName ||
				gotBucketIdData.resourceName != tt.bucketIdData.resourceName) {
				t.Errorf("Test_disassembleResourceId failed, gotNs= [%v], wantNs= [%v], "+
					"gotName= [%v], wantName= [%v], "+
					"gotBucketName= [%v], wantBucketName= [%v], "+
					"gotErr= [%v], wantErr= [%v]",
					gotBucketIdData.acSecretNameSpace, tt.bucketIdData.acSecretNameSpace,
					gotBucketIdData.acSecretName, tt.bucketIdData.acSecretName,
					gotBucketIdData.resourceName, tt.bucketIdData.resourceName,
					gotErr, tt.wantErr)
				return
			}
		})
	}
}
