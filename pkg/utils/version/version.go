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
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

var mutex sync.Mutex

const (
	defaultNamespace = "huawei-cosi"
	envNameSpace     = "env-namepsace"
)

// RegisterVersion used for register container version to configmap
func RegisterVersion(containerName, version, kubeConfigPath string) error {
	namespace := os.Getenv(envNameSpace)
	if namespace == "" {
		namespace = defaultNamespace
	}

	kubeConfig, err := utils.GetKubeConfig(kubeConfigPath)
	if err != nil {
		return fmt.Errorf("get kube config failed, error is [%v]", err)
	}
	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("init k8s client failed, error is [%v]", err)
	}

	err = initVersionConfigMap(k8sClient, containerName, version, namespace)
	if err != nil {
		return fmt.Errorf("init version file failed, error is [%v]", err)
	}

	return nil
}

func initVersionConfigMap(k8sClient kubernetes.Interface, containerName, version, namespace string) error {
	mutex.Lock()
	defer mutex.Unlock()
	log.Infof("Init version is [%s], osArch is [%s]", version, OSArch)

	cmName := versionConfigMapName
	cm, err := k8sClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metaV1.GetOptions{})
	if apiErrors.IsNotFound(err) {
		err = createVersionConfigMap(k8sClient, containerName, version, namespace, cmName)
		if err != nil {
			return err
		}
	} else if err != nil {
		errMsg := fmt.Sprintf("get configMap [%s] failed, error is [%v]", cmName, err)
		return errors.New(errMsg)
	}

	for true {
		cm, err = k8sClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metaV1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get configMap [%s] failed, error is [%v]", cmName, err)
		}

		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data[containerName] = version
		cm, err = k8sClient.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, metaV1.UpdateOptions{})
		if err != nil && apiErrors.IsConflict(err) {
			time.Sleep(time.Second)
			continue
		} else if err != nil {
			errMsg := fmt.Sprintf("update configMap [%s] failed, error is [%v]", cmName, err)
			return errors.New(errMsg)
		}

		break
	}

	return nil
}

func createVersionConfigMap(k8sClient kubernetes.Interface, containerName, version, namespace, cmName string) error {
	cm := &coreV1.ConfigMap{}
	cm.Name = cmName
	cm.Namespace = namespace
	cm.Data = make(map[string]string)
	cm.Data[containerName] = version

	_, err := k8sClient.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, metaV1.CreateOptions{})
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		errMsg := fmt.Sprintf("create configMap [%s] failed, error is [%v]", cmName, err)
		return errors.New(errMsg)
	}

	return nil
}
