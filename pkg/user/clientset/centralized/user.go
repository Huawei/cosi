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
	"fmt"

	"github.com/huawei/cosi-driver/pkg/user/api"
)

const (
	localUserType = "0"
	// userARNFormat is the ARN format for IAM users(arn:aws:iam::{accountId}:user/{userName})
	userARNFormat = "arn:aws:iam::%s:user/%s"
)

// CreateUser creates an object storage user
func (c *Client) CreateUser(ctx context.Context, input *api.CreateUserInput) (*api.CreateUserOutput, error) {
	httpFn := func(ret interface{}) error {
		body := CreateUserRequest{
			Name:     input.UserName,
			UserType: localUserType,
			VstoreId: c.getVStoreID(),
		}
		return c.httpClient.POST(ctx, c.GetUrl("/OBJECT_USER"), body, ret)
	}

	resp, err := doRequest[CreateUserResponse](ctx, c, httpFn)
	if err != nil {
		return nil, err
	}

	data := resp.Data
	arn := fmt.Sprintf(userARNFormat, c.getVStoreID(), data.Name)
	return &api.CreateUserOutput{
		UserName: data.Name,
		UserID:   data.Id,
		Arn:      arn,
	}, nil
}

// GetUser queries object user information
// Returns empty result if user does not exist (not an error)
func (c *Client) GetUser(ctx context.Context, input *api.GetUserInput) (*api.GetUserOutput, error) {
	httpFn := func(ret interface{}) error {
		query := map[string]string{
			"name":     input.UserName,
			"vstoreId": c.getVStoreID(),
		}
		return c.httpClient.GET(ctx, c.GetUrl("/OBJECT_USER"), query, ret)
	}

	resp, err := doRequest[GetUserResponse](ctx, c, httpFn)
	if err != nil {
		return nil, err
	}

	if resp.Data.Id == "" {
		return nil, nil
	}

	arn := fmt.Sprintf(userARNFormat, c.getVStoreID(), resp.Data.Name)
	return &api.GetUserOutput{
		UserName: resp.Data.Name,
		UserID:   resp.Data.Id,
		Arn:      arn,
	}, nil
}

// DeleteUser deletes an object user
// Supports idempotent operation (no error if user does not exist)
func (c *Client) DeleteUser(ctx context.Context, input *api.DeleteUserInput) (*api.DeleteUserOutput, error) {
	httpFn := func(ret interface{}) error {
		queryParams := map[string]string{
			"name":     input.UserName,
			"vstoreId": c.getVStoreID(),
		}
		return c.httpClient.DELETE(ctx, c.GetUrl("/OBJECT_USER"), queryParams, ret)
	}

	resp, err := doRequest[DeleteUserResponse](ctx, c, httpFn)
	if err != nil {
		if resp.Error.Code == userNotExist {
			return &api.DeleteUserOutput{}, nil
		}
		return nil, err
	}

	return &api.DeleteUserOutput{}, nil
}
