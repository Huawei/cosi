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
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/huawei/cosi-driver/pkg/s3/policy"
	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// PutBucketPolicy applies the policy to the bucket
func (s *S3Agent) PutBucketPolicy(ctx context.Context, bucketName string, bp *policy.BucketPolicy,
	exceptionalErrCodes []string) error {
	log.AddContext(ctx).Infof("start to put bucket [%s] policy [%+v]", bucketName, bp)

	policyString, err := bp.ToJsonString()
	if err != nil {
		return fmt.Errorf("bucket policy to json string failed, error is [%v]", err)
	}

	input := &s3.PutBucketPolicyInput{
		Bucket: &bucketName,
		Policy: &policyString,
	}

	_, err = s.Client.PutBucketPolicy(input)
	if err != nil {
		var awsErr awserr.Error
		if !errors.As(err, &awsErr) {
			return fmt.Errorf("convert err to aws err failed, origin err is [%v]", err)
		}

		if !utils.ContainsElement(exceptionalErrCodes, awsErr.Code()) {
			return fmt.Errorf("put bucket policy failed, error is [%v]", err)
		} else {
			msg := fmt.Sprintf("exceptional case about putting bucket policy, message is [%s]", awsErr)
			log.AddContext(ctx).Infof(msg)
			return nil
		}
	}

	log.AddContext(ctx).Infof("put bucket [%s] policy successfully", bucketName)
	return nil
}

// GetBucketPolicy get bucket policy
func (s *S3Agent) GetBucketPolicy(ctx context.Context, bucketName string,
	exceptionalErrCodes []string) (*policy.BucketPolicy, error) {
	log.AddContext(ctx).Infof("start to get bucket [%s] policy", bucketName)

	input := &s3.GetBucketPolicyInput{
		Bucket: &bucketName,
	}

	out, err := s.Client.GetBucketPolicy(input)
	if err != nil {
		var awsErr awserr.Error
		if !errors.As(err, &awsErr) {
			return nil, fmt.Errorf("convert err to aws err failed, origin err is [%v]", err)
		}

		if !utils.ContainsElement(exceptionalErrCodes, awsErr.Code()) {
			return nil, fmt.Errorf("get bucket policy failed, error is [%v]", err)
		} else {
			msg := fmt.Sprintf("exceptional case about getting bucket policy, message is [%s]", awsErr)
			log.AddContext(ctx).Infof(msg)
			return nil, nil
		}
	}

	bp := &policy.BucketPolicy{}
	err = json.Unmarshal([]byte(*out.Policy), bp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal bucket policy failed, error is [%v]", err)
	}

	log.AddContext(ctx).Infof("get bucket [%s] policy successfully", bucketName)
	return bp, nil
}

// DeleteBucketPolicy delete the policy on bucket
func (s *S3Agent) DeleteBucketPolicy(ctx context.Context, bucketName string, exceptionalErrCodes []string) error {
	log.AddContext(ctx).Infof("start to delete bucket [%s] policy", bucketName)
	input := &s3.DeleteBucketPolicyInput{
		Bucket: &bucketName,
	}

	_, err := s.Client.DeleteBucketPolicy(input)
	if err != nil {
		var awsErr awserr.Error
		if !errors.As(err, &awsErr) {
			return fmt.Errorf("convert err to aws err failed, origin err is [%v]", err)
		}

		if !utils.ContainsElement(exceptionalErrCodes, awsErr.Code()) {
			return fmt.Errorf("delete bucket policy failed, error is [%v]", err)
		} else {
			msg := fmt.Sprintf("exceptional case about deleting bucket policy, message is [%s]", awsErr)
			log.AddContext(ctx).Infof(msg)
			return nil
		}
	}

	log.AddContext(ctx).Infof("delete bucket [%s] policy successfully", bucketName)
	return nil
}
