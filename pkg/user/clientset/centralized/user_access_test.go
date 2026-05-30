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
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/huawei/cosi-driver/pkg/user/api"
)

var createUserAccessResponseBody = `{
    "data": {
        "accessKey": "fake-ak",
        "secretKey": "fake-sk"
    },
    "error": {
        "code": 0,
        "description": ""
    }
}`

func TestCreateUserAccess(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := createUserAccessResponseBody
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.CreateUserAccessInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUserAccess(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, output, "output should not be nil")
	assert.Equal(t, "fake-ak", output.AccessKeyId, "access key ID should match")
	assert.Equal(t, "fake-sk", output.SecretAccessKey, "secret access key should match")
}

func TestCreateUserAccessWhenAuthFails(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: false,
		authError:       errors.New("auth failed"),
	}

	client := NewMockClient(t, mockSession, nil)

	ctx := context.Background()
	input := &api.CreateUserAccessInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUserAccess(ctx, input)

	// Assert
	assert.Error(t, err, "should error when authentication fails")
	assert.Nil(t, output, "output should be nil")
	assert.Contains(t, err.Error(), "authentication failed", "error message should indicate auth failure")
}

func TestCreateUserAccessWhenHTTPCallFails(t *testing.T) {
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
	input := &api.CreateUserAccessInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.CreateUserAccess(ctx, input)

	// Assert
	assert.Error(t, err, "should error when HTTP call fails")
	assert.Nil(t, output, "output should be nil")
	assert.Contains(t, err.Error(), "connection refused", "error message should indicate HTTP failure")
}

func TestDeleteUserAccess(t *testing.T) {
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
	input := &api.DeleteUserAccessInput{
		UserName:    "test-user",
		AccessKeyId: "fake-ak",
	}

	// Act
	output, err := client.DeleteUserAccess(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, output, "output should not be nil")
}

func TestDeleteUserAccessWhenAKNotFound(t *testing.T) {
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
	input := &api.DeleteUserAccessInput{
		UserName:    "test-user",
		AccessKeyId: "non-existent-ak",
	}

	// Act
	output, err := client.DeleteUserAccess(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when AK not found (idempotent)")
	assert.NotNil(t, output, "output should not be nil")
}

var listUserAccessKeysResponseBody = `{
    "data": [
        {
            "id": "ak-1",
            "accessKey": "fake-ak",
            "status": true,
            "createTime": "2026-01-01",
            "modifyTime": "2026-01-01"
        },
        {
            "id": "ak-2",
            "accessKey": "fake-sk",
            "status": true,
            "createTime": "2026-01-02",
            "modifyTime": "2026-01-02"
        }
    ],
    "error": {
        "code": 0,
        "description": ""
    }
}`

func TestListUserAccessKeys(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := listUserAccessKeysResponseBody
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.ListUserAccessKeysInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.ListUserAccessKeys(ctx, input)

	// Assert
	accessKeysLength := 2
	assert.NoError(t, err, "should not error when request succeeds")
	assert.NotNil(t, output, "output should not be nil")
	assert.Len(t, output.AccessKeys, accessKeysLength, "should return 2 access keys")
	assert.Contains(t, output.AccessKeys, "ak-1", "should contain first access key id")
	assert.Contains(t, output.AccessKeys, "ak-2", "should contain second access key id")
}

func TestListUserAccessKeysWhenEmpty(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "device-123",
	}

	mockHTTPClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":[],"error":{"code":0,"description":""}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	client := NewMockClient(t, mockSession, mockHTTPClient)

	ctx := context.Background()
	input := &api.ListUserAccessKeysInput{
		UserName: "test-user",
	}

	// Act
	output, err := client.ListUserAccessKeys(ctx, input)

	// Assert
	assert.NoError(t, err, "should not error when no access keys")
	assert.NotNil(t, output, "output should not be nil")
	assert.Empty(t, output.AccessKeys, "should return empty access keys list")
}
