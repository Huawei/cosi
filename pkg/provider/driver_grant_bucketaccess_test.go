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
	patches := gomonkey.ApplyFuncReturn(checkDriverGrantBucketAccessRequest, nil).
		ApplyFuncReturn(fetchDataFromResourceId, bcResource, bcSecret, nil).
		ApplyFuncReturn(checkBucketExistence, nil).
		ApplyFuncReturn(registerUser, userData, nil).
		ApplyFuncReturn(setBucketPolicy, nil)

	// act
	gotResponse, gotErr := s.DriverGrantBucketAccess(ctx, req)

	// assert
	if !reflect.DeepEqual(gotResponse, wantResponse) || gotErr != nil {
		t.Errorf("Test_ProvisionerServer_DriverGrantBucketAccess_Success failed, "+
			"wantResponse= [%v], gotResponse= [%v], wantErr= nil, gotErr= [%v]", wantResponse, gotResponse, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_RegisterUser_NewUser_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	accountSecret := &coreV1.Secret{}
	userName := "user-demo"
	userArn := "arn-id"
	userId := "user-id"
	userAk := "ak-id"
	userSk := "sk-id"
	req := &cosispec.DriverGrantBucketAccessRequest{Name: userName}
	c := &poe.Client{}
	createUserResp := &api.CreateUserOutput{UserName: userName, UserID: userId, Arn: userArn}
	createUserAccessResp := &api.CreateUserAccessOutput{AccessKeyId: userAk, SecretAccessKey: userSk}

	wantUserData := userInfo{userArn: userArn, accessKeyId: userAk, accessSecretKey: userSk}

	// mock
	mock := gomonkey.
		ApplyFunc(user.NewUserClient, func(user.Config) (api.UserAPI, error) {
			return c, nil
		}).ApplyMethod(reflect.TypeOf(c), "GetUser",
		func(_ *poe.Client, ctx context.Context, in *api.GetUserInput) (*api.GetUserOutput, error) {
			return nil, nil
		}).ApplyMethod(reflect.TypeOf(c), "CreateUser",
		func(_ *poe.Client, ctx context.Context, in *api.CreateUserInput) (*api.CreateUserOutput, error) {
			return createUserResp, nil
		}).ApplyMethod(reflect.TypeOf(c), "CreateUserAccess",
		func(_ *poe.Client, ctx context.Context, in *api.CreateUserAccessInput) (*api.CreateUserAccessOutput, error) {
			return createUserAccessResp, nil
		})

	// act
	gotUserData, gotErr := registerUser(ctx, req, accountSecret)

	// assert
	if reflect.DeepEqual(gotUserData, wantUserData) || gotErr != nil {
		t.Errorf("Test_RegisterUser_NewUser_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", gotUserData, wantUserData, gotErr)
	}

	//cleanup
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
	mock := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(agent.Config) (*agent.S3Agent, error) {
			return c, nil
		}).ApplyMethod(reflect.TypeOf(c), "GetBucketPolicy",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) (*policy.BucketPolicy, error) {
			return nil, nil
		}).ApplyMethod(reflect.TypeOf(c), "PutBucketPolicy",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string, bp *policy.BucketPolicy) error {
			return nil
		})

	// act
	gotErr := setBucketPolicy(ctx, req, accountSecret, userData, bucketName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_SetBucketPolicy_NewPolicy_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
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
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_EmptyBucketId failed, "+
			"gotErr= [%s], wantErr= [%s]", gotErr, wantErr)
	}
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
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_EmptyUserName failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
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
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_IAMAuthenticationType failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
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
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_UnknownAuthenticationType failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
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
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_MissingAccountSecretName failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
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
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_MissingAccountSecretNamespace failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
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
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_InvalidBucketPolicyModel failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}
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
	if gotErr != nil {
		t.Errorf("Test_CheckDriverGrantBucketAccessRequest_NormalCase failed, gotErr= [%v], wantErr= [nil]", gotErr)
	}
}

func Test_BuildCredentials_Success(t *testing.T) {
	// arrange
	bcAccountSecret := &coreV1.Secret{
		Data: map[string][]byte{
			"endpoint": []byte("https://example.com"),
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
	if !reflect.DeepEqual(wantCred.Secrets, gotCredDetails[s3Protocol].Secrets) {
		t.Errorf("Test_BuildCredentials_Success failed, gotCredSecret= [%v], wantCredSecret= [%v]",
			gotCredDetails[s3Protocol].Secrets, wantCred.Secrets)
	}
}

func Test_CheckBucketExistence_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &agent.S3Agent{}
	bucketName := "bucket-demo"
	secret := &coreV1.Secret{}

	// mock
	mock := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(agent.Config) (*agent.S3Agent, error) {
			return c, nil
		}).ApplyMethod(reflect.TypeOf(c), "CheckBucketExist",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) error {
			return nil
		})

	// act
	gotErr := checkBucketExistence(ctx, secret, bucketName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_CheckBucketExistence_Success failed, wantErr= nil, gotErr= [%v]", gotErr)
	}

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
	mock := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(agent.Config) (*agent.S3Agent, error) {
			return nil, angentErr
		})

	// act
	gotErr := checkBucketExistence(ctx, secret, bucketName)

	// assert
	if gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_CheckBucketExistence_NewAgent_Failed failed, wangErr= [%v], gotErr= [%v]", wantErr, gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
