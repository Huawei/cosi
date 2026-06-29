/*
 Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

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

package provider

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	cosiclientset "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned"

	"github.com/huawei/cosi-driver/pkg/utils"
)

func Test_NewProvisionerServer_Success(t *testing.T) {
	// arrange
	provisioner := "cosi.huawei.com"
	kubeConfigPath := ""

	// mock
	kubeConfig := &rest.Config{}

	patches := gomonkey.ApplyFunc(utils.GetKubeConfig, func(_ string, _ utils.ClientConfig) (*rest.Config, error) {
		return kubeConfig, nil
	})

	// act
	config := utils.ClientConfig{
		QPS:   utils.DefaultClientQPS,
		Burst: utils.DefaultClientBurst,
	}
	server, gotErr := NewProvisionerServer(provisioner, kubeConfigPath, config)

	// assert
	if gotErr != nil {
		t.Errorf("Test_NewProvisionerServer_Success failed, gotErr= [%v], wantErr= nil", gotErr)
	}

	if server == nil {
		t.Errorf("Test_NewProvisionerServer_Success failed, server= nil, want= not nil")
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_NewProvisionerServer_GetKubeConfigFailed(t *testing.T) {
	// arrange
	provisioner := "cosi.huawei.com"
	kubeConfigPath := ""
	wantErr := fmt.Errorf("get kube config failed, error is [demo error]")

	// mock
	patches := gomonkey.ApplyFunc(utils.GetKubeConfig, func(_ string, _ utils.ClientConfig) (*rest.Config, error) {
		return nil, wantErr
	})

	// act
	_, gotErr := NewProvisionerServer(provisioner, kubeConfigPath,
		utils.ClientConfig{QPS: utils.DefaultClientQPS, Burst: utils.DefaultClientBurst})

	// assert
	if gotErr == nil {
		t.Errorf("Test_NewProvisionerServer_GetKubeConfigFailed failed, gotErr= nil, wantErr= [%v]", wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_NewProvisionerServer_NewK8sClientFailed(t *testing.T) {
	// arrange
	provisioner := "cosi.huawei.com"
	kubeConfigPath := ""
	kubeConfig := &rest.Config{}
	wantErr := fmt.Errorf("new k8s client failed, error is [demo error]")

	// mock
	patches := gomonkey.ApplyFunc(utils.GetKubeConfig, func(_ string, _ utils.ClientConfig) (*rest.Config, error) {
		return kubeConfig, nil
	}).ApplyFunc(kubernetes.NewForConfig, func(_ *rest.Config) (*kubernetes.Clientset, error) {
		return nil, wantErr
	})

	// act
	_, gotErr := NewProvisionerServer(provisioner, kubeConfigPath,
		utils.ClientConfig{QPS: utils.DefaultClientQPS, Burst: utils.DefaultClientBurst})

	// assert
	if gotErr == nil {
		t.Errorf("Test_NewProvisionerServer_NewK8sClientFailed failed, gotErr= nil, wantErr= [%v]", wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_NewProvisionerServer_NewCosiClientFailed(t *testing.T) {
	// arrange
	provisioner := "cosi.huawei.com"
	kubeConfigPath := ""
	kubeConfig := &rest.Config{}
	k8sClient := &kubernetes.Clientset{}
	wantErr := fmt.Errorf("new cosi client failed, error is [demo error]")

	// mock
	patches := gomonkey.ApplyFunc(utils.GetKubeConfig, func(_ string, _ utils.ClientConfig) (*rest.Config, error) {
		return kubeConfig, nil
	}).ApplyFunc(kubernetes.NewForConfig, func(_ *rest.Config) (*kubernetes.Clientset, error) {
		return k8sClient, nil
	}).ApplyFunc(cosiclientset.NewForConfig, func(_ *rest.Config) (*cosiclientset.Clientset, error) {
		return nil, wantErr
	})

	// act
	_, gotErr := NewProvisionerServer(provisioner, kubeConfigPath,
		utils.ClientConfig{QPS: utils.DefaultClientQPS, Burst: utils.DefaultClientBurst})

	// assert
	if gotErr == nil {
		t.Errorf("Test_NewProvisionerServer_NewCosiClientFailed failed, gotErr= nil, wantErr= not nil")
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func TestNewProvisionerServerKubeConfigNil(t *testing.T) {
	// arrange
	provisioner := "cosi.huawei.com"
	kubeConfigPath := ""

	// mock: return nil kubeConfig without error, and mock NewForConfig to avoid panic
	patches := gomonkey.ApplyFunc(utils.GetKubeConfig, func(_ string, _ utils.ClientConfig) (*rest.Config, error) {
		return nil, nil
	}).ApplyFunc(kubernetes.NewForConfig, func(_ *rest.Config) (*kubernetes.Clientset, error) {
		return nil, fmt.Errorf("kube config is nil")
	})

	// act
	_, gotErr := NewProvisionerServer(provisioner, kubeConfigPath,
		utils.ClientConfig{QPS: utils.DefaultClientQPS, Burst: utils.DefaultClientBurst})

	// assert
	if gotErr == nil {
		t.Errorf("TestNewProvisionerServerKubeConfigNil failed, gotErr= nil, wantErr= not nil")
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}
