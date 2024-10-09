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

// Package api defines the user related interface
package api

// CreateUserInput define CreateUser interface input
type CreateUserInput struct {
	UserName string
}

// CreateUserOutput define CreateUser interface output
type CreateUserOutput struct {
	UserName string
	UserID   string
	Arn      string
}

// GetUserInput define GetUser interface input
type GetUserInput struct {
	UserName string
}

// GetUserOutput define GetUser interface output
type GetUserOutput struct {
	UserName string
	UserID   string
	Arn      string
}

// DeleteUserInput define DeleteUser interface input
type DeleteUserInput struct {
	UserName string
}

// DeleteUserOutput define DeleteUser interface output
type DeleteUserOutput struct {
	_ struct{}
}

// CreateUserAccessInput define CreateUserAccess interface input
type CreateUserAccessInput struct {
	UserName string
}

// CreateUserAccessOutput define CreateUserAccess interface output
type CreateUserAccessOutput struct {
	AccessKeyId     string
	SecretAccessKey string
}

// DeleteUserAccessInput define DeleteUserAccess interface input
type DeleteUserAccessInput struct {
	UserName    string
	AccessKeyId string
}

// DeleteUserAccessOutput define DeleteUserAccess interface output
type DeleteUserAccessOutput struct {
	_ struct{}
}

// ListUserAccessKeysInput define ListUserAccessKeys interface input
type ListUserAccessKeysInput struct {
	UserName string
}

// ListUserAccessKeysOutput define ListUserAccessKeys interface output
type ListUserAccessKeysOutput struct {
	AccessKeys []string
}
