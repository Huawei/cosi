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
	"strings"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	assembleSymbol    = "/"
	disassembleLength = 3
)

func assembleResourceId(accountSecretNameSpace, accountSecretName, resourceName string) string {
	b := strings.Builder{}
	b.WriteString(accountSecretNameSpace)
	b.WriteString(assembleSymbol)
	b.WriteString(accountSecretName)
	b.WriteString(assembleSymbol)
	b.WriteString(resourceName)
	return b.String()
}

type resourceIdInfo struct {
	acSecretNameSpace string
	acSecretName      string
	resourceName      string
}

func disassembleResourceId(resourceId string) (*resourceIdInfo, error) {
	list := strings.Split(resourceId, assembleSymbol)
	if len(list) != disassembleLength {
		return nil, fmt.Errorf("invalid format of input [%s]", resourceId)
	}

	acSecretNameSpace := list[0]
	acSecretName := list[1]
	resourceName := list[2]

	if acSecretNameSpace == "" || acSecretName == "" || resourceName == "" {
		return nil, fmt.Errorf("invalid value of input [%s]", resourceId)
	}

	return &resourceIdInfo{
		acSecretNameSpace: acSecretNameSpace,
		acSecretName:      acSecretName,
		resourceName:      resourceName,
	}, nil
}

func fetchDataFromResourceId(resourceId string, client kubernetes.Interface) (*resourceIdInfo, *coreV1.Secret, error) {
	resourceIdData, err := disassembleResourceId(resourceId)
	if err != nil {
		return nil, nil, fmt.Errorf("disassemble resourceId failed, error is [%v]", err)
	}

	secret, err := client.CoreV1().Secrets(resourceIdData.acSecretNameSpace).
		Get(context.TODO(), resourceIdData.acSecretName, metaV1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("get account secret [%s/%s] failed, error is [%v]",
			resourceIdData.acSecretNameSpace, resourceIdData.acSecretName, err)
	}

	return resourceIdData, secret, nil
}
