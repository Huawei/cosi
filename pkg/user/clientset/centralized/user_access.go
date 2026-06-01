/*
 Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.

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

// Package centralized implements a centralized client for Huawei OceanStor object storage.
package centralized

import (
	"context"

	"github.com/huawei/cosi-driver/pkg/user/api"
)

const (
	userOwnType  = "1"
	userNotExist = 1092615946
)

// CreateUserAccess creates object storage access credentials (AK/SK) for a user
func (c *Client) CreateUserAccess(ctx context.Context,
	input *api.CreateUserAccessInput) (*api.CreateUserAccessOutput, error) {
	httpFn := func(ret interface{}) error {
		body := CreateAccessKeyRequest{
			OwnerName: input.UserName,
			OwnerType: userOwnType,
			VstoreId:  c.getVStoreID(),
		}
		return c.httpClient.POST(ctx, c.GetUrl("/OBJECT_AKSK"), body, ret)
	}

	resp, err := doRequest[CreateAccessKeyResponse](ctx, c, httpFn)
	if err != nil {
		return nil, err
	}

	data := resp.Data
	return &api.CreateUserAccessOutput{
		AccessKeyId:     data.AccessKey,
		SecretAccessKey: data.SecretKey,
	}, nil
}

// DeleteUserAccess deletes object storage access credentials
// Supports idempotent operation (no error if AK does not exist)
func (c *Client) DeleteUserAccess(ctx context.Context,
	input *api.DeleteUserAccessInput) (*api.DeleteUserAccessOutput, error) {
	httpFn := func(ret interface{}) error {
		queryParams := map[string]string{
			"id":        input.AccessKeyId,
			"ownerType": userOwnType,
			"vstoreId":  c.getVStoreID(),
		}
		return c.httpClient.DELETE(ctx, c.GetUrl("/OBJECT_AKSK"), queryParams, ret)
	}

	resp, err := doRequest[DeleteAccessKeyResponse](ctx, c, httpFn)
	if err != nil {
		if resp.Error.Code == userNotExist {
			return &api.DeleteUserAccessOutput{}, nil
		}
		return nil, err
	}
	return &api.DeleteUserAccessOutput{}, nil
}

// ListUserAccessKeys lists all access key IDs for a user (does not return secret keys)
func (c *Client) ListUserAccessKeys(ctx context.Context,
	input *api.ListUserAccessKeysInput) (*api.ListUserAccessKeysOutput, error) {
	httpFn := func(ret interface{}) error {
		query := map[string]string{
			"ownerName": input.UserName,
			"ownerType": userOwnType,
			"vstoreId":  c.getVStoreID(),
		}
		return c.httpClient.GET(ctx, c.GetUrl("/OBJECT_AKSK"), query, ret)
	}

	resp, err := doRequest[ListAccessKeysResponse](ctx, c, httpFn)
	if err != nil {
		if resp.Error.Code == userNotExist {
			return &api.ListUserAccessKeysOutput{}, nil
		}
		return nil, err
	}

	data := resp.Data
	accessKeys := make([]string, 0, len(data))
	for _, ak := range data {
		accessKeys = append(accessKeys, ak.Id)
	}

	return &api.ListUserAccessKeysOutput{AccessKeys: accessKeys}, nil
}
