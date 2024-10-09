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

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/s3/policy"
	"github.com/huawei/cosi-driver/pkg/user"
	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/user/clientset/poe"
)

func Test_removeUser_Normal_Success(t *testing.T) {
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
		t.Errorf("Test_removeUser_Normal_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_removeBucketPolicyStatement_Normal_Success(t *testing.T) {
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
		t.Errorf("Test_removeBucketPolicyStatement_Normal_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_removeBucketPolicyStatement_PolicyNotExist(t *testing.T) {
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
		t.Errorf("Test_removeBucketPolicyStatement_PolicyNotExist failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_removeBucketPolicyStatement_StatementNotExist(t *testing.T) {
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
		t.Errorf("Test_removeBucketPolicyStatement_StatementNotExist failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	//cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
