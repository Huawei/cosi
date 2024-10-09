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

// Package agent provides s3 agent and its apis
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/huawei/cosi-driver/pkg/s3/errors"
	"github.com/huawei/cosi-driver/pkg/s3/policy"
)

func Test_S3Agent_PutBucketPolicy_Success(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "PutBucketPolicy",
		func(_ *s3.S3, input *s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error) {
			return nil, nil
		})

	// act
	gotErr := s3Agent.PutBucketPolicy(ctx, "bucket-demo", &policy.BucketPolicy{}, []string{})

	// assert
	if gotErr != nil {
		t.Errorf("Test_S3Agent_PutBucketPolicy_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_PutBucketPolicy_Failed(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()
	var errCode = "other code"
	var errMsg = "other error"
	clientPutBucketPolicyErr := awserr.New(errCode, errMsg, fmt.Errorf("s3 client error"))
	wantErr := fmt.Errorf("put bucket policy failed, error is [%v]", clientPutBucketPolicyErr)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "PutBucketPolicy",
		func(_ *s3.S3, input *s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error) {
			return nil, clientPutBucketPolicyErr
		})

	// act
	gotErr := s3Agent.PutBucketPolicy(ctx, "bucket-demo", &policy.BucketPolicy{}, []string{})

	// assert
	if !reflect.DeepEqual(wantErr, gotErr) {
		t.Errorf("Test_S3Agent_PutBucketPolicy_Failed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_GetBucketPolicy_Success(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()

	sid := "sid-test"
	e := policy.EffectAllow
	userArn := "arn:aws:iam::domain-id:user/user-name"
	ac := policy.AllowedReadActions
	bucketName := "bucket-name"
	s := policy.NewStatementBuilder().WithSID(sid).WithEffect(e).WithPrincipals(userArn).WithActions(ac).
		WithResources(bucketName).WithSubResources(bucketName).Build()
	wantPolicy := policy.NewBucketPolicy(*s)
	policyByte, _ := json.Marshal(wantPolicy)
	mockPolicyString := string(policyByte)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "GetBucketPolicy",
		func(_ *s3.S3, input *s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error) {
			return &s3.GetBucketPolicyOutput{Policy: &mockPolicyString}, nil
		})

	// act
	gotPolicy, gotErr := s3Agent.GetBucketPolicy(ctx, "bucket-demo", []string{})

	// assert
	if !reflect.DeepEqual(wantPolicy, gotPolicy) || gotErr != nil {
		t.Errorf("Test_S3Agent_GetBucketPolicy_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_GetBucketPolicy_Failed(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()
	var errCode = "other code"
	var errMsg = "other error"
	clientGetBucketPolicyErr := awserr.New(errCode, errMsg, fmt.Errorf("s3 client error"))
	wantErr := fmt.Errorf("get bucket policy failed, error is [%v]", clientGetBucketPolicyErr)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "GetBucketPolicy",
		func(_ *s3.S3, input *s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error) {
			return nil, clientGetBucketPolicyErr
		})

	// act
	_, gotErr := s3Agent.GetBucketPolicy(ctx, "bucket-demo", []string{})

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("Test_S3Agent_GetBucketPolicy_Failed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_GetBucketPolicy_PolicyNotExist_Exceptional_Case(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()
	bucketName := "bucket-demo"

	var errCode = errors.ErrNoSuchBucketPolicy
	var errMsg = "The bucket policy does not exist"
	wantAwsErr := awserr.New(errCode, errMsg, fmt.Errorf("s3 client error"))

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "GetBucketPolicy",
		func(_ *s3.S3, input *s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error) {
			return nil, wantAwsErr
		})

	// act
	gotBp, gotErr := s3Agent.GetBucketPolicy(ctx, bucketName, []string{errors.ErrNoSuchBucketPolicy})

	// assert
	if gotBp != nil || gotErr != nil {
		t.Errorf("Test_S3Agent_GetBucketPolicy_PolicyNotExist_Exceptional_Case failed, gotErr= [%v], wantErr= nil, "+
			"gotBp= [%v], wantBp= nil", gotErr, gotBp)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_GetBucketPolicy_BucketNotExist_NotExceptional_Case(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()
	bucketName := "bucket-demo"

	var errCode = errors.ErrNoSuchBucket
	var errMsg = "The bucket does not exist"
	wantAwsErr := awserr.New(errCode, errMsg, fmt.Errorf("s3 client error"))
	wantErr := fmt.Errorf("get bucket policy failed, error is [%v]", wantAwsErr)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "GetBucketPolicy",
		func(_ *s3.S3, input *s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error) {
			return nil, wantAwsErr
		})

	// act
	gotBp, gotErr := s3Agent.GetBucketPolicy(ctx, bucketName, []string{})

	// assert
	if gotBp != nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_S3Agent_GetBucketPolicy_BucketNotExist_NotExceptional_Case failed, gotErr= [%v], wantErr= nil, "+
			"gotBp= [%v], wantBp= nil", gotErr, gotBp)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_DeleteBucketPolicy_Success(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "DeleteBucketPolicy",
		func(_ *s3.S3, input *s3.DeleteBucketPolicyInput) (*s3.DeleteBucketPolicyOutput, error) {
			return nil, nil
		})

	// act
	gotErr := s3Agent.DeleteBucketPolicy(ctx, "bucket-demo", []string{})

	// assert
	if gotErr != nil {
		t.Errorf("Test_S3Agent_DeleteBucketPolicy_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_DeleteBucketPolicy_BucketNotExist_Exceptional_Case(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	ctx := context.TODO()
	var errCode = errors.ErrNoSuchBucket
	var errMsg = "The bucket does not exist"
	wantAwsErr := awserr.New(errCode, errMsg, fmt.Errorf("s3 client error"))

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "DeleteBucketPolicy",
		func(_ *s3.S3, input *s3.DeleteBucketPolicyInput) (*s3.DeleteBucketPolicyOutput, error) {
			return nil, wantAwsErr
		})

	// act
	gotErr := s3Agent.DeleteBucketPolicy(ctx, "bucket-demo", []string{errors.ErrNoSuchBucket})

	// assert
	if gotErr != nil {
		t.Errorf("Test_S3Agent_DeleteBucketPolicy_BucketNotExist failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
