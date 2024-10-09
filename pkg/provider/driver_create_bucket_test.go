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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
)

func Test_provisionerServer_DriverCreateBucket_Success(t *testing.T) {
	// arrange
	s3Agent := &agent.S3Agent{
		Client: &s3.S3{},
	}
	s := &provisionerServer{K8sClient: fake.NewSimpleClientset()}
	bucketName := "bucketName"
	acSecretName := "fake-secret"
	acSecretNameSpace := "huawei-cosi"
	req := &cosispec.DriverCreateBucketRequest{
		Name: bucketName,
		Parameters: map[string]string{
			"accountSecretName":      acSecretName,
			"accountSecretNamespace": acSecretNameSpace,
			"bucketACL":              "bucketACL",
			"bucketLocation":         "bucketLocation",
		},
	}
	want := &cosispec.DriverCreateBucketResponse{
		BucketId: assembleResourceId(acSecretNameSpace, acSecretName, bucketName),
	}

	// mock
	mocks := gomonkey.ApplyFunc(newS3Client,
		func(ctx context.Context, clientset kubernetes.Interface,
			parameters map[string]string) (*agent.S3Agent, error) {
			return s3Agent, nil
		}).ApplyMethod(reflect.TypeOf(s3Agent), "CreateBucket",
		func(_ *agent.S3Agent, ctx context.Context, bucketName, acl, location string) error {
			return nil
		})

	// act
	got, gotErr := s.DriverCreateBucket(context.TODO(), req)

	// assert
	if gotErr != nil || !reflect.DeepEqual(want, got) {
		t.Errorf("Test_provisionerServer_DriverCreateBucket_Success failed, got= [%v], want= [%v]"+
			"gotErr= [%v], wantErr= [%v]", got, want, gotErr, nil)
	}

	// cleanup
	t.Cleanup(func() {
		mocks.Reset()
	})
}

func Test_provisionerServer_DriverCreateBucket_Failed(t *testing.T) {
	// arrange
	s3Agent := &agent.S3Agent{
		Client: &s3.S3{},
	}
	s := &provisionerServer{K8sClient: fake.NewSimpleClientset()}
	bucketName := "bucketName"
	req := &cosispec.DriverCreateBucketRequest{
		Name: bucketName,
		Parameters: map[string]string{
			"accountSecretName":      "fake-secret",
			"accountSecretNamespace": "huawei-cosi",
			"bucketACL":              "bucketACL",
			"bucketLocation":         "bucketLocation",
		},
	}

	errCodeBucketAlreadyExistsErr := awserr.New(s3.ErrCodeBucketAlreadyExists, "s3 failed", nil)
	msg := fmt.Sprintf("create bucket [%s] failed, error is [%v]", bucketName, errCodeBucketAlreadyExistsErr)
	wantErr := status.Error(codes.Internal, msg)

	// mock
	mocks := gomonkey.ApplyFunc(newS3Client,
		func(ctx context.Context, clientset kubernetes.Interface,
			parameters map[string]string) (*agent.S3Agent, error) {
			return s3Agent, nil
		}).ApplyMethod(reflect.TypeOf(s3Agent), "CreateBucket",
		func(_ *agent.S3Agent, ctx context.Context, bucketName, acl, location string) error {
			return errCodeBucketAlreadyExistsErr
		})

	// act
	_, gotErr := s.DriverCreateBucket(context.TODO(), req)

	// assert
	if !reflect.DeepEqual(wantErr, gotErr) {
		t.Errorf("Test_provisionerServer_DriverCreateBucket_ErrCodeBucketAlreadyExists_Failed failed, "+
			"gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mocks.Reset()
	})
}

func Test_provisionerServer_newS3Client_Success(t *testing.T) {
	// arrange
	s3Agent := &agent.S3Agent{
		Client: &s3.S3{},
	}
	want := s3Agent
	s := &provisionerServer{K8sClient: fake.NewSimpleClientset()}
	parameters := map[string]string{
		"accountSecretName":      "fake-secret",
		"accountSecretNamespace": "huawei-cosi",
		"bucketACL":              "bucketACL",
		"bucketLocation":         "bucketLocation",
	}
	accountSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-secret",
			Namespace: "huawei-cosi"},
		Data: map[string][]byte{
			"accessKey": []byte("accessKey"),
			"secretKey": []byte("secretKey"),
			"endpoint":  []byte("endpoint"),
		},
	}

	// mock
	_, _ = s.K8sClient.CoreV1().Secrets("huawei-cosi").Create(context.TODO(),
		accountSecret, metav1.CreateOptions{})

	mock := gomonkey.ApplyFunc(agent.NewS3Agent,
		func(cfg agent.Config) (*agent.S3Agent, error) {
			return s3Agent, nil
		})

	// act
	got, gotErr := newS3Client(context.TODO(), s.K8sClient, parameters)

	// assert
	if gotErr != nil || !reflect.DeepEqual(want, got) {
		t.Errorf("Test_provisionerServer_DriverCreateBucket_Success failed, got= [%v], want= [%v]"+
			", gotErr[%v] ,wantErr nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		_ = s.K8sClient.CoreV1().Secrets("huawei-cosi").Delete(context.TODO(),
			"fake-secret", metav1.DeleteOptions{})
		mock.Reset()
	})
}
