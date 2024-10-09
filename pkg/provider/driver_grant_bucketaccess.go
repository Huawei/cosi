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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/s3/errors"
	"github.com/huawei/cosi-driver/pkg/s3/policy"
	"github.com/huawei/cosi-driver/pkg/user"
	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// DriverGrantBucketAccess is used to grants access to an account
func (s *provisionerServer) DriverGrantBucketAccess(ctx context.Context,
	req *cosispec.DriverGrantBucketAccessRequest) (*cosispec.DriverGrantBucketAccessResponse, error) {
	defer utils.RecoverPanic(ctx)

	// When processing the same bucket, it needs to add mutex lock.
	// Otherwise, policies will be overwritten abnormally in concurrent scenarios.
	s.keyLock.Lock(req.GetBucketId())
	defer s.keyLock.Unlock(req.GetBucketId())

	log.AddContext(ctx).Infof("handle DriverGrantBucketAccess request, request is [%+v]", req)

	err := checkDriverGrantBucketAccessRequest(req)
	if err != nil {
		msg := fmt.Sprintf("check DriverGrantBucketAccessRequest failed, error is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	bucketIdData, bcAccountSecret, err := fetchDataFromResourceId(req.GetBucketId(), s.K8sClient)
	if err != nil {
		msg := fmt.Sprintf("fetch data from resourceId [%s] failed, error is [%v]", req.GetBucketId(), err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	err = checkBucketExistence(ctx, bcAccountSecret, bucketIdData.resourceName)
	if err != nil {
		msg := fmt.Sprintf("check bucket existence failed, error is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	bacAccountSecret, err := s.K8sClient.CoreV1().Secrets(req.Parameters[accountSecretNamespace]).
		Get(ctx, req.Parameters[accountSecretName], metaV1.GetOptions{})
	if err != nil {
		msg := fmt.Sprintf("failed to get account secret from paramters, error is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	userData, err := registerUser(ctx, req, bacAccountSecret)
	if err != nil {
		msg := fmt.Sprintf("register user failed, error is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	err = setBucketPolicy(ctx, req, bcAccountSecret, userData, bucketIdData.resourceName)
	if err != nil {
		msg := fmt.Sprintf("set bucket policy about user failed, error is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	log.AddContext(ctx).Infof("handle DriverGrantBucketAccess request successfully")
	return &cosispec.DriverGrantBucketAccessResponse{
		AccountId:   assembleResourceId(bacAccountSecret.Namespace, bacAccountSecret.Name, req.GetName()),
		Credentials: buildCredentials(bcAccountSecret, userData),
	}, nil
}

func checkDriverGrantBucketAccessRequest(req *cosispec.DriverGrantBucketAccessRequest) error {
	if req.GetBucketId() == "" {
		return fmt.Errorf("empty bucket id")
	}

	if req.GetName() == "" {
		return fmt.Errorf("empty user name")
	}

	if req.GetAuthenticationType() == cosispec.AuthenticationType_IAM {
		return fmt.Errorf("IAM authentication type not implemented")
	}

	if req.GetAuthenticationType() != cosispec.AuthenticationType_Key {
		return fmt.Errorf("unknown authentication type")
	}

	// Req parameters is passed down from bucketAccessClass parameters
	_, exist := req.Parameters[accountSecretName]
	if !exist {
		return fmt.Errorf("account secret name value is empty")
	}

	_, exist = req.Parameters[accountSecretNamespace]
	if !exist {
		return fmt.Errorf("account secret namespace value is empty")
	}

	// BucketPolicyModel is optional
	policModel, exist := req.Parameters[bucketPolicyModel]
	if !exist {
		return nil
	}

	if policModel != bucketPolicyModelRW && policModel != bucketPolicyModelRO {
		return fmt.Errorf("invalid bucketPolicy model [%s]", policModel)
	}

	return nil
}

type userInfo struct {
	userArn         string
	accessKeyId     string
	accessSecretKey string
}

func registerUser(ctx context.Context, req *cosispec.DriverGrantBucketAccessRequest,
	bacAccountSecret *coreV1.Secret) (*userInfo, error) {
	userName := req.GetName()

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
		return nil, fmt.Errorf("new user client failed, error is [%v]", err)
	}

	var userArn string
	getUserResp, err := userClient.GetUser(ctx, &api.GetUserInput{UserName: userName})
	if err != nil {
		return nil, fmt.Errorf("get user failed, error is [%v]", err)
	}

	// If user not exist, then create one.
	if getUserResp != nil {
		userArn = getUserResp.Arn
	} else {
		createUserResp, err := userClient.CreateUser(ctx, &api.CreateUserInput{UserName: userName})
		if err != nil {
			return nil, fmt.Errorf("create user [%s] failed, error is [%v]", userName, err)
		}

		userArn = createUserResp.Arn
	}

	// If user access lost, a new one must be issued.
	accessResp, err := userClient.CreateUserAccess(ctx, &api.CreateUserAccessInput{UserName: userName})
	if err != nil {
		return nil, fmt.Errorf("create user [%s] access failed, error is [%v]", userName, err)
	}

	return &userInfo{
		userArn:         userArn,
		accessKeyId:     accessResp.AccessKeyId,
		accessSecretKey: accessResp.SecretAccessKey,
	}, nil
}

func setBucketPolicy(ctx context.Context, req *cosispec.DriverGrantBucketAccessRequest,
	bcAccountSecret *coreV1.Secret, userData *userInfo, bucketName string) error {
	s3Agent, err := agent.NewS3Agent(
		agent.Config{
			SecretKey: string(bcAccountSecret.Data[sk]),
			AccessKey: string(bcAccountSecret.Data[ak]),
			Endpoint:  string(bcAccountSecret.Data[endpoint]),
			RootCA:    bcAccountSecret.Data[rootCA],
		})
	if err != nil {
		return fmt.Errorf("new s3 agent failed, error is [%v]", err)
	}

	bp, err := s3Agent.GetBucketPolicy(ctx, bucketName, errors.NewExceptionalErrCodes(errors.ErrNoSuchBucketPolicy))
	if err != nil {
		return fmt.Errorf("get bucket [%s] policy failed, error is [%v]", bucketName, err)
	}

	// Default action is RW model
	model := req.Parameters[bucketPolicyModel]
	actions := policy.AllowedReadWriteActions
	if model == bucketPolicyModelRO {
		actions = policy.AllowedReadActions
	}

	userName := req.GetName()
	statement := policy.NewStatementBuilder().
		WithSID(userName).
		WithEffect(policy.EffectAllow).
		WithPrincipals(userData.userArn).
		WithActions(actions).
		WithResources(bucketName).
		WithSubResources(bucketName).
		Build()
	if bp == nil {
		bp = policy.NewBucketPolicy(*statement)
	} else {
		bp = bp.ModifyStatement(*statement)
	}

	err = s3Agent.PutBucketPolicy(ctx, bucketName, bp, errors.EmptyExceptionalErrCodes)
	if err != nil {
		return fmt.Errorf("put bucket [%s] policy about user [%s] failed, "+
			"error is [%v]", bucketName, userName, err)
	}

	return nil
}

func buildCredentials(bcAccountSecret *coreV1.Secret, userData *userInfo) map[string]*cosispec.CredentialDetails {
	cred := &cosispec.CredentialDetails{
		Secrets: map[string]string{
			accessAk: userData.accessKeyId,
			accessSk: userData.accessSecretKey,
			endpoint: string(bcAccountSecret.Data[endpoint]),
		},
	}
	credDetails := make(map[string]*cosispec.CredentialDetails)
	credDetails[s3Protocol] = cred

	return credDetails
}

func checkBucketExistence(ctx context.Context, bcAccountSecret *coreV1.Secret, bucketName string) error {
	s3Agent, err := agent.NewS3Agent(
		agent.Config{
			SecretKey: string(bcAccountSecret.Data[sk]),
			AccessKey: string(bcAccountSecret.Data[ak]),
			Endpoint:  string(bcAccountSecret.Data[endpoint]),
			RootCA:    bcAccountSecret.Data[rootCA],
		})
	if err != nil {
		return fmt.Errorf("new s3 agent failed, error is [%v]", err)
	}

	return s3Agent.CheckBucketExist(ctx, bucketName)
}
