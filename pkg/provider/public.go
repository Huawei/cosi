/*
 Copyright (c) Huawei Technologies Co., Ltd. 2024-2026. All rights reserved.

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
	"strconv"
	"strings"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/huawei/cosi-driver/pkg/user"
	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/utils/log"
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

// buildClientFromSecret creates a UserAPI client based on the credentials
// found in the provided Secret.
//
// Client type selection logic:
// - If username AND password are present → Centralized client
// - If accessKey AND secretKey are present → POE client
// - If both are present → Centralized takes precedence
// - If neither pair is complete → returns error
func buildClientFromSecret(ctx context.Context, secret *coreV1.Secret) (api.UserAPI, error) {
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("invalid secret: data is nil")
	}

	cfg := user.Config{
		AccessKey: string(secret.Data[ak]),
		SecretKey: string(secret.Data[sk]),
		Endpoint:  string(secret.Data[endpoint]),
		RootCA:    secret.Data[rootCA],
		Username:  string(secret.Data[username]),
		Password:  string(secret.Data[password]),
	}

	// Parse MaxConcurrent from secret if provided
	if maxConcurrentValue := secret.Data[maxConcurrent]; len(maxConcurrentValue) > 0 {
		if value, err := strconv.Atoi(string(maxConcurrentValue)); err == nil && value > 0 {
			cfg.MaxConcurrent = value
		} else {
			log.AddContext(ctx).Warningf(`invalid maxConcurrent value [%s], using default`, string(maxConcurrentValue))
		}
	}

	// Centralized client takes precedence
	if cfg.Username != "" && cfg.Password != "" {
		cfg.ClientType = user.CentralizedType
		return user.NewUserClient(cfg)
	}

	// Check if POE credentials are complete
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		cfg.ClientType = user.PoeType
		return user.NewUserClient(cfg)
	}

	// Neither credential pair is complete
	return nil, fmt.Errorf("incomplete credentials")
}
