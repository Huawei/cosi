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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	coreV1 "k8s.io/api/core/v1"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/s3/errors"
	"github.com/huawei/cosi-driver/pkg/user"
	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// DriverRevokeBucketAccess is used to revokes all access to a particular bucket from a principal
func (s *provisionerServer) DriverRevokeBucketAccess(ctx context.Context,
	req *cosispec.DriverRevokeBucketAccessRequest) (*cosispec.DriverRevokeBucketAccessResponse, error) {
	defer utils.RecoverPanic(ctx)

	// When processing the same bucket, it needs to add mutex lock.
	// Otherwise, policies will be overwritten abnormally in concurrent scenarios.
	s.keyLock.Lock(req.GetBucketId())
	defer s.keyLock.Unlock(req.GetBucketId())

	log.AddContext(ctx).Infof("handle DriverRevokeBucketAccess request, request is [%+v]", req)

	err := checkDriverRevokeBucketAccess(req)
	if err != nil {
		msg := fmt.Sprintf("check DriverGrantBucketAccessRequest failed, error is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	accountIdData, bacAccountSecret, err := fetchDataFromResourceId(req.GetAccountId(), s.K8sClient)
	if err != nil {
		msg := fmt.Sprintf("fetch data from resourceId [%s] failed, error is [%v]", req.GetAccountId(), err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	userName := accountIdData.resourceName
	err = removeUser(ctx, bacAccountSecret, userName)
	if err != nil {
		msg := fmt.Sprintf("remove user [%s] failed, error is [%v]", userName, err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	bucketIdData, bcAccountSecret, err := fetchDataFromResourceId(req.GetBucketId(), s.K8sClient)
	if err != nil {
		msg := fmt.Sprintf("fetch data from resourceId [%s] failed, error is [%v]", req.GetBucketId(), err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	bucketName := bucketIdData.resourceName
	err = removeBucketPolicyStatement(ctx, bcAccountSecret, bucketName, userName)
	if err != nil {
		msg := fmt.Sprintf("remove bucket policy statement of user [%s] failed, "+
			"error is [%v]", userName, err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	log.AddContext(ctx).Infof("handle DriverRevokeBucketAccess request successfully")
	return &cosispec.DriverRevokeBucketAccessResponse{}, nil
}

func checkDriverRevokeBucketAccess(req *cosispec.DriverRevokeBucketAccessRequest) error {
	if req.GetBucketId() == "" {
		return fmt.Errorf("empty bucket id")
	}

	if req.GetAccountId() == "" {
		return fmt.Errorf("empty account id")
	}

	return nil
}

func removeUser(ctx context.Context, bacAccountSecret *coreV1.Secret, userName string) error {
	// The user client type information will be obtained from the req config in the future.
	// Currently, default poe type
	userClient, err := user.NewUserClient(
		user.Config{
			ClientType: user.PoeType,
			AccessKey:  string(bacAccountSecret.Data[ak]),
			SecretKey:  string(bacAccountSecret.Data[sk]),
			Endpoint:   string(bacAccountSecret.Data[endpoint]),
			RootCA:     bacAccountSecret.Data[rootCA],
		})
	if err != nil {
		return fmt.Errorf("new user client failed, error is [%v]", err)
	}

	listUserAksResp, err := userClient.ListUserAccessKeys(ctx,
		&api.ListUserAccessKeysInput{UserName: userName})
	if err != nil {
		return fmt.Errorf("list user [%s] access keys failed, error is [%v]", userName, err)
	}

	if len(listUserAksResp.AccessKeys) > 0 {
		for _, accessKey := range listUserAksResp.AccessKeys {
			_, err = userClient.DeleteUserAccess(ctx,
				&api.DeleteUserAccessInput{UserName: userName, AccessKeyId: accessKey})
			if err != nil {
				return fmt.Errorf("delete user [%s] access key [%s] failed, "+
					"error is [%v]", userName, accessKey, err)
			}
		}
	}

	_, err = userClient.DeleteUser(ctx, &api.DeleteUserInput{UserName: userName})
	if err != nil {
		return fmt.Errorf("delete user [%s] failed, error is [%v]", userName, err)
	}

	return nil
}

func removeBucketPolicyStatement(ctx context.Context, accountSecret *coreV1.Secret, bucketName, userName string) error {
	s3Agent, err := agent.NewS3Agent(
		agent.Config{
			SecretKey: string(accountSecret.Data[sk]),
			AccessKey: string(accountSecret.Data[ak]),
			Endpoint:  string(accountSecret.Data[endpoint]),
			RootCA:    accountSecret.Data[rootCA],
		})
	if err != nil {
		return fmt.Errorf("new s3 agent failed, error is [%v]", err)
	}

	bp, err := s3Agent.GetBucketPolicy(ctx, bucketName,
		errors.NewExceptionalErrCodes(errors.ErrNoSuchBucket, errors.ErrNoSuchBucketPolicy))
	if err != nil {
		return fmt.Errorf("get bucket [%s] policy failed, error is [%v]", bucketName, err)
	}

	if bp == nil {
		log.AddContext(ctx).Infof("skip remove policy operation")
		return nil
	}

	editedBp := bp.RemoveStatement(userName)
	if reflect.DeepEqual(editedBp, bp) {
		log.AddContext(ctx).Infof("bucket [%s] policy has no statement about user [%s], "+
			"skip remove policy operation", bucketName, userName)
		return nil
	}

	// If policy have other statements, then using PutBucketPolicy to update.
	// If policy do not have any statement, then using DeleteBucketPolicy to delete policy statement,
	// because put policy with empty statement will fail.
	if len(editedBp.Statement) > 0 {
		err = s3Agent.PutBucketPolicy(ctx, bucketName, editedBp,
			errors.NewExceptionalErrCodes(errors.ErrNoSuchBucket, errors.ErrNoSuchBucketPolicy))
		if err != nil {
			return fmt.Errorf("remove bucket [%s] policy about user [%s] failed, "+
				"error is [%v]", bucketName, userName, err)
		}
	} else {
		log.AddContext(ctx).Infof("bucket [%s] policy statement is empty, delete bucket policy directly", bucketName)
		err = s3Agent.DeleteBucketPolicy(ctx, bucketName,
			errors.NewExceptionalErrCodes(errors.ErrNoSuchBucket, errors.ErrNoSuchBucketPolicy))
		if err != nil {
			return fmt.Errorf("delete bucket [%s] entire policy failed, error is [%v]", bucketName, err)
		}
	}

	return nil
}
