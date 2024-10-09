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

	cosispec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// NewDriver return a new cosi driver
func NewDriver(ctx context.Context, driverName, kubeConfigPath string) (cosispec.IdentityServer,
	cosispec.ProvisionerServer, error) {
	is, err := NewIdentityServer(driverName)
	if err != nil {
		log.AddContext(ctx).Errorf("failed to create identityServer server, error is [%v]", err)
		return nil, nil, err
	}

	ps, err := NewProvisionerServer(driverName, kubeConfigPath)
	if err != nil {
		log.AddContext(ctx).Errorf("failed to create provisioner server, error is [%v]", err)
		return nil, nil, err
	}

	return is, ps, nil
}
