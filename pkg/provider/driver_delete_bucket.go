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
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// DriverDeleteBucket is used to delete the bucket in the backend
func (s *provisionerServer) DriverDeleteBucket(ctx context.Context,
	req *cosispec.DriverDeleteBucketRequest) (*cosispec.DriverDeleteBucketResponse, error) {
	defer utils.RecoverPanic(ctx)
	log.AddContext(ctx).Infof("handle DriverDeleteBucket request, request is [%+v]", req.BucketId)

	bucketIdData, bcAccountSecret, err := fetchDataFromResourceId(req.GetBucketId(), s.K8sClient)
	if err != nil {
		msg := fmt.Sprintf("fetch data from resourceId [%s] failed, error is [%v]", req.GetBucketId(), err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	s3Agent, err := agent.NewS3Agent(
		agent.Config{
			SecretKey: string(bcAccountSecret.Data[sk]),
			AccessKey: string(bcAccountSecret.Data[ak]),
			Endpoint:  string(bcAccountSecret.Data[endpoint]),
			RootCA:    bcAccountSecret.Data[rootCA],
		})
	if err != nil {
		msg := fmt.Sprintf("new s3 client failed, err is [%v]", err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	err = s3Agent.DeleteBucket(ctx, bucketIdData.resourceName)
	if err != nil {
		msg := fmt.Sprintf("failed to delete bucket [%s], err is [%v]", bucketIdData.resourceName, err)
		log.AddContext(ctx).Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	log.AddContext(ctx).Infof("handle DriverDeleteBucket request successfully")
	return &cosispec.DriverDeleteBucketResponse{}, nil
}
