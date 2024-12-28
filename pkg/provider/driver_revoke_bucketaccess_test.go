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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/s3/policy"
	"github.com/huawei/cosi-driver/pkg/user"
	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/user/clientset/poe"
	"github.com/huawei/cosi-driver/pkg/utils/keylock"
)

func Test_ProvisionerServer_DriverRevokeBucketAccess_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	req := &cosispec.DriverRevokeBucketAccessRequest{}
	s := &provisionerServer{
		K8sClient: fake.NewSimpleClientset(),
		keyLock:   keylock.NewKeyLock(keyLockSize),
	}
	bacResource := &resourceIdInfo{}
	bacSecret := &coreV1.Secret{}
	bcResource := &resourceIdInfo{}
	bcSecret := &coreV1.Secret{}

	wantResponse := &cosispec.DriverRevokeBucketAccessResponse{}

	// mock
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, nil).
		ApplyFuncReturn(fetchDataFromResourceId, bacResource, bacSecret, nil).
		ApplyFuncReturn(removeUser, nil).
		ApplyFuncReturn(fetchDataFromResourceId, bcResource, bcSecret, nil).
		ApplyFuncReturn(removeBucketPolicyStatement, nil)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	if !reflect.DeepEqual(gotResponse, wantResponse) || gotErr != nil {
		t.Errorf("Test_ProvisionerServer_DriverRevokeBucketAccess_Success failed, "+
			"wantResponse= [%v], gotResponse= [%v], wantErr= nil, gotErr= [%v]", wantResponse, gotResponse, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_ProvisionerServer_DriverRevokeBucketAccess_CheckDriverRevokeBucketAccess_Failed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	req := &cosispec.DriverRevokeBucketAccessRequest{}
	s := &provisionerServer{
		K8sClient: fake.NewSimpleClientset(),
		keyLock:   keylock.NewKeyLock(keyLockSize),
	}
	checkErr := fmt.Errorf("check error")
	msg := fmt.Sprintf("check DriverGrantBucketAccessRequest failed, error is [%v]", checkErr)
	wantErr := status.Error(codes.Internal, msg)

	// mock
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, checkErr)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	if gotResponse != nil || !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("Test_ProvisionerServer_DriverRevokeBucketAccess_CheckDriverRevokeBucketAccess_Failed failed, "+
			"wantResponse= nil, gotResponse= [%v], wantErr= [%v], gotErr= [%v]", gotResponse, wantErr, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_ProvisionerServer_DriverRevokeBucketAccess_FetchBacDataFromResourceId_Failed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	req := &cosispec.DriverRevokeBucketAccessRequest{}
	s := &provisionerServer{
		K8sClient: fake.NewSimpleClientset(),
		keyLock:   keylock.NewKeyLock(keyLockSize),
	}
	fetchErr := fmt.Errorf("fetch error")
	msg := fmt.Sprintf("fetch data from resourceId [%s] failed, error is [%v]", req.GetAccountId(), fetchErr)
	wantErr := status.Error(codes.Internal, msg)

	// mock
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, nil).
		ApplyFuncReturn(fetchDataFromResourceId, nil, nil, fetchErr)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	if gotResponse != nil || !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("Test_ProvisionerServer_DriverRevokeBucketAccess_FetchBacDataFromResourceId_Failed failed, "+
			"wantResponse= nil, gotResponse= [%v], wantErr= [%v], gotErr= [%v]", gotResponse, wantErr, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_ProvisionerServer_DriverRevokeBucketAccess_RemoveUser_Failed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	req := &cosispec.DriverRevokeBucketAccessRequest{}
	s := &provisionerServer{
		K8sClient: fake.NewSimpleClientset(),
		keyLock:   keylock.NewKeyLock(keyLockSize),
	}

	resource := &resourceIdInfo{}
	sec := &coreV1.Secret{}

	removeUserErr := fmt.Errorf("remove user error")
	msg := fmt.Sprintf("remove user [%s] failed, error is [%v]", resource.resourceName, removeUserErr)
	wantErr := status.Error(codes.Internal, msg)

	// mock
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, nil).
		ApplyFuncReturn(fetchDataFromResourceId, resource, sec, nil).
		ApplyFuncReturn(removeUser, removeUserErr)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	if gotResponse != nil || !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("Test_ProvisionerServer_DriverRevokeBucketAccess_RemoveUser_Failed failed, "+
			"wantResponse= nil, gotResponse= [%v], wantErr= [%v], gotErr= [%v]", gotResponse, wantErr, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_ProvisionerServer_DriverRevokeBucketAccess_RemoveBucketPolicyStatement_Failed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	req := &cosispec.DriverRevokeBucketAccessRequest{}
	s := &provisionerServer{
		K8sClient: fake.NewSimpleClientset(),
		keyLock:   keylock.NewKeyLock(keyLockSize),
	}
	bacResource := &resourceIdInfo{}
	bacSecret := &coreV1.Secret{}
	bcResource := &resourceIdInfo{}
	bcSecret := &coreV1.Secret{}

	removeBucketPolicyStatementErr := fmt.Errorf("remove bp statement error")
	msg := fmt.Sprintf("remove bucket policy statement of user [%s] failed, "+
		"error is [%v]", bacResource.resourceName, removeBucketPolicyStatementErr)
	wantErr := status.Error(codes.Internal, msg)

	// mock
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, nil).
		ApplyFuncReturn(fetchDataFromResourceId, bacResource, bacSecret, nil).
		ApplyFuncReturn(removeUser, nil).
		ApplyFuncReturn(fetchDataFromResourceId, bcResource, bcSecret, nil).
		ApplyFuncReturn(removeBucketPolicyStatement, removeBucketPolicyStatementErr)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	if gotResponse != nil || !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("Test_ProvisionerServer_DriverRevokeBucketAccess_RemoveBucketPolicyStatement_Failed failed, "+
			"wantResponse= nil, gotResponse= [%v], wantErr= [%v], gotErr= [%v]", gotResponse, wantErr, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_RemoveUser_Normal_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	accountSecret := &coreV1.Secret{}
	userName := "user-demo"
	c := &poe.Client{}
	listUserAksResp := &api.ListUserAccessKeysOutput{AccessKeys: []string{"ak-1"}}

	// mock
	mock := gomonkey.
		ApplyFunc(user.NewUserClient, func(user.Config) (api.UserAPI, error) {
			return c, nil
		}).ApplyMethod(reflect.TypeOf(c), "ListUserAccessKeys",
		func(_ *poe.Client, ctx context.Context, in *api.ListUserAccessKeysInput) (*api.ListUserAccessKeysOutput, error) {
			return listUserAksResp, nil
		}).ApplyMethod(reflect.TypeOf(c), "DeleteUserAccess",
		func(_ *poe.Client, ctx context.Context, in *api.DeleteUserAccessInput) (*api.DeleteUserAccessOutput, error) {
			return nil, nil
		}).ApplyMethod(reflect.TypeOf(c), "DeleteUser",
		func(_ *poe.Client, ctx context.Context, in *api.DeleteUserInput) (*api.DeleteUserOutput, error) {
			return nil, nil
		})

	// act
	gotErr := removeUser(ctx, accountSecret, userName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_RemoveUser_Normal_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_RemoveBucketPolicyStatement_Normal_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &agent.S3Agent{}
	accountSecret := &coreV1.Secret{}
	userName := "user-demo"
	bucketName := "bucket-demo"

	statement := policy.Statement{
		Sid: userName,
	}
	mockBp := &policy.BucketPolicy{Statement: []policy.Statement{statement}}

	// mock
	mock := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(agent.Config) (*agent.S3Agent, error) {
			return c, nil
		}).ApplyMethod(reflect.TypeOf(c), "GetBucketPolicy",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) (*policy.BucketPolicy, error) {
			return mockBp, nil
		}).ApplyMethod(reflect.TypeOf(c), "PutBucketPolicy",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string, bp *policy.BucketPolicy) error {
			return nil
		}).ApplyMethod(reflect.TypeOf(c), "DeleteBucketPolicy",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) error {
			return nil
		})

	// act
	gotErr := removeBucketPolicyStatement(ctx, accountSecret, bucketName, userName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_RemoveBucketPolicyStatement_Normal_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_RemoveBucketPolicyStatement_PolicyNotExist(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &agent.S3Agent{}
	accountSecret := &coreV1.Secret{}
	userName := "user-demo"
	bucketName := "bucket-demo"

	// mock
	mock := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(agent.Config) (*agent.S3Agent, error) {
			return c, nil
		}).ApplyMethod(reflect.TypeOf(c), "GetBucketPolicy",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) (*policy.BucketPolicy, error) {
			return nil, nil
		})

	// act
	gotErr := removeBucketPolicyStatement(ctx, accountSecret, bucketName, userName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_RemoveBucketPolicyStatement_PolicyNotExist failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_RemoveBucketPolicyStatement_StatementNotExist(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &agent.S3Agent{}
	accountSecret := &coreV1.Secret{}
	userName := "user-demo"
	bucketName := "bucket-demo"

	statmentUserName := "user-demo-2"
	statement := policy.Statement{
		Sid: statmentUserName,
	}
	mockBp := policy.NewBucketPolicy(statement)

	// mock
	mock := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(agent.Config) (*agent.S3Agent, error) {
			return c, nil
		}).ApplyMethod(reflect.TypeOf(c), "GetBucketPolicy",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) (*policy.BucketPolicy, error) {
			return mockBp, nil
		})

	// act
	gotErr := removeBucketPolicyStatement(ctx, accountSecret, bucketName, userName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_RemoveBucketPolicyStatement_StatementNotExist failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_CheckDriverRevokeBucketAccess_Success(t *testing.T) {
	// arrange
	req := &cosispec.DriverRevokeBucketAccessRequest{BucketId: "id", AccountId: "id"}

	// act
	gotErr := checkDriverRevokeBucketAccess(req)

	// assert
	if gotErr != nil {
		t.Errorf("Test_CheckDriverRevokeBucketAccess_Success failed, "+
			"gotErr= [%v], wantErr= nil", gotErr)
	}
}

func Test_CheckDriverRevokeBucketAccess_EmptyBucketId(t *testing.T) {
	// arrange
	req := &cosispec.DriverRevokeBucketAccessRequest{}
	wantErr := fmt.Errorf("empty bucket id")

	// act
	gotErr := checkDriverRevokeBucketAccess(req)

	// assert
	if gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverRevokeBucketAccess_EmptyBucketId failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
}

func Test_CheckDriverRevokeBucketAccess_EmptyAccountId(t *testing.T) {
	// arrange
	req := &cosispec.DriverRevokeBucketAccessRequest{BucketId: "id"}
	wantErr := fmt.Errorf("empty account id")

	// act
	gotErr := checkDriverRevokeBucketAccess(req)

	// assert
	if gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverRevokeBucketAccess_EmptyAccountId failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
}
