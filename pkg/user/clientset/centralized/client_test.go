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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidateConfigWhenUsernameMissing(t *testing.T) {
	// Arrange
	config := &Config{
		Username: "",
		Password: "password123",
		Endpoint: "https://xxxx.com:8088",
	}

	// Act
	err := config.ValidateConfig()

	// Assert
	assert.Error(t, err, "should return error when username is missing")
	assert.Equal(t, ErrMissingUsername, err, "error should be ErrMissingUsername")
}

func TestConfigValidateConfigWhenPasswordMissing(t *testing.T) {
	// Arrange
	config := &Config{
		Username: "admin",
		Password: "",
		Endpoint: "https://xxxx.com:8088",
	}

	// Act
	err := config.ValidateConfig()

	// Assert
	assert.Error(t, err, "should return error when password is missing")
	assert.Equal(t, ErrMissingPassword, err, "error should be ErrMissingPassword")
}

func TestConfigValidateConfigWhenEndpointMissing(t *testing.T) {
	// Arrange
	config := &Config{
		Username: "admin",
		Password: "password123",
		Endpoint: "",
	}

	// Act
	err := config.ValidateConfig()

	// Assert
	assert.Error(t, err, "should return error when endpoint is missing")
	assert.Equal(t, ErrMissingEndpoint, err, "error should be ErrMissingEndpoint")
}

func TestConfigValidateConfigWhenAllFieldsValid(t *testing.T) {
	// Arrange
	config := &Config{
		Username: "admin",
		Password: "password123",
		Endpoint: "https://xxxx.com:8088",
	}

	// Act
	err := config.ValidateConfig()

	// Assert
	assert.NoError(t, err, "should not return error when all fields are valid")
}

func TestNewClientWhenInvalidConfig(t *testing.T) {
	// Arrange
	config := &Config{
		Username: "",
		Password: "password123",
		Endpoint: "https://xxxx.com:8088",
	}

	// Act
	client, err := NewClient(config)

	// Assert
	assert.Error(t, err, "should return error when config is invalid")
	assert.Nil(t, client, "client should be nil")
}

func TestNewClientWhenValidConfig(t *testing.T) {
	// Arrange
	config := &Config{
		Username:      "admin",
		Password:      "password123",
		Endpoint:      "https://xxxx.com:8088",
		RootCA:        []byte{}, // Empty byte slice for testing
		MaxConcurrent: 50,
	}

	// Act
	client, err := NewClient(config)

	// Assert
	assert.NoError(t, err, "should not return error when config is valid")
	assert.NotNil(t, client, "client should not be nil")
	assert.Equal(t, config, client.config, "client config should match input config")
}

func TestNewClientWithDefaultMaxConcurrent(t *testing.T) {
	// Arrange
	config := &Config{
		Username: "admin",
		Password: "password123",
		Endpoint: "https://xxxx.com:8088",
	}

	// Act
	client, err := NewClient(config)

	// Assert
	assert.NoError(t, err, "should not return error when config is valid")
	assert.NotNil(t, client, "client should not be nil")
}

func TestCentralizedClientEnsureAuthenticatedWhenAlreadyAuthenticated(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
	}
	client := &Client{
		session: mockSession,
	}
	ctx := context.Background()

	// Act
	err := client.ensureAuthenticated(ctx)

	// Assert
	assert.NoError(t, err, "should not error when already authenticated")
}

func TestCentralizedClientEnsureAuthenticatedWhenNotAuthenticated(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: false,
		deviceID:        "test-device-id",
	}
	client := &Client{
		session: mockSession,
	}
	ctx := context.Background()

	// Act
	err := client.ensureAuthenticated(ctx)

	// Assert
	assert.NoError(t, err, "should authenticate successfully")
	assert.True(t, mockSession.isAuthenticated, "session should be authenticated")
}

func TestCentralizedClientEnsureAuthenticatedWhenAuthFails(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: false,
		authError:       context.DeadlineExceeded,
	}
	client := &Client{
		session: mockSession,
	}
	ctx := context.Background()

	// Act
	err := client.ensureAuthenticated(ctx)

	// Assert
	assert.Error(t, err, "should return error when authentication fails")
	assert.Contains(t, err.Error(), "authentication failed", "error message should indicate auth failure")
}

