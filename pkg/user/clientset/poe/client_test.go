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
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/cosi-driver/pkg/utils"
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

func Test_Client_Call_Success(t *testing.T) {
	// arrange
	pec := &Client{HttpClient: &http.Client{}, Endpoint: "ip//port"}
	ctx := context.TODO()
	param := make(map[string]string, 0)

	// mock
	patches := gomonkey.
		ApplyPrivateMethod(pec, "getStringToSign", func(_ *Client) string { return "getStringToSign" }).
		ApplyPrivateMethod(pec, "getSignature", func(_ *Client) (string, error) { return "signature", nil }).
		ApplyFuncReturn(utils.GetSortedUrlQueryString, "url").
		ApplyFuncReturn(http.NewRequest, &http.Request{}, nil).
		ApplyMethodReturn(pec.HttpClient, "Do", &http.Response{
			StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{"result":"ok"}`))}, nil).
		ApplyFuncReturn(io.ReadAll, []byte{}, nil)

	// act
	_, gotErr := pec.Call(ctx, param)

	// assert
	if gotErr != nil {
		t.Errorf("Test_Client_Call_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}
