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

import "fmt"

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code        int64  `json:"code"`
	Description string `json:"description"`
}

// HttpResponse represents a generic HTTP response with data and error information
type HttpResponse[T any] struct {
	Data  T         `json:"data,omitempty"`
	Error ErrorInfo `json:"error"`
}

// IsSuccess checks if the response is successful
func (r HttpResponse[T]) IsSuccess() bool {
	return r.Error.Code == 0
}

// LogString returns the string for logging, sensitive fields are omitted
func (r HttpResponse[T]) LogString() string {
	if marshaler, ok := any(r.Data).(LogMarshaler); ok {
		return marshaler.LogString()
	}
	return emptyFlag
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Scope    string `json:"scope"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	UserId     string `json:"userid"`
	RoleId     string `json:"roleId"`
	DeviceId   string `json:"deviceid"`
	Username   string `json:"username"`
	UserScope  string `json:"userscope"`
	VStoreId   string `json:"vstoreId"`
	VStoreName string `json:"vstoreName"`
	IbaseToken string `json:"iBaseToken"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct{}

// LogoutResponse represents a logout response
type LogoutResponse struct{}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Name            string `json:"name"`
	UserType        string `json:"userType,omitempty"`
	UserDescription string `json:"userDescription,omitempty"`
	VstoreId        string `json:"vstoreId,omitempty"`
}

// CreateUserResponse represents a response to create a user
type CreateUserResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// GetUserRequest represents a request to get user information
type GetUserRequest struct {
	Name     string `json:"name,omitempty"`
	Id       string `json:"id,omitempty"`
	VstoreId string `json:"vstoreId,omitempty"`
}

// User represents a user object
type User struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	UserDescription string `json:"userDescription"`
	Path            string `json:"path"`
	UserType        string `json:"userType"`
	CreateTime      string `json:"createTime"`
}

// GetUserResponse represents a response to get user information
type GetUserResponse struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	UserDescription string `json:"userDescription"`
	Path            string `json:"path"`
	UserType        string `json:"userType"`
	CreateTime      string `json:"createTime"`
}

// DeleteUserRequest represents a request to delete a user
type DeleteUserRequest struct {
	Id       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	VstoreId string `json:"vstoreId,omitempty"`
}

// DeleteUserResponse represents a response to delete a user
type DeleteUserResponse struct{}

// CreateAccessKeyRequest represents a request to create an access key
type CreateAccessKeyRequest struct {
	AccessKey      string `json:"accessKey,omitempty"`
	SecretKey      string `json:"secretKey,omitempty"`
	OwnerType      string `json:"ownerType"`
	OwnerId        string `json:"ownerId,omitempty"`
	OwnerName      string `json:"ownerName,omitempty"`
	UserType       string `json:"userType,omitempty"`
	DomainPassword string `json:"domainPassword,omitempty"`
	VstoreId       string `json:"vstoreId,omitempty"`
}

// CreateAccessKeyResponse represents a response to create an access key
type CreateAccessKeyResponse struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// DeleteAccessKeyRequest represents a request to delete an access key
type DeleteAccessKeyRequest struct {
	Id        string `json:"id"`
	OwnerType string `json:"ownerType"`
	VstoreId  string `json:"vstoreId,omitempty"`
}

// DeleteAccessKeyResponse represents a response to delete an access key
type DeleteAccessKeyResponse struct{}

// ListAccessKeysRequest represents a request to list access keys
type ListAccessKeysRequest struct {
	OwnerName string `json:"ownerName,omitempty"`
	OwnerId   string `json:"ownerId,omitempty"`
	OwnerType string `json:"ownerType"`
	VstoreId  string `json:"vstoreId,omitempty"`
}

// AccessKeyInfo represents access key information
type AccessKeyInfo struct {
	Id         string `json:"id"`
	AccessKey  string `json:"accessKey"`
	CreateTime string `json:"createTime"`
	ModifyTime string `json:"modifyTime"`
}

// ListAccessKeysResponse represents a response to list access keys
type ListAccessKeysResponse []AccessKeyInfo

// LogString returns the string for logging, sensitive fields are omitted
func (r ListAccessKeysResponse) LogString() string {
	return fmt.Sprintf(`{"count":%d}`, len(r))
}

// LogString returns the string for logging, sensitive fields are omitted
func (r LoginRequest) LogString() string {
	return fmt.Sprintf(`{"username":"%s","scope":"%s"}`, r.Username, r.Scope)
}

// LogString returns the string for logging, sensitive fields are omitted
func (r LoginResponse) LogString() string {
	return fmt.Sprintf(`{"userid":"%s","deviceid":"%s","vstoreId":"%s"}`, r.UserId, r.DeviceId, r.VStoreId)
}

// LogString returns the string for logging, sensitive fields are omitted
func (r CreateAccessKeyRequest) LogString() string {
	return fmt.Sprintf(`{"ownerType":"%s","ownerId":"%s","ownerName":"%s","userType":"%s","vstoreId":"%s"}`,
		r.OwnerType, r.OwnerId, r.OwnerName, r.UserType, r.VstoreId)
}

// LogString returns the string for logging, sensitive fields are omitted
func (r CreateAccessKeyResponse) LogString() string {
	return `{}`
}

// LogString returns the string for logging, sensitive fields are omitted
func (r AccessKeyInfo) LogString() string {
	return fmt.Sprintf(`{"id":"%s","createTime":"%s","modifyTime":"%s"}`, r.Id, r.CreateTime, r.ModifyTime)
}
