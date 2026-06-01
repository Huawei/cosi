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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/huawei/cosi-driver/pkg/user/api"
)

func TestCreateUser(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{"id":"user-123","name":"test-user"},"error":{"code":0,"description":""}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.CreateUserInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUser(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, output, "output should not be nil")
	assert.Equal(t, "test-user", output.UserName, "user name should match")
	assert.Equal(t, "user-123", output.UserID, "user ID should match")
}

func TestCreateUserWhenAuthFails(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: false,
		authError:       errors.New("auth failed"),
	}

	client := NewMockClient(t, mockSession, nil)

	ctx := context.Background()
	input := &api.CreateUserInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUser(ctx, input)

	// Assert
	assert.Error(t, err, "should error when authentication fails")
	assert.Nil(t, output, "output should be nil")
	assert.Contains(t, err.Error(), "authentication failed", "error message should indicate auth failure")
}

func TestCreateUserWhenHTTPCallFails(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			return nil
		},
		errFunc: func() error {
			return errors.New("connection refused")
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.CreateUserInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUser(ctx, input)

	// Assert
	assert.Error(t, err, "should error when HTTP call fails")
	assert.Nil(t, output, "output should be nil")
	assert.Contains(t, err.Error(), "connection refused", "error message should indicate HTTP failure")
}

func TestCreateUserWhenAPIReturnsError(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{},"error":{"code":1077949069,"description":"User already exists"}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.CreateUserInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUser(ctx, input)

	// Assert
	assert.Error(t, err, "should error when API returns error")
	assert.Nil(t, output, "output should be nil")
	assert.Contains(t, err.Error(), "User already exists", "error message should indicate user already exists")
}

func TestCreateUserIncludesVStoreId(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
		vstoreID:        "vstore-456",
		baseURL:         "https://xxx.com/deviceManager/rest/device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{"id":"user-123","name":"test-user"},"error":{"code":0,"description":""}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
		setHeadersCalled: make(map[string]string),
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.CreateUserInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUser(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, output, "output should not be nil")
	assert.Equal(t, "test-user", output.UserName, "user name should match")
	assert.Equal(t, "arn:aws:iam::vstore-456:user/test-user", output.Arn, "ARN should match")
}

var getUserResponse = `{
    "data": {
        "id": "user-123",
        "name": "test-user",
        "userDescription": "test",
        "path": "/bucket",
        "userType": "normal",
        "createTime": "2024-01-01"
    },
    "error": {
        "code": 0,
        "description": ""
    }
}`

func TestGetUser(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{"id":"user-123","name":"test-user"},"error":{"code":0,"description":""}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.GetUserInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.GetUser(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, output, "output should not be nil")
	assert.Equal(t, "test-user", output.UserName, "user name should match")
	assert.Equal(t, "user-123", output.UserID, "user ID should match")
}

var getUserNotFoundResponse = `{
    "data": {
        "id": "",
        "name": "",
        "userDescription": "",
        "path": "",
        "userType": "",
        "createTime": ""
    },
    "error": {
        "code": 0,
        "description": ""
    }
}`

func TestGetUserWhenNotFound(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{},"error":{"code":1077949068,"description":"User does not exist"}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.GetUserInput{
		UserName: "non-existent-user",
	}

	// Act
	output, err := client.GetUser(ctx, input)

	// Assert
	assert.Error(t, err, "should error when user does not exist")
	assert.Nil(t, output, "output should be nil")
	assert.Contains(t, err.Error(), "User does not exist", "error message should indicate user not found")
}

func TestDeleteUser(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{},"error":{"code":0,"description":""}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.DeleteUserInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.DeleteUser(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, output, "output should not be nil")
}

func TestDeleteUserWhenNotFound(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{},"error":{"code":1092615946,"description":"User does not exist"}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.DeleteUserInput{
		UserName: "non-existent-user",
	}

	// Act
	output, err := client.DeleteUser(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when user not found (idempotent)")
	assert.NotNil(t, output, "output should not be nil")
}

func TestDoRequest(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: false,
		deviceID:        "device-123",
	}

	client := &Client{
		session:   mockSession,
		semaphore: make(chan struct{}, 1),
	}
	client.semaphore <- struct{}{}

	ctx := context.Background()

	httpFn := func(ret interface{}) error {
		response := `{"data":{"id":"user-123","name":"test-user"},"error":{"code":0,"description":""}}`
		return json.Unmarshal([]byte(response), ret)
	}

	// Act
	resp, err := doRequest[struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}](ctx, client, httpFn)

	// Assert
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, resp.Data, "result data should not be nil")
	assert.Equal(t, "user-123", resp.Data.Id, "ID should match")
	assert.Equal(t, "test-user", resp.Data.Name, "name should match")
}

func TestDoRequestWhenSemaphoreAcquireFails(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	client := &Client{
		session:   mockSession,
		semaphore: make(chan struct{}, 1),
	}
	client.semaphore <- struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	httpFn := func(ret interface{}) error {
		response := `{"data":{"id":"user-123"},"error":{"code":0,"description":""}}`
		return json.Unmarshal([]byte(response), ret)
	}

	// Act: First acquire the semaphore to simulate blocking
	_ = <-client.semaphore

	_, err := doRequest[struct {
		Id string `json:"id"`
	}](ctx, client, httpFn)

	// Assert
	assert.Error(t, err, "should error when cannot acquire semaphore")
	assert.Contains(t, err.Error(), "context canceled", "error should indicate context cancellation")
}
