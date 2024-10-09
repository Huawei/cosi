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

// Package poe provides poe client and poe apis
package poe

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestNewPoeClient_Success(t *testing.T) {
	// arrange
	ak := "test-ak"
	sk := "test-sk"
	endpoint := "https://x.xx.xx.xx:443"
	wantEndpoint := "https://x.xx.xx.xx:9443"
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := &http.Client{Timeout: connectionTimeout, Transport: tr}
	rootCA := []byte("")
	wantClient := &Client{
		AccessKey:  ak,
		SecretKey:  sk,
		Endpoint:   wantEndpoint,
		HttpClient: httpClient,
		RootCA:     rootCA,
	}

	// act
	gotClient, err := NewPoeClient(endpoint, ak, sk, rootCA)

	// assert
	if !reflect.DeepEqual(gotClient, wantClient) || err != nil {
		t.Errorf("TestNewPoeClient_Success failed, gotClient= [%v], wantClient= [%v]", gotClient, wantClient)
	}
}

func TestNewPoeClient_InvalidEndpoint(t *testing.T) {
	// arrange
	ak := "test-ak"
	sk := "test-sk"
	endpoint := "https://x.xx.xx.xx:443:xx"
	errMsg := "parse \"https://x.xx.xx.xx:443:xx\": invalid port \":xx\" after host"
	wantErr := fmt.Errorf("url parse endpoint [%s] failed, error is [%v]", endpoint, errMsg)

	// act
	gotClient, gotErr := NewPoeClient(endpoint, ak, sk, []byte(""))

	// assert
	if gotClient != nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("TestNewPoeClient_InvalidEndpoint failed, gotClient= [%v], wantClient= [%v], "+
			"gotErr= [%v], wantErr= [%v]", gotClient, nil, gotErr, wantErr)
	}
}
