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
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/cosi-driver/pkg/utils"
)

func Test_NewDriver_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	driverName := "demo"
	opts := KubeConfigOptions{
		KubeConfigPath: "demo-path",
		ClientConfig: utils.ClientConfig{
			QPS:   utils.DefaultClientQPS,
			Burst: utils.DefaultClientBurst,
		},
	}
	is := &identityServer{}
	ps := &provisionerServer{}

	// mock
	patches := gomonkey.ApplyFuncReturn(NewIdentityServer, is, nil).
		ApplyFuncReturn(NewProvisionerServer, ps, nil)

	// act
	_, _, gotErr := NewDriver(ctx, driverName, opts)

	// assert
	if gotErr != nil {
		t.Errorf("Test_NewDriver_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_NewDriver_NewIdentityServer_Failed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	driverName := "demo"
	opts := KubeConfigOptions{
		KubeConfigPath: "demo-path",
		ClientConfig: utils.ClientConfig{
			QPS:   utils.DefaultClientQPS,
			Burst: utils.DefaultClientBurst,
		},
	}
	newErr := fmt.Errorf("new is error")
	wantErr := newErr

	// mock
	patches := gomonkey.ApplyFuncReturn(NewIdentityServer, nil, newErr)

	// act
	_, _, gotErr := NewDriver(ctx, driverName, opts)

	// assert
	if gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_NewDriver_NewIdentityServer_Failed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_NewDriver_NewProvisionerServer_Failed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	driverName := "demo"
	opts := KubeConfigOptions{
		KubeConfigPath: "demo-path",
		ClientConfig: utils.ClientConfig{
			QPS:   utils.DefaultClientQPS,
			Burst: utils.DefaultClientBurst,
		},
	}
	newErr := fmt.Errorf("new is error")
	wantErr := newErr
	is := &identityServer{}

	// mock
	patches := gomonkey.ApplyFuncReturn(NewIdentityServer, is, nil).
		ApplyFuncReturn(NewProvisionerServer, nil, newErr)

	// act
	_, _, gotErr := NewDriver(ctx, driverName, opts)

	// assert
	if gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_NewDriver_NewProvisionerServer_Failed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func TestNewDriverCustomQPSAndBurstSuccess(t *testing.T) {
	// arrange
	ctx := context.TODO()
	driverName := "cosi.custom.com"
	opts := KubeConfigOptions{
		KubeConfigPath: "custom-path",
		ClientConfig: utils.ClientConfig{
			QPS:   100.5,
			Burst: 200,
		},
	}
	is := &identityServer{}
	ps := &provisionerServer{}

	// mock
	patches := gomonkey.ApplyFuncReturn(NewIdentityServer, is, nil).
		ApplyFuncReturn(NewProvisionerServer, ps, nil)

	// act
	gotIS, gotPS, gotErr := NewDriver(ctx, driverName, opts)

	// assert
	if gotErr != nil {
		t.Errorf("TestNewDriverCustomQPSAndBurstSuccess failed, gotErr= [%v], wantErr= nil", gotErr)
	}
	if gotIS == nil {
		t.Errorf("TestNewDriverCustomQPSAndBurstSuccess failed, gotIS= nil, wantErr= not nil")
	}
	if gotPS == nil {
		t.Errorf("TestNewDriverCustomQPSAndBurstSuccess failed, gotPS= nil, wantErr= not nil")
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func TestNewDriverDefaultQPSAndBurstSuccess(t *testing.T) {
	// arrange
	ctx := context.TODO()
	driverName := ""
	opts := KubeConfigOptions{
		KubeConfigPath: "demo-path",
		ClientConfig: utils.ClientConfig{
			QPS:   utils.DefaultClientQPS,
			Burst: utils.DefaultClientBurst,
		},
	}
	is := &identityServer{}
	ps := &provisionerServer{}

	// mock
	patches := gomonkey.ApplyFuncReturn(NewIdentityServer, is, nil).
		ApplyFuncReturn(NewProvisionerServer, ps, nil)

	// act
	_, _, gotErr := NewDriver(ctx, driverName, opts)

	// assert
	if gotErr != nil {
		t.Errorf("TestNewDriverDefaultQPSAndBurstSuccess failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func TestNewDriverBothServersFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	driverName := "demo"
	opts := KubeConfigOptions{
		KubeConfigPath: "demo-path",
		ClientConfig: utils.ClientConfig{
			QPS:   utils.DefaultClientQPS,
			Burst: utils.DefaultClientBurst,
		},
	}
	isErr := fmt.Errorf("new is error")
	psErr := fmt.Errorf("new ps error")

	// mock
	patches := gomonkey.ApplyFuncReturn(NewIdentityServer, nil, isErr).
		ApplyFuncReturn(NewProvisionerServer, nil, psErr)

	// act
	_, _, gotErr := NewDriver(ctx, driverName, opts)

	// assert
	if gotErr == nil {
		t.Errorf("TestNewDriverBothServersFailed failed, gotErr= nil, wantErr= not nil")
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func TestNewDriverEmptyNameSuccess(t *testing.T) {
	// arrange
	ctx := context.TODO()
	driverName := ""
	opts := KubeConfigOptions{
		KubeConfigPath: "",
		ClientConfig: utils.ClientConfig{
			QPS:   1.0,
			Burst: 5,
		},
	}
	is := &identityServer{}
	ps := &provisionerServer{}

	// mock
	patches := gomonkey.ApplyFuncReturn(NewIdentityServer, is, nil).
		ApplyFuncReturn(NewProvisionerServer, ps, nil)

	// act
	gotIS, gotPS, gotErr := NewDriver(ctx, driverName, opts)

	// assert
	if gotErr != nil {
		t.Errorf("TestNewDriverEmptyNameSuccess failed, gotErr= [%v], wantErr= nil", gotErr)
	}
	if gotIS == nil {
		t.Errorf("TestNewDriverEmptyNameSuccess failed, gotIS= nil")
	}
	if gotPS == nil {
		t.Errorf("TestNewDriverEmptyNameSuccess failed, gotPS= nil")
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}
