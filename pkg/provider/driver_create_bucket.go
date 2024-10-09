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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// DriverCreateBucket is used to create the bucket in the backend
func (s *provisionerServer) DriverCreateBucket(ctx context.Context,
	req *cosispec.DriverCreateBucketRequest) (*cosispec.DriverCreateBucketResponse, error) {
	defer utils.RecoverPanic(ctx)
	log.AddContext(ctx).Infof("handle DriverCreateBucket request, request is [%+v]", req)

	err := checkDriverCreateBucketRequest(req)
	if err != nil {
		msg := fmt.Sprintf("check DriverCreateBucket failed, error is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	bucketName := req.GetName()
	parameters := req.GetParameters()
	s3Client, err := newS3Client(ctx, s.K8sClient, parameters)
	if err != nil {
		msg := fmt.Sprintf("new s3 client failed, err is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	err = s3Client.CreateBucket(ctx, bucketName, parameters[bucketACL], parameters[bucketLocation])
	if err != nil {
		msg := fmt.Sprintf("create bucket [%s] failed, error is [%v]", bucketName, err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	log.AddContext(ctx).Infof("handle DriverCreateBucket request successfully")
	return &cosispec.DriverCreateBucketResponse{
		BucketId: assembleResourceId(parameters[accountSecretNamespace], parameters[accountSecretName], bucketName),
	}, nil
}

func newS3Client(ctx context.Context, clientset kubernetes.Interface,
	parameters map[string]string) (*agent.S3Agent, error) {
	accountSecret, err := clientset.CoreV1().Secrets(parameters[accountSecretNamespace]).
		Get(ctx, parameters[accountSecretName], metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get account secret, error is [%v]", err)
	}

	s3Agent, err := agent.NewS3Agent(
		agent.Config{
			SecretKey: string(accountSecret.Data[sk]),
			AccessKey: string(accountSecret.Data[ak]),
			Endpoint:  string(accountSecret.Data[endpoint]),
			RootCA:    accountSecret.Data[rootCA],
		})
	if err != nil {
		return nil, fmt.Errorf("new s3 agent failed, error is [%v]", err)
	}

	return s3Agent, nil
}

func checkDriverCreateBucketRequest(req *cosispec.DriverCreateBucketRequest) error {
	if req.GetName() == "" {
		return fmt.Errorf("empty bucket name")
	}

	if req.GetParameters() == nil {
		return fmt.Errorf("empty bucket Parameters")
	}

	parameters := req.GetParameters()
	if parameters[accountSecretName] == "" {
		return fmt.Errorf("accountSecretName value is empty")
	}

	if parameters[accountSecretNamespace] == "" {
		return fmt.Errorf("accountSecretNamespace value is empty")
	}
	return nil
}