func TestCentralizedClientGetSemaphore(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		semaphore <- struct{}{}
	}
	client := &Client{
		semaphore: semaphore,
	}

	// Act
	result := client.getSemaphore()

	// Assert
	assert.NotNil(t, result, "should return non-nil semaphore")
	assert.Equal(t, semaphore, result, "should return the same semaphore")
}

func TestCentralizedClientAcquireSemaphoreWhenSuccess(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	client := &Client{
		semaphore: semaphore,
	}
	ctx := context.Background()

	// Act
	err := client.acquireSemaphore(ctx)

	// Assert
	assert.NoError(t, err, "should acquire semaphore successfully")
}

func TestCentralizedClientAcquireSemaphoreWhenContextCancelled(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	client := &Client{
		semaphore: semaphore,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err := client.acquireSemaphore(ctx)

	// Assert
	assert.Error(t, err, "should return error when context is cancelled")
	assert.Equal(t, context.Canceled, err, "error should be context.Canceled")
}

func TestCentralizedClientReleaseSemaphore(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	client := &Client{
		semaphore: semaphore,
	}

	// Act
	client.releaseSemaphore()

	// Assert
	select {
	case <-semaphore:
		assert.True(t, true, "semaphore should be released")
	default:
		assert.Fail(t, "semaphore was not released")
	}
}

func TestCentralizedClientGetHTTPClient(t *testing.T) {
	// Arrange
	mockHTTPClient := &mockHTTPClient{}
	client := &Client{
		httpClient: mockHTTPClient,
	}

	// Act
	result := client.getHTTPClient()

	// Assert
	assert.NotNil(t, result, "should return non-nil HTTP client")
	assert.Equal(t, mockHTTPClient, result, "should return the same HTTP client")
}

func TestCentralizedClientGetSession(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
	}
	client := &Client{
		session: mockSession,
	}

	// Act
	result := client.getSession()

	// Assert
	assert.NotNil(t, result, "should return non-nil session")
	assert.Equal(t, mockSession, result, "should return the same session")
}

func TestCentralizedClientGetVStoreID(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
		vstoreID:        "vstore-123",
	}
	client := &Client{
		session: mockSession,
	}

	// Act
	result := client.getVStoreID()

	// Assert
	assert.Equal(t, "vstore-123", result, "should return session's VStore ID")
}

func TestCentralizedClientCloseWhenSessionCloseSucceeds(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
		closeCalled:     false,
	}
	client := &Client{
		session: mockSession,
	}
	ctx := context.Background()

	// Act
	err := client.Close(ctx)

	// Assert
	assert.NoError(t, err, "should not error when session close succeeds")
	assert.True(t, mockSession.closeCalled, "session Close should be called")
}

func TestCentralizedClientCloseWhenSessionCloseFails(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
		closeError:      assert.AnError,
	}
	client := &Client{
		session: mockSession,
	}
	ctx := context.Background()

	// Act
	err := client.Close(ctx)

	// Assert
	assert.Error(t, err, "should return error when session close fails")
	assert.Contains(t, err.Error(), "failed to close session", "error message should indicate close failure")
}

func TestCentralizedClientGetUrl(t *testing.T) {
	// Arrange
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
		baseURL:         "https://xxx.com/deviceManager/rest/test-device-id",
	}
	client := &Client{
		session: mockSession,
	}
	path := "/users"

	// Act
	result := client.GetUrl(path)

	// Assert
	expected := "https://xxx.com/deviceManager/rest/test-device-id/users"
	assert.Equal(t, expected, result, "should return correct URL")
}

func TestDoRequestWhenAuthenticationFails(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	mockSession := &mockAuthenticator{
		isAuthenticated: false,
		authError:       errors.New("auth failed"),
	}
	client := &Client{
		semaphore: semaphore,
		session:   mockSession,
	}
	ctx := context.Background()

	httpFn := func(ret interface{}) error {
		return nil
	}

	// Act
	result, err := doRequest[any](ctx, client, httpFn)

	// Assert
	assert.Error(t, err, "should return error when authentication fails")
	assert.Contains(t, err.Error(), "authentication failed", "error message should indicate auth failure")
	assert.Zero(t, result.Error.Code, "response error code should be zero")
}

