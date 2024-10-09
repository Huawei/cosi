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
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Test_S3Agent_CreateBucket_Success(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	bucketName := ""
	acl := ""
	location := ""

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "CreateBucket",
		func(_ *s3.S3, input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
			return nil, nil
		})

	// act
	gotErr := s3Agent.CreateBucket(context.TODO(), bucketName, acl, location)

	// assert
	if gotErr != nil {
		t.Errorf("Test_S3Agent_CreateBucket_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_CreateBucket_Failed(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	bucketName := ""
	acl := ""
	location := ""
	CreateBucketErr := fmt.Errorf("internal error")
	wantErr := fmt.Errorf("create bucket failed, error is [%v]", CreateBucketErr)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "CreateBucket",
		func(_ *s3.S3, input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
			return nil, CreateBucketErr
		})

	// act
	gotErr := s3Agent.CreateBucket(context.TODO(), bucketName, acl, location)

	// assert
	if !reflect.DeepEqual(wantErr, gotErr) {
		t.Errorf("Test_S3Agent_CreateBucket_Failed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_DeleteBucket_Success(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	bucketName := ""

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "DeleteBucket",
		func(_ *s3.S3, input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
			return nil, nil
		})

	// act
	gotErr := s3Agent.DeleteBucket(context.TODO(), bucketName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_S3Agent_DeleteBucket_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_DeleteBucket_BucketNotExit_Success(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	bucketName := ""
	bucketExitErr := awserr.New(s3.ErrCodeNoSuchBucket, "The specified bucket does not exist", nil)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "DeleteBucket",
		func(_ *s3.S3, input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
			return nil, bucketExitErr
		})

	// act
	gotErr := s3Agent.DeleteBucket(context.TODO(), bucketName)

	// assert
	if gotErr != nil {
		t.Errorf("Test_S3Agent_DeleteBucket_BucketNotExit_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_S3Agent_DeleteBucket_Failed(t *testing.T) {
	// arrange
	s3Client := &s3.S3{}
	s3Agent := S3Agent{Client: s3Client}
	bucketName := ""
	var wantErr error = nil

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(s3Client), "DeleteBucket",
		func(_ *s3.S3, input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
			return nil, wantErr
		})

	// act
	gotErr := s3Agent.DeleteBucket(context.TODO(), bucketName)

	// assert
	if !reflect.DeepEqual(wantErr, gotErr) {
		t.Errorf("Test_S3Agent_DeleteBucket_Failed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
