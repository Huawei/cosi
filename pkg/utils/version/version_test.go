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

// Package version used to set and clean the service version
package version

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/huawei/cosi-driver/pkg/utils"
)

func Test_InitVersionConfigMap_Success(t *testing.T) {
	// arrange
	fakeK8sClient := fake.NewSimpleClientset()
	containerName := "cosi-test"
	version := "v1.0.0"
	namespace := "cosi-test"

	// act
	gotErr := initVersionConfigMap(fakeK8sClient, containerName, version, namespace)

	// assert
	if gotErr != nil {
		t.Errorf("Test_InitVersionConfigMap_Success failed, gotErr= [%v], wantErr= [%v]", gotErr, nil)
	}
}

func Test_InitVersionConfigMap_Create_Failed(t *testing.T) {
	// arrange
	fakeK8sClient := fake.NewSimpleClientset()
	containerName := "cosi-test"
	version := "v1.0.0"
	namespace := "cosi-test"
	errMsg := "create version cm failed"
	wantErr := errors.New(errMsg)

	// mock
	p := gomonkey.ApplyFunc(createVersionConfigMap,
		func(k8sClient kubernetes.Interface, containerName, version, namespace string) error {
			return wantErr
		})

	// act
	gotErr := initVersionConfigMap(fakeK8sClient, containerName, version, namespace)

	// assert
	if !reflect.DeepEqual(wantErr, gotErr) {
		t.Errorf("Test_InitVersionConfigMap_Create_Failed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestRegisterVersionGetKubeConfigFailed(t *testing.T) {
	// arrange
	containerName := "cosi-test"
	version := "v1.0.0"
	kubeConfigPath := ""

	// mock
	p := gomonkey.ApplyFunc(utils.GetKubeConfig, func(_ string, _ utils.ClientConfig) (*rest.Config, error) {
		return nil, errors.New("get kube config failed")
	})

	// act
	gotErr := RegisterVersion(containerName, version, kubeConfigPath)

	// assert
	if gotErr == nil {
		t.Errorf("TestRegisterVersionGetKubeConfigFailed failed, gotErr= nil, wantErr= not nil")
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestRegisterVersionNewK8sClientFailed(t *testing.T) {
	// arrange
	containerName := "cosi-test"
	version := "v1.0.0"
	kubeConfigPath := ""

	// mock
	p := gomonkey.ApplyFuncSeq(utils.GetKubeConfig, []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, nil}},
	})
	p.ApplyFuncSeq(kubernetes.NewForConfig, []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, errors.New("new k8s client failed")}},
	})

	// act
	gotErr := RegisterVersion(containerName, version, kubeConfigPath)

	// assert
	if gotErr == nil {
		t.Errorf("TestRegisterVersionNewK8sClientFailed failed, gotErr= nil, wantErr= not nil")
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestRegisterVersionWithCustomNamespace(t *testing.T) {
	// arrange
	customNamespace := "custom-namespace"

	// set custom namespace
	err := os.Setenv(envNameSpace, customNamespace)
	if err != nil {
		t.Fatalf("TestRegisterVersionWithCustomNamespace setup failed, os.Setenv error= [%v]", err)
	}
	defer os.Unsetenv(envNameSpace)

	// assert: verify the namespace environment variable is set correctly
	gotNamespace := os.Getenv(envNameSpace)
	if gotNamespace != customNamespace {
		t.Errorf("TestRegisterVersionWithCustomNamespace failed, gotNamespace= [%v], wantNamespace= [%v]", gotNamespace,
			customNamespace)
	}
}
