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
	"fmt"

	"k8s.io/client-go/kubernetes"
	cosiclientset "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/keylock"
)

type provisionerServer struct {
	Provisioner  string
	K8sClient    kubernetes.Interface
	BucketClient cosiclientset.Interface
	keyLock      *keylock.KeyMutexLock
}

var _ cosispec.ProvisionerServer = &provisionerServer{}

// NewProvisionerServer return a new cosi ProvisionerServer
func NewProvisionerServer(provisioner, kubeConfigPath string) (cosispec.ProvisionerServer, error) {
	kubeConfig, err := utils.GetKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("get kube config failed, error is [%v]", err)
	}

	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("new k8s client failed, error is [%v]", err)
	}

	cosiClient, err := cosiclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("new cosi client failed, error is [%v]", err)
	}

	return &provisionerServer{
		Provisioner:  provisioner,
		K8sClient:    k8sClient,
		BucketClient: cosiClient,
		keyLock:      keylock.NewKeyLock(keyLockSize),
	}, nil
}
