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

// Package poe provides poe client and poe apis
package poe

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/huawei/cosi-driver/pkg/user/api"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

const (
	createUserAccess   = "CreateAccessKey"
	deleteUserAccess   = "DeleteAccessKey"
	listUserAccessKeys = "ListAccessKeys"
)

// CreateUserAccess is used to create user access on backend
func (pec *Client) CreateUserAccess(ctx context.Context,
	in *api.CreateUserAccessInput) (*api.CreateUserAccessOutput, error) {
	log.AddContext(ctx).Infof("start to create user access, input is [%+v]", in)

	paramMap := make(map[string]string, 0)
	paramMap[actionKey] = createUserAccess
	paramMap[userNameKey] = in.UserName
	body, err := pec.Call(ctx, paramMap)
	if err != nil {
		return nil, err
	}

	resp := &createUserAccessResponse{}
	err = xml.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	log.AddContext(ctx).Infof("create user access success, storage request id is [%s]", resp.ResponseMetadata.RequestId)
	return &api.CreateUserAccessOutput{
		AccessKeyId:     resp.CreateUserResult.AccessKey.AccessKeyId,
		SecretAccessKey: resp.CreateUserResult.AccessKey.SecretAccessKey,
	}, nil
}

// DeleteUserAccess is used to delete user access on backend
func (pec *Client) DeleteUserAccess(ctx context.Context,
	in *api.DeleteUserAccessInput) (*api.DeleteUserAccessOutput, error) {
	log.AddContext(ctx).Infof("start to delete user access, input is [%+v]", in)

	paramMap := make(map[string]string, 0)
	paramMap[actionKey] = deleteUserAccess
	paramMap[userNameKey] = in.UserName
	paramMap[accessKeyIdKey] = in.AccessKeyId
	body, err := pec.Call(ctx, paramMap)
	if err != nil {
		if errors.Is(err, errNoSuchUserAccess) {
			msg := fmt.Sprintf("user access [%s/%s] is not exist", in.UserName, in.AccessKeyId)
			log.AddContext(ctx).Infof(msg)
			return &api.DeleteUserAccessOutput{}, nil
		}

		return nil, err
	}

	resp := &deleteUserAccessResponse{}
	err = xml.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	log.AddContext(ctx).Infof("delete user access success, storage request id is [%s]", resp.ResponseMetadata.RequestId)
	return &api.DeleteUserAccessOutput{}, nil
}

// ListUserAccessKeys is used to list user access keys on backend
func (pec *Client) ListUserAccessKeys(ctx context.Context,
	in *api.ListUserAccessKeysInput) (*api.ListUserAccessKeysOutput, error) {
	log.AddContext(ctx).Infof("start to list user access keys, input is [%+v]", in)

	paramMap := make(map[string]string, 0)
	paramMap[actionKey] = listUserAccessKeys
	paramMap[userNameKey] = in.UserName
	body, err := pec.Call(ctx, paramMap)
	if err != nil {
		if errors.Is(err, errNoSuchUser) {
			msg := fmt.Sprintf("user [%s] is not exist", in.UserName)
			log.AddContext(ctx).Infof(msg)
			return &api.ListUserAccessKeysOutput{}, nil
		}

		return nil, err
	}

	resp := &listAccessKeysResponse{}
	err = xml.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	var accessKeys []string
	for _, m := range resp.ListAccessKeysResult.AccessKeyMetadata.Members {
		accessKeys = append(accessKeys, m.AccessKeyId)
	}

	log.AddContext(ctx).Infof("list user access keys success, storage request id is [%s]", resp.ResponseMetadata.RequestId)
	return &api.ListUserAccessKeysOutput{
		AccessKeys: accessKeys,
	}, nil
}
