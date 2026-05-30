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
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/user"
	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/user/clientset/poe"
	"github.com/huawei/cosi-driver/pkg/utils/keylock"
)

func Test_ProvisionerServer_DriverGrantBucketAccess_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	req := &cosispec.DriverGrantBucketAccessRequest{}
	s := &provisionerServer{
		K8sClient: fake.NewSimpleClientset(),
		keyLock:   keylock.NewKeyLock(keyLockSize),
	}
	bacSecret := &coreV1.Secret{}
	bcResource := &resourceIdInfo{}
	bcSecret := &coreV1.Secret{}
	userData := &userInfo{}

	_, _ = s.K8sClient.CoreV1().Secrets(bacSecret.Namespace).Create(ctx, bacSecret, metaV1.CreateOptions{})

	wantResponse := &cosispec.DriverGrantBucketAccessResponse{
		AccountId:   assembleResourceId(bacSecret.Namespace, bacSecret.Name, req.Name),
		Credentials: buildCredentials(bcSecret, userData),
	}

	// mock
	patches := gomonkey.ApplyFuncReturn(checkDriverGrantBucketAccessRequest, nil)
	patches.ApplyFuncReturn(fetchDataFromResourceId, bcResource, bcSecret, nil)
	patches.ApplyFuncReturn(checkBucketExistence, nil)
	patches.ApplyFuncReturn(registerUser, userData, nil)
	patches.ApplyFuncReturn(setBucketPolicy, nil)

	// act
	gotResponse, gotErr := s.DriverGrantBucketAccess(ctx, req)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantResponse, gotResponse)

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_RegisterUser_NewUser_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	accountSecret := &coreV1.Secret{
		Data: map[string][]byte{
			ak:       []byte("fake-ak"),
			sk:       []byte("fake-sk"),
			endpoint: []byte("https://xxxx.com:8088"),
		},
	}
	userName := "user-demo"
	userArn := "arn-id"
	userId := "user-id"
	userAk := "ak-id"
	userSk := "sk-id"
	req := &cosispec.DriverGrantBucketAccessRequest{Name: userName}
	c := &poe.Client{}
	createUserResp := &api.CreateUserOutput{UserName: userName, UserID: userId, Arn: userArn}
	createUserAccessResp := &api.CreateUserAccessOutput{AccessKeyId: userAk, SecretAccessKey: userSk}
	wantUserData := &userInfo{userArn: userArn, accessKeyId: userAk, accessSecretKey: userSk}

	// mock
	mock := gomonkey.ApplyFuncReturn(user.NewUserClient, c, nil)
	mock.ApplyMethodReturn(c, "GetUser", nil, nil)
	mock.ApplyMethodReturn(c, "CreateUser", createUserResp, nil)
	mock.ApplyMethodReturn(c, "CreateUserAccess", createUserAccessResp, nil)

	// act
	gotUserData, gotErr := registerUser(ctx, req, accountSecret)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantUserData, gotUserData)

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_SetBucketPolicy_NewPolicy_Success(t *testing.T) {
	// arrange
	userName := "user-demo"
	userArn := "arn-id"
	userAk := "ak-id"
	userSk := "sk-id"
	bucketName := "bucket-demo"

	ctx := context.TODO()
	c := &agent.S3Agent{}
	accountSecret := &coreV1.Secret{}
	req := &cosispec.DriverGrantBucketAccessRequest{Name: userName}
	userData := &userInfo{userArn: userArn, accessKeyId: userAk, accessSecretKey: userSk}

	// mock
	mock := gomonkey.ApplyFuncReturn(agent.NewS3Agent, c, nil)
	mock.ApplyMethodReturn(c, "GetBucketPolicy", nil, nil)
	mock.ApplyMethodReturn(c, "PutBucketPolicy", nil)

	// act
	gotErr := setBucketPolicy(ctx, req, accountSecret, userData, bucketName)

	// assert
	assert.NoError(t, gotErr)

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_CheckDriverGrantBucketAccessRequest_EmptyBucketId(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = ""
	wantErr := fmt.Errorf("empty bucket id")

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverGrantBucketAccessRequest_EmptyUserName(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = "bucketId"
	req.Name = ""
	wantErr := fmt.Errorf("empty user name")

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverGrantBucketAccessRequest_IAMAuthenticationType(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = "bucketId"
	req.Name = "userName"
	req.AuthenticationType = cosispec.AuthenticationType_IAM

	wantErr := fmt.Errorf("IAM authentication type not implemented")

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverGrantBucketAccessRequest_UnknownAuthenticationType(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = "bucketId"
	req.Name = "userName"
	req.AuthenticationType = cosispec.AuthenticationType_UnknownAuthenticationType

	wantErr := fmt.Errorf("unknown authentication type")

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverGrantBucketAccessRequest_MissingAccountSecretName(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = "bucketId"
	req.Name = "userName"
	req.AuthenticationType = cosispec.AuthenticationType_Key
	req.Parameters = make(map[string]string)

	wantErr := fmt.Errorf("account secret name value is empty")

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverGrantBucketAccessRequest_MissingAccountSecretNamespace(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = "bucketId"
	req.Name = "userName"
	req.AuthenticationType = cosispec.AuthenticationType_Key
	req.Parameters = map[string]string{
		accountSecretName: "accountSecret",
	}

	wantErr := fmt.Errorf("account secret namespace value is empty")

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverGrantBucketAccessRequest_InvalidBucketPolicyModel(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = "bucketId"
	req.Name = "userName"
	req.AuthenticationType = cosispec.AuthenticationType_Key
	req.Parameters = map[string]string{
		accountSecretName:      "accountSecret",
		accountSecretNamespace: "accountSecretNamespace",
		bucketPolicyModel:      "invalidModel",
	}

	wantErr := fmt.Errorf("invalid bucketPolicy model [invalidModel]")

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())
}

func Test_CheckDriverGrantBucketAccessRequest_NormalCase(t *testing.T) {
	// arrange
	req := &cosispec.DriverGrantBucketAccessRequest{}
	req.BucketId = "bucketId"
	req.Name = "userName"
	req.AuthenticationType = cosispec.AuthenticationType_Key
	req.Parameters = map[string]string{
		accountSecretName:      "accountSecret",
		accountSecretNamespace: "accountSecretNamespace",
		bucketPolicyModel:      bucketPolicyModelRW,
	}

	// act
	gotErr := checkDriverGrantBucketAccessRequest(req)

	// assert
	assert.NoError(t, gotErr)
}

func Test_BuildCredentials_Success(t *testing.T) {
	// arrange
	bcAccountSecret := &coreV1.Secret{
		Data: map[string][]byte{
			"endpoint": []byte("https://xxxx.com"),
		},
	}
	userData := &userInfo{
		accessKeyId:     "accessKeyId",
		accessSecretKey: "accessSecretKey",
	}

	// act
	gotCredDetails := buildCredentials(bcAccountSecret, userData)

	// assert
	wantCred := cosispec.CredentialDetails{
		Secrets: map[string]string{
			accessAk: userData.accessKeyId,
			accessSk: userData.accessSecretKey,
			endpoint: string(bcAccountSecret.Data[endpoint]),
		},
	}
	assert.Equal(t, wantCred.Secrets, gotCredDetails[s3Protocol].Secrets)
}

func Test_CheckBucketExistence_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &agent.S3Agent{}
	bucketName := "bucket-demo"
	secret := &coreV1.Secret{}

	// mock
	mock := gomonkey.ApplyFuncReturn(agent.NewS3Agent, c, nil)
	mock.ApplyMethodReturn(c, "CheckBucketExist", nil)

	// act
	gotErr := checkBucketExistence(ctx, secret, bucketName)

	// assert
	assert.NoError(t, gotErr)

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_CheckBucketExistence_NewAgent_Failed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	bucketName := "bucket-demo"
	secret := &coreV1.Secret{}
	angentErr := fmt.Errorf("s3 new agent error")
	wantErr := fmt.Errorf("new s3 agent failed, error is [%v]", angentErr)

	// mock
	mock := gomonkey.ApplyFuncReturn(agent.NewS3Agent, nil, angentErr)

	// act
	gotErr := checkBucketExistence(ctx, secret, bucketName)

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr.Error(), gotErr.Error())

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
