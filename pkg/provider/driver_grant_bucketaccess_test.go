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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	coreV1 "k8s.io/api/core/v1"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/s3/policy"
	"github.com/huawei/cosi-driver/pkg/user"
	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/user/clientset/poe"
)

func Test_registerUser_NewUser_Success(t *testing.T) {
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
		t.Errorf("Test_registerUser_NewUser_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", gotUserData, wantUserData, gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_setBucketPolicy_NewPolicy_Success(t *testing.T) {
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
		t.Errorf("Test_setBucketPolicy_NewPolicy_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
