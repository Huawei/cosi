/*
 Copyright (c) Huawei Technologies Co., Ltd. 2024-2026. All rights reserved.

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
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	assert.NoError(t, gotErr)
	assert.Equal(t, wantResponse, gotResponse)

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
	msg := fmt.Sprintf("check DriverRevokeBucketAccessRequest failed, error is [%v]", checkErr)
	wantErr := status.Error(codes.Internal, msg)

	// mock
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, checkErr)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	assert.Nil(t, gotResponse)
	assert.Equal(t, wantErr.Error(), gotErr.Error())

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
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, nil)
	patches.ApplyFuncReturn(fetchDataFromResourceId, nil, nil, fetchErr)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	assert.Nil(t, gotResponse)
	assert.Equal(t, wantErr.Error(), gotErr.Error())

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
	assert.Nil(t, gotResponse)
	assert.Equal(t, wantErr.Error(), gotErr.Error())

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
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, nil)
	patches.ApplyFuncReturn(fetchDataFromResourceId, bacResource, bacSecret, nil).
		ApplyFuncReturn(removeUser, nil).
		ApplyFuncReturn(fetchDataFromResourceId, bcResource, bcSecret, nil).
		ApplyFuncReturn(removeBucketPolicyStatement, removeBucketPolicyStatementErr)

	// act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// assert
	assert.Nil(t, gotResponse)
	assert.Equal(t, wantErr.Error(), gotErr.Error())

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_RemoveUser_Normal_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	accountSecret := &coreV1.Secret{
		Data: map[string][]byte{
			ak:       []byte("accessKey123"),
			sk:       []byte("secretKey123"),
			endpoint: []byte("https://xxxx.com:8088"),
		},
	}
	userName := "user-demo"
	c := &poe.Client{}
	listUserAksResp := &api.ListUserAccessKeysOutput{AccessKeys: []string{"ak-1"}}

	// mock
	mock := gomonkey.ApplyFuncReturn(user.NewUserClient, c, nil).
		ApplyMethodReturn(c, "ListUserAccessKeys", listUserAksResp, nil).
		ApplyMethodReturn(c, "DeleteUserAccess", nil, nil).
		ApplyMethodReturn(c, "DeleteUser", nil, nil)

	// act
	gotErr := removeUser(ctx, accountSecret, userName)

	// assert
	assert.NoError(t, gotErr)

	// cleanup
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
	mock := gomonkey.ApplyFuncReturn(agent.NewS3Agent, c, nil)
	mock.ApplyMethodReturn(c, "GetBucketPolicy", mockBp, nil)
	mock.ApplyMethodReturn(c, "PutBucketPolicy", nil)
	mock.ApplyMethodReturn(c, "DeleteBucketPolicy", nil)

	// act
	gotErr := removeBucketPolicyStatement(ctx, accountSecret, bucketName, userName)

	// assert
	assert.NoError(t, gotErr)

	// cleanup
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
	mock := gomonkey.ApplyFuncReturn(agent.NewS3Agent, c, nil).
		ApplyMethodReturn(c, "GetBucketPolicy", nil, nil)

	// act
	gotErr := removeBucketPolicyStatement(ctx, accountSecret, bucketName, userName)

	// assert
	assert.NoError(t, gotErr)

	// cleanup
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

	statementUserName := "user-demo-2"
	statement := policy.Statement{
		Sid: statementUserName,
	}
	mockBp := policy.NewBucketPolicy(statement)

	// mock
	mock := gomonkey.ApplyFuncReturn(agent.NewS3Agent, c, nil).
		ApplyMethodReturn(c, "GetBucketPolicy", mockBp, nil)

	// act
	gotErr := removeBucketPolicyStatement(ctx, accountSecret, bucketName, userName)

	// assert
	assert.NoError(t, gotErr)

	// cleanup
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
	assert.NoError(t, gotErr)
}

func Test_CheckDriverRevokeBucketAccess_EmptyBucketId(t *testing.T) {
	// arrange
	req := &cosispec.DriverRevokeBucketAccessRequest{}
	wantErr := fmt.Errorf("empty bucket id")

	// act
	gotErr := checkDriverRevokeBucketAccess(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverRevokeBucketAccess_EmptyAccountId(t *testing.T) {
	// arrange
	req := &cosispec.DriverRevokeBucketAccessRequest{BucketId: "id"}
	wantErr := fmt.Errorf("empty account id")

	// act
	gotErr := checkDriverRevokeBucketAccess(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func TestDriverRevokeBucketAccessWithCentralizedClientSuccess(t *testing.T) {
	// Arrange
	ctx := context.TODO()
	req := &cosispec.DriverRevokeBucketAccessRequest{
		BucketId:  "namespace/bucket/bucket-name",
		AccountId: "default/account-secret/user-demo",
	}
	s := &provisionerServer{K8sClient: fake.NewSimpleClientset(), keyLock: keylock.NewKeyLock(keyLockSize)}

	// Create account secret with centralized credentials (username/password)
	accountSecret := &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{Name: "account-secret", Namespace: "default"},
		Data: map[string][]byte{
			username: []byte("admin"),
			password: []byte("password123"),
			endpoint: []byte("https://xxxx.com:8088"),
		},
	}
	bcSecret := &coreV1.Secret{
		Data: map[string][]byte{
			ak:       []byte("accessKey123"),
			sk:       []byte("secretKey123"),
			endpoint: []byte("https://xxxx.com:8088"),
			rootCA:   []byte(""),
		},
	}
	resource := &resourceIdInfo{acSecretNameSpace: "default", acSecretName: "account-secret", resourceName: "user-demo"}

	// Create secrets in k8s
	_, err := s.K8sClient.CoreV1().Secrets("default").Create(ctx, accountSecret, metaV1.CreateOptions{})
	assert.NoError(t, err)

	_, err = s.K8sClient.CoreV1().Secrets("default").Create(ctx, bcSecret, metaV1.CreateOptions{})
	assert.NoError(t, err)

	// Mock
	patches := gomonkey.ApplyFuncReturn(checkDriverRevokeBucketAccess, nil).
		ApplyFuncReturn(fetchDataFromResourceId, resource, accountSecret, nil).
		ApplyFuncReturn(removeUser, nil).
		ApplyFuncReturn(removeBucketPolicyStatement, nil)
	defer patches.Reset()

	// Act
	gotResponse, gotErr := s.DriverRevokeBucketAccess(ctx, req)

	// Assert
	assert.NoError(t, gotErr, "DriverRevokeBucketAccess should not fail")
	assert.NotNil(t, gotResponse, "response should not be nil")
}