func TestDoRequestWhenHTTPCallFails(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
	}
	client := &Client{
		semaphore: semaphore,
		session:   mockSession,
	}
	ctx := context.Background()

	httpError := errors.New("HTTP call failed")
	httpFn := func(ret interface{}) error {
		return httpError
	}

	// Act
	result, err := doRequest[any](ctx, client, httpFn)

	// Assert
	assert.Error(t, err, "should return error when HTTP call fails")
	assert.Equal(t, httpError, err, "error should be the HTTP call error")
	assert.Zero(t, result.Error.Code, "response error code should be zero")
}

func TestDoRequestWhenAPIReturnsErrorCode(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
	}
	client := &Client{
		semaphore: semaphore,
		session:   mockSession,
	}
	ctx := context.Background()

	httpFn := func(ret interface{}) error {
		if resp, ok := ret.(*HttpResponse[any]); ok {
			resp.Error = ErrorInfo{
				Code:        404,
				Description: "Not Found",
			}
		}
		return nil
	}

	// Act
	result, err := doRequest[any](ctx, client, httpFn)

	// Assert
	assert.Error(t, err, "should return error when API returns error code")
	assert.Contains(t, err.Error(), "API error", "error message should indicate API error")
	assert.Equal(t, int64(404), result.Error.Code, "response error code should be 404")
}

func TestDoRequestWhenSuccess(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
	}
	client := &Client{
		semaphore: semaphore,
		session:   mockSession,
	}
	ctx := context.Background()

	expectedData := map[string]interface{}{"key": "value"}
	httpFn := func(ret interface{}) error {
		if resp, ok := ret.(*HttpResponse[any]); ok {
			resp.Data = expectedData
		}
		return nil
	}

	// Act
	result, err := doRequest[any](ctx, client, httpFn)

	// Assert
	assert.NoError(t, err, "should not return error when request succeeds")
	assert.Zero(t, result.Error.Code, "response error code should be zero")
	assert.Equal(t, expectedData, result.Data, "response data should match expected data")
}

func TestDoRequestWithSemaphoreRelease(t *testing.T) {
	// Arrange
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}
	mockSession := &mockAuthenticator{
		isAuthenticated: true,
		deviceID:        "test-device-id",
	}
	client := &Client{
		semaphore: semaphore,
		session:   mockSession,
	}
	ctx := context.Background()

	httpFn := func(ret interface{}) error {
		return nil
	}

	// Act
	_, err := doRequest[any](ctx, client, httpFn)

	// Assert
	assert.NoError(t, err, "should not return error")
	select {
	case <-semaphore:
		assert.True(t, true, "semaphore should be released after request")
	default:
		assert.Fail(t, "semaphore was not released")
	}
}

// mockAuthenticator is a mock implementation of Authenticator for testing
type mockAuthenticator struct {
	isAuthenticated bool
	deviceID        string
	vstoreID        string
	baseURL         string
	authError       error
	closeCalled     bool
	closeError      error
}

func (m *mockAuthenticator) Authenticate(ctx context.Context) (string, error) {
	if m.authError != nil {
		return "", m.authError
	}
	m.isAuthenticated = true
	return m.deviceID, nil
}

func (m *mockAuthenticator) GetDeviceID() string               { return m.deviceID }
func (m *mockAuthenticator) GetVStoreID() string               { return m.vstoreID }
func (m *mockAuthenticator) GetBaseURL() string                { return m.baseURL }
func (m *mockAuthenticator) GetAuthHeaders() map[string]string { return nil }
func (m *mockAuthenticator) IsAuthenticated() bool             { return m.isAuthenticated }
func (m *mockAuthenticator) Close(ctx context.Context) error {
	m.closeCalled = true
	if m.closeError != nil {
		return m.closeError
	}
	return nil
}

// NewMockClient creates a Client instance with the given session and initializes the semaphore.
// If httpClient is provided, it will be set on the client.
// This helper function reduces boilerplate in test functions.
func NewMockClient(t *testing.T, session *mockAuthenticator, httpClient *mockHTTPClient) *Client {
	t.Helper()
	client := &Client{
		session:   session,
		semaphore: make(chan struct{}, 1),
	}
	client.semaphore <- struct{}{}

	if httpClient != nil {
		client.httpClient = httpClient
	}

	return client
}
