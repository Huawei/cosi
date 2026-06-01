/*
 Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

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

package provider

import (
	"context"
	"testing"

	cosispec "sigs.k8s.io/container-object-storage-interface-spec"
)

func Test_IdentityServer_DriverGetInfo_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	provisionerName := "cosi.huawei.com"
	req := &cosispec.DriverGetInfoRequest{}
	wantResp := &cosispec.DriverGetInfoResponse{
		Name: provisionerName,
	}

	is, err := NewIdentityServer(provisionerName)
	if err != nil {
		t.Errorf("Test_IdentityServer_DriverGetInfo_Success failed, create identityServer error= [%v]", err)
		return
	}

	// act
	gotResp, gotErr := is.DriverGetInfo(ctx, req)

	// assert
	if gotErr != nil {
		t.Errorf("Test_IdentityServer_DriverGetInfo_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	if gotResp == nil {
		t.Error("Test_IdentityServer_DriverGetInfo_Success failed, gotResp is nil")
	}

	if gotResp.Name != wantResp.Name {
		t.Errorf("Test_IdentityServer_DriverGetInfo_Success failed, gotName= [%s], wantName= [%s]",
			gotResp.Name, wantResp.Name)
	}
}

func Test_IdentityServer_DriverGetInfo_EmptyName(t *testing.T) {
	// arrange
	ctx := context.TODO()
	provisionerName := ""
	req := &cosispec.DriverGetInfoRequest{}

	is, err := NewIdentityServer(provisionerName)
	if err != nil {
		t.Errorf("Test_IdentityServer_DriverGetInfo_EmptyName failed, create identityServer error= [%v]", err)
		return
	}

	// act
	gotResp, gotErr := is.DriverGetInfo(ctx, req)

	// assert
	if gotErr != nil {
		t.Errorf("Test_IdentityServer_DriverGetInfo_EmptyName failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	if gotResp == nil {
		t.Error("Test_IdentityServer_DriverGetInfo_EmptyName failed, gotResp is nil")
	}

	if gotResp.Name != "" {
		t.Errorf("Test_IdentityServer_DriverGetInfo_EmptyName failed, gotName= [%s], wantName= empty",
			gotResp.Name)
	}
}
