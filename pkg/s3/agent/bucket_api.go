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
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// CreateBucket creates a bucket with the given name
func (s *S3Agent) CreateBucket(ctx context.Context, bucketName, acl, location string) error {
	log.AddContext(ctx).Infof("start to create bucket, the bucketName is [%s], "+
		"acl is [%s], location is [%s]", bucketName, acl, location)

	bucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		ACL:    aws.String(acl),
	}

	if location != "" {
		bucketInput.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(location),
		}
	}

	_, err := s.Client.CreateBucket(bucketInput)
	if err != nil {
		return fmt.Errorf("create bucket failed, error is [%v]", err)
	}

	log.AddContext(ctx).Infof("create bucket [%s] successfully", bucketName)
	return nil
}

// DeleteBucket function deletes a bucket with the given name
func (s *S3Agent) DeleteBucket(ctx context.Context, bucketName string) error {
	log.AddContext(ctx).Infof("start to delete bucket [%s]", bucketName)

	_, err := s.Client.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(bucketName)})
	if err != nil {
		var awsErr awserr.Error
		if !errors.As(err, &awsErr) {
			return fmt.Errorf("convert err to aws err failed, origin err is [%v]", err)
		}

		if awsErr.Code() == s3.ErrCodeNoSuchBucket {
			log.AddContext(ctx).Infof("bucket [%s] does not exist", bucketName)
			return nil
		} else {
			return fmt.Errorf("delete bucket failed, error is [%v]", err)
		}
	}

	log.AddContext(ctx).Infof("delete bucket [%s] successfully", bucketName)
	return nil
}

// CheckBucketExist function check whether the bucket exists
func (s *S3Agent) CheckBucketExist(ctx context.Context, bucketName string) error {
	log.AddContext(ctx).Infof("start to check bucket [%s] existence", bucketName)

	_, err := s.Client.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(bucketName)})
	if err != nil {
		return err
	}

	log.AddContext(ctx).Infof("check bucket [%s] existence successfully", bucketName)
	return nil
}
