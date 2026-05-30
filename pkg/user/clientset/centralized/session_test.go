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

package centralized

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	// Arrange
	endpoint := "https://xxxx.com:8088"
	username := "admin"
	password := "password123"

	// Act
	session := NewSession(endpoint, username, password)

	// Assert
	assert.NotNil(t, session, "session should not be nil")
	assert.Equal(t, endpoint, session.endpoint, "endpoint should match")
	assert.Equal(t, username, session.username, "username should match")
	assert.Equal(t, password, session.password, "password should match")
	assert.Equal(t, systemVStoreID, session.vstoreId, "vstoreId should be systemVStoreID")
	assert.Equal(t, initDeviceID, session.deviceID, "deviceID should be initDeviceID")
	assert.False(t, session.authenticated, "session should not be authenticated")
}

func TestSessionSetHTTPClient(t *testing.T) {
	// Arrange
	session := NewSession("https://xxxx.com:8088", "admin", "password")
	mockClient := &mockHTTPClient{}

	// Act
	session.SetHTTPClient(mockClient)

	// Assert
	assert.NotNil(t, session.httpClient, "httpClient should be set")
}

func TestSessionAuthenticateWhenAlreadyAuthenticated(t *testing.T) {
	// Arrange
	session := &Session{
		authenticated: true,
		deviceID:      "device-123",
	}
	ctx := context.Background()

	// Act
	deviceID, err := session.Authenticate(ctx)

	// Assert
	assert.NoError(t, err, "should not error when already authenticated")
	assert.Equal(t, "device-123", deviceID, "should return deviceID")
}

func TestSessionAuthenticateWhenLoginSucceeds(t *testing.T) {
	// Arrange
	session := &Session{
		username:   "admin",
		password:   "password123",
		endpoint:   "https://xxxx.com:8088",
		vstoreId:   systemVStoreID,
		deviceID:   initDeviceID,
		httpClient: nil,
	}

	resp := `{"data":{"userid":"user-123","deviceid":"device-456","vstoreId":"vstore-1","iBaseToken":"token-abc"},
             "error":{"code":0,"description":""}}`
	mockClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(resp)),
			}
		},
	}
	session.httpClient = mockClient
	ctx := context.Background()

	// Act
	deviceID, err := session.Authenticate(ctx)

	// Assert
	assert.NoError(t, err, "should not error when login succeeds")
	assert.Equal(t, "device-456", deviceID, "should return correct deviceID")
	assert.True(t, session.authenticated, "session should be authenticated")
	assert.Equal(t, "token-abc", session.ibaToken, "ibaToken should be set")
	assert.Equal(t, "vstore-1", session.vstoreId, "vstoreId should be updated")
}

func TestSessionAuthenticateWhenLoginFails(t *testing.T) {
	// Arrange
	session := &Session{
		username:   "admin",
		password:   "password123",
		endpoint:   "https://xxxx.com:8088",
		vstoreId:   systemVStoreID,
		deviceID:   initDeviceID,
		httpClient: nil,
	}

	mockClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{},"error":{"code":1077949069,"description":"Invalid username or password"}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	session.httpClient = mockClient
	ctx := context.Background()

	// Act
	deviceID, err := session.Authenticate(ctx)

	// Assert
	assert.Error(t, err, "should error when login fails")
	assert.Empty(t, deviceID, "deviceID should be empty")
	assert.False(t, session.authenticated, "session should not be authenticated")
}

func TestSessionAuthenticateWhenHTTPCallFails(t *testing.T) {
	// Arrange
	session := &Session{
		username:   "admin",
		password:   "password123",
		endpoint:   "https://xxxx.com:8088",
		vstoreId:   systemVStoreID,
		deviceID:   initDeviceID,
		httpClient: nil,
	}

	mockClient := &mockHTTPClient{
		errFunc: func() error {
			return assert.AnError
		},
	}
	session.httpClient = mockClient
	ctx := context.Background()

	// Act
	deviceID, err := session.Authenticate(ctx)

	// Assert
	assert.Error(t, err, "should error when HTTP call fails")
	assert.Empty(t, deviceID, "deviceID should be empty")
}

func TestSessionGetDeviceID(t *testing.T) {
	// Arrange
	session := &Session{
		deviceID: "device-123",
	}

	// Act
	deviceID := session.GetDeviceID()

	// Assert
	assert.Equal(t, "device-123", deviceID, "should return correct deviceID")
}

