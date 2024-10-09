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

package user

import (
	"fmt"

	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/user/clientset/poe"
)

const (
	// PoeType is the poe type of user client
	PoeType = "POE"
)

// Config contains the cfg information required for init UserClient
type Config struct {
	ClientType string
	AccessKey  string
	SecretKey  string
	Endpoint   string
	RootCA     []byte
}

// NewUserClient return a user client according to api type
func NewUserClient(config Config) (api.UserAPI, error) {
	switch config.ClientType {
	case PoeType:
		return poe.NewPoeClient(config.Endpoint, config.AccessKey, config.SecretKey, config.RootCA)
	default:
		return nil, fmt.Errorf("unknown user client type [%s]", config.ClientType)
	}
}
