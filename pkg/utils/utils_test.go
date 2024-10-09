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

// Package utils provides a lot of utility function
package utils

import (
	"testing"
)

func Test_GetSortedUrlQueryString(t *testing.T) {
	// arrange
	normalMap := make(map[string]string)
	normalMap["Action"] = "3"
	normalMap["AccountName"] = "2"
	normalMap["AccountId"] = "1"
	normalMap["AWSAccessKeyId"] = "4"
	normalMap["SignatureVersion"] = "7"
	normalMap["SignatureMethod"] = "6"
	normalMap["Timestamp"] = "2006-01-02T15:04:05.000Z"
	normalMap["Email"] = "myEmail@test.com"
	want := "AWSAccessKeyId=4&AccountId=1&AccountName=2&Action=3&Email=myEmail%40test.com&" +
		"SignatureMethod=6&SignatureVersion=7&Timestamp=2006-01-02T15%3A04%3A05.000Z"

	// act
	got := GetSortedUrlQueryString(normalMap)

	// assert
	if got != want {
		t.Errorf("Test_GetSortedUrlQueryString failed, got= [%s], want= [%s]", got, want)
	}
}

func Test_ContainsElement(t *testing.T) {
	// arrange
	type args struct {
		elements []string
		target   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil-cases",
			args{elements: nil, target: ""},
			false,
		},
		{"false-cases",
			args{elements: []string{"finalizers-1", "finalizers-2"}, target: "finalizers-3"},
			false,
		},
		{"true-cases",
			args{elements: []string{"finalizers-1", "finalizers-2"}, target: "finalizers-1"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// act
			got := ContainsElement(tt.args.elements, tt.args.target)

			// assert
			if got != tt.want {
				t.Errorf("Test_ContainsElement failed, got= [%v], want= [%v]", got, tt.want)
			}
		})
	}
}

func Test_BuildTLSConfig_WithRootCA(t *testing.T) {
	// arrange
	rootCA := []byte("123")

	// act
	tlsConfig, gotErr := BuildTLSConfig(rootCA)

	// assert
	if gotErr != nil || tlsConfig.InsecureSkipVerify == true || tlsConfig.RootCAs == nil {
		t.Errorf("Test_BuildTLSConfig_WithRootCA failed, gotErr= [%v], gotConfig= [%v]", gotErr, tlsConfig)
	}
}

func Test_BuildTLSConfig_Empty_RootCA(t *testing.T) {
	// arrange
	rootCA := []byte("")

	// act
	tlsConfig, gotErr := BuildTLSConfig(rootCA)

	// assert
	if gotErr != nil || tlsConfig.InsecureSkipVerify == false || tlsConfig.RootCAs != nil {
		t.Errorf("Test_BuildTLSConfig_Empty_RootCA failed, gotErr= [%v], gotConfig= [%v]", gotErr, tlsConfig)
	}
}