func TestSessionGetBaseURL(t *testing.T) {
	// Arrange
	session := &Session{
		endpoint: "https://xxxx.com:8088/",
		deviceID: "device-123",
	}

	// Act
	baseURL := session.GetBaseURL()

	// Assert
	assert.Equal(t, "https://xxxx.com:8088/deviceManager/rest/device-123", baseURL, "should return correct base URL")
}

func TestSessionGetBaseURLWithoutTrailingSlash(t *testing.T) {
	// Arrange
	session := &Session{
		endpoint: "https://xxxx.com:8088",
		deviceID: "device-456",
	}

	// Act
	baseURL := session.GetBaseURL()

	// Assert
	assert.Equal(t, "https://xxxx.com:8088/deviceManager/rest/device-456", baseURL)
}

func TestSessionGetAuthHeaders(t *testing.T) {
	// Arrange
	session := &Session{
		ibaToken: "token-abc-123",
	}

	// Act
	headers := session.GetAuthHeaders()

	// Assert
	assert.NotNil(t, headers, "headers should not be nil")
	assert.Equal(t, "token-abc-123", headers["iBaseToken"], "should return correct iBaseToken")
}

func TestSessionIsAuthenticated(t *testing.T) {
	// Arrange
	session := &Session{
		authenticated: true,
	}

	// Act
	authenticated := session.IsAuthenticated()

	// Assert
	assert.True(t, authenticated, "should return true when authenticated")
}

func TestSessionIsAuthenticatedWhenNotAuthenticated(t *testing.T) {
	// Arrange
	session := &Session{
		authenticated: false,
	}

	// Act
	authenticated := session.IsAuthenticated()

	// Assert
	assert.False(t, authenticated, "should return false when not authenticated")
}

func TestSessionGetVStoreID(t *testing.T) {
	// Arrange
	session := &Session{
		vstoreId: "vstore-123",
	}

	// Act
	vstoreID := session.GetVStoreID()

	// Assert
	assert.Equal(t, "vstore-123", vstoreID, "should return correct vstoreID")
}

func TestSessionCloseWhenNotAuthenticated(t *testing.T) {
	// Arrange
	session := &Session{
		authenticated: false,
		deviceID:      "device-123",
	}
	ctx := context.Background()

	// Act
	err := session.Close(ctx)

	// Assert
	assert.NoError(t, err, "should not error when closing unauthenticated session")
	assert.False(t, session.authenticated, "should remain unauthenticated")
	assert.Empty(t, session.deviceID, "deviceID should be cleared")
}

func TestSessionCloseWhenHTTPClientNil(t *testing.T) {
	// Arrange
	session := &Session{
		authenticated: true,
		deviceID:      "device-123",
		httpClient:    nil,
	}
	ctx := context.Background()

	// Act
	err := session.Close(ctx)

	// Assert
	assert.NoError(t, err, "should not error when httpClient is nil")
	assert.True(t, session.authenticated, "should remain authenticated when httpClient is nil")
	assert.Equal(t, "device-123", session.deviceID, "deviceID should remain unchanged")
}

func TestSessionCloseWhenLogoutSucceeds(t *testing.T) {
	// Arrange
	session := &Session{
		authenticated: true,
		deviceID:      "device-123",
		vstoreId:      "vstore-1",
		ibaToken:      "token-abc",
		httpClient:    nil,
	}

	mockClient := &mockHTTPClient{
		responseFunc: func() *http.Response {
			responseBody := `{"data":{},"error":{"code":0,"description":""}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
		},
	}
	session.httpClient = mockClient
	ctx := context.Background()

	// Act
	err := session.Close(ctx)

	// Assert
	assert.NoError(t, err, "should not error when logout succeeds")
	assert.False(t, session.authenticated, "should be unauthenticated after close")
	assert.Empty(t, session.deviceID, "deviceID should be cleared")
	assert.Empty(t, session.vstoreId, "vstoreId should be cleared")
	assert.Empty(t, session.ibaToken, "ibaToken should be cleared")
}

func TestSessionCloseWhenLogoutFails(t *testing.T) {
	// Arrange
	session := &Session{
		authenticated: true,
		deviceID:      "device-123",
		httpClient:    nil,
	}

	mockClient := &mockHTTPClient{
		errFunc: func() error {
			return assert.AnError
		},
	}
	session.httpClient = mockClient
	ctx := context.Background()

	// Act
	err := session.Close(ctx)

	// Assert
	assert.Error(t, err, "should error when logout fails")
	assert.Contains(t, err.Error(), "logout failed", "error message should indicate logout failure")
	assert.False(t, session.authenticated, "should be unauthenticated after close")
}
