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
	fakeK8sClient "k8s.io/client-go/kubernetes/fake"
	fakeBucketClient "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned/fake"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/s3/agent"
)

func Test_provisionerServer_DriverDeleteBucket_Success(t *testing.T) {
	// arrange
	s3Agent := &agent.S3Agent{
		Client: &s3.S3{},
	}
	s := &provisionerServer{
		K8sClient:    fakeK8sClient.NewSimpleClientset(),
		BucketClient: fakeBucketClient.NewSimpleClientset(),
	}
	acSecretName := "fake-secret"
	acSecretNameSpace := "huawei-cosi"
	ctx := context.TODO()
	bucketName := "bucket-name"
	bucketId := assembleResourceId(acSecretNameSpace, acSecretName, bucketName)
	req := &cosispec.DriverDeleteBucketRequest{
		BucketId: bucketId,
	}

	want := &cosispec.DriverDeleteBucketResponse{}
	accountSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: acSecretName, Namespace: acSecretNameSpace},
		Data: map[string][]byte{
			"accessKey": []byte("accessKey"),
			"secretKey": []byte("secretKey"),
			"endpoint":  []byte("endpoint"),
		},
	}

	// mock
	_, _ = s.K8sClient.CoreV1().Secrets("huawei-cosi").Create(ctx, accountSecret, metav1.CreateOptions{})
	mocks := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(cfg agent.Config) (*agent.S3Agent, error) {
			return s3Agent, nil
		}).ApplyMethod(reflect.TypeOf(s3Agent), "DeleteBucket",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) error {
			return nil
		})

	// act
	got, gotErr := s.DriverDeleteBucket(ctx, req)

	// assert
	if gotErr != nil || !reflect.DeepEqual(want, got) {
		t.Errorf("Test_provisionerServer_DriverDeleteBucket_Success failed, got= [%v], want= [%v]"+
			"gotErr= [%v], wantErr= [%v]", got, want, gotErr, nil)
	}

	// cleanup
	t.Cleanup(func() {
		_ = s.K8sClient.CoreV1().Secrets("huawei-cosi").Delete(context.TODO(), "fake-secret", metav1.DeleteOptions{})
		mocks.Reset()
	})
}

func Test_provisionerServer_DriverDeleteBucket_NotExit_Success(t *testing.T) {
	// arrange
	s3Agent := &agent.S3Agent{
		Client: &s3.S3{},
	}
	s := &provisionerServer{
		K8sClient:    fakeK8sClient.NewSimpleClientset(),
		BucketClient: fakeBucketClient.NewSimpleClientset(),
	}
	want := &cosispec.DriverDeleteBucketResponse{}
	ctx := context.TODO()
	acSecretName := "fake-secret"
	acSecretNameSpace := "huawei-cosi"
	bucketName := "bucket-name"
	bucketId := assembleResourceId(acSecretNameSpace, acSecretName, bucketName)
	req := &cosispec.DriverDeleteBucketRequest{
		BucketId: bucketId,
	}
	accountSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "fake-secret", Namespace: "huawei-cosi"},
		Data: map[string][]byte{
			"accessKey": []byte("accessKey"),
			"secretKey": []byte("secretKey"),
			"endpoint":  []byte("endpoint"),
		},
	}
	ns := "huawei-cosi"

	// mocks
	_, _ = s.K8sClient.CoreV1().Secrets(ns).Create(ctx, accountSecret, metav1.CreateOptions{})
	mocks := gomonkey.
		ApplyFunc(agent.NewS3Agent, func(cfg agent.Config) (*agent.S3Agent, error) {
			return s3Agent, nil
		}).ApplyMethod(reflect.TypeOf(s3Agent), "DeleteBucket",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) error {
			return nil
		})

	// act
	got, gotErr := s.DriverDeleteBucket(ctx, req)

	// assert
	if gotErr != nil || !reflect.DeepEqual(want, got) {
		t.Errorf("Test_provisionerServer_DriverDeleteBucket_NotExit_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= [%v]", got, want, gotErr, nil)
	}

	// cleanup
	t.Cleanup(func() {
		_ = s.K8sClient.CoreV1().Secrets(ns).Delete(ctx, accountSecret.Name, metav1.DeleteOptions{})
		mocks.Reset()
	})
}

func Test_provisionerServer_DriverDeleteBucket_Failed(t *testing.T) {
	// arrange
	s3Agent := &agent.S3Agent{
		Client: &s3.S3{},
	}
	s := &provisionerServer{
		K8sClient:    fakeK8sClient.NewSimpleClientset(),
		BucketClient: fakeBucketClient.NewSimpleClientset(),
	}
	ctx := context.TODO()
	acSecretName := "fake-secret"
	acSecretNameSpace := "huawei-cosi"
	bucketName := "bucket-name"
	bucketId := assembleResourceId(acSecretNameSpace, acSecretName, bucketName)
	req := &cosispec.DriverDeleteBucketRequest{
		BucketId: bucketId,
	}

	deleteFailErr := awserr.New("code", "msg", fmt.Errorf("fail to delete bucket"))
	msg := fmt.Sprintf("failed to delete bucket [%s], err is [%v]", bucketName, deleteFailErr)
	wantErr := status.Error(codes.Internal, msg)
	accountSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "fake-secret", Namespace: "huawei-cosi"},
		Data: map[string][]byte{
			"accessKey": []byte("accessKey"),
			"secretKey": []byte("secretKey"),
			"endpoint":  []byte("endpoint"),
		},
	}

	// mocks
	_, _ = s.K8sClient.CoreV1().Secrets("huawei-cosi").Create(ctx, accountSecret, metav1.CreateOptions{})
	mocks := gomonkey.ApplyFunc(agent.NewS3Agent,
		func(cfg agent.Config) (*agent.S3Agent, error) {
			return s3Agent, nil
		}).ApplyMethod(reflect.TypeOf(s3Agent), "DeleteBucket",
		func(_ *agent.S3Agent, ctx context.Context, bucketName string) error {
			return deleteFailErr
		})

	// act
	_, gotErr := s.DriverDeleteBucket(ctx, req)

	// assert
	if gotErr == nil || !reflect.DeepEqual(wantErr, gotErr) {
		t.Errorf("Test_provisionerServer_DriverDeleteBucket_Failed failed, gotErr= [%v], wantErr= [%v]",
			gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		_ = s.K8sClient.CoreV1().Secrets("huawei-cosi").Delete(ctx, "fake-secret", metav1.DeleteOptions{})
		mocks.Reset()
	})
}
