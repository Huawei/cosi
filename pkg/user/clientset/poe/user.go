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
	createUserAction = "CreateUser"
	getUserAction    = "GetUser"
	deleteUserAction = "DeleteUser"
)

// CreateUser is used to create user on backend
func (pec *Client) CreateUser(ctx context.Context, in *api.CreateUserInput) (*api.CreateUserOutput, error) {
	log.AddContext(ctx).Infof("start to create user, input is [%+v]", in)

	paramMap := make(map[string]string, 0)
	paramMap[actionKey] = createUserAction
	paramMap[userNameKey] = in.UserName
	body, err := pec.Call(ctx, paramMap)
	if err != nil {
		return nil, err
	}

	resp := &createUserResponse{}
	err = xml.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	log.AddContext(ctx).Infof("create user success, storage request id is [%s]", resp.ResponseMetadata.RequestId)
	return &api.CreateUserOutput{
		UserName: resp.CreateUserResult.User.UserName,
		UserID:   resp.CreateUserResult.User.UserID,
		Arn:      resp.CreateUserResult.User.Arn,
	}, nil
}

// GetUser is used to get user on backend.
func (pec *Client) GetUser(ctx context.Context, in *api.GetUserInput) (*api.GetUserOutput, error) {
	log.AddContext(ctx).Infof("start to get user, input is [%+v]", in)

	paramMap := make(map[string]string, 0)
	paramMap[actionKey] = getUserAction
	paramMap[userNameKey] = in.UserName
	body, err := pec.Call(ctx, paramMap)
	if err != nil {
		if errors.Is(err, errNoSuchUser) {
			msg := fmt.Sprintf("user [%s] not exist", in.UserName)
			log.AddContext(ctx).Infof(msg)
			return nil, nil
		}

		return nil, err
	}

	resp := &getUserResponse{}
	err = xml.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	log.AddContext(ctx).Infof("get user success, storage request id is [%s]", resp.ResponseMetadata.RequestId)
	return &api.GetUserOutput{
		UserName: resp.GetUserResult.User.UserName,
		UserID:   resp.GetUserResult.User.UserID,
		Arn:      resp.GetUserResult.User.Arn,
	}, nil
}

// DeleteUser is used to delete user on backend.
func (pec *Client) DeleteUser(ctx context.Context, in *api.DeleteUserInput) (*api.DeleteUserOutput, error) {
	log.AddContext(ctx).Infof("start to delete user, input is [%+v]", in)

	paramMap := make(map[string]string, 0)
	paramMap[actionKey] = deleteUserAction
	paramMap[userNameKey] = in.UserName
	body, err := pec.Call(ctx, paramMap)
	if err != nil {
		if errors.Is(err, errNoSuchUser) {
			msg := fmt.Sprintf("user [%s] is not exist", in.UserName)
			log.AddContext(ctx).Infof(msg)
			return &api.DeleteUserOutput{}, nil
		}

		return nil, err
	}

	resp := &deleteUserResponse{}
	err = xml.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	log.AddContext(ctx).Infof("delete user success, storage request id is [%s]", resp.ResponseMetadata.RequestId)
	return &api.DeleteUserOutput{}, nil
}
