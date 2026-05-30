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
	"fmt"
	"sync"
)

var (
	// ErrMissingUsername indicates the username configuration is missing
	ErrMissingUsername = errors.New("username is required")
	// ErrMissingPassword indicates the password configuration is missing
	ErrMissingPassword = errors.New("password is required")
	// ErrMissingEndpoint indicates the endpoint configuration is missing
	ErrMissingEndpoint = errors.New("endpoint is required")
)

// Config holds the configuration for Client
type Config struct {
	// Username is the username for authentication
	Username string
	// Password is the password for authentication
	Password string
	// Endpoint is the storage system endpoint URL
	Endpoint string
	// RootCA is the path to the root CA certificate (optional)
	RootCA []byte
	// MaxConcurrent is the maximum number of concurrent requests (optional)
	MaxConcurrent int
}

// ValidateConfig validates the client configuration
func (c *Config) ValidateConfig() error {
	if c.Username == "" {
		return ErrMissingUsername
	}
	if c.Password == "" {
		return ErrMissingPassword
	}
	if c.Endpoint == "" {
		return ErrMissingEndpoint
	}
	return nil
}

// Client implements the UserAPI interface for centralized storage
type Client struct {
	config     *Config
	httpClient HTTPClient
	session    Authenticator
	semaphore  chan struct{}
	mu         sync.Mutex
}

// NewClient creates a new Client instance
func NewClient(config *Config) (*Client, error) {
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	newHTTPClient := NewHTTPClient(config.Endpoint, config.RootCA)
	session := NewSession(config.Endpoint, config.Username, config.Password)
	session.SetHTTPClient(newHTTPClient)

	maxConcurrent := GetMaxConcurrent(config.MaxConcurrent)
	semaphore := GetSemaphoreManager().GetSemaphore(config.Endpoint, maxConcurrent)

	return &Client{
		config:     config,
		httpClient: newHTTPClient,
		session:    session,
		semaphore:  semaphore,
	}, nil
}

// ensureAuthenticated ensures the session is authenticated before making API calls.
// It performs lazy authentication - only logs in on the first call.
// This method is thread-safe and uses the session's mutex for synchronization.
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session.IsAuthenticated() {
		return nil
	}

	_, err := c.session.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

// getSemaphore returns the semaphore for concurrency control
func (c *Client) getSemaphore() chan struct{} {
	return c.semaphore
}

// acquireSemaphore acquires the semaphore, blocking if necessary
func (c *Client) acquireSemaphore(ctx context.Context) error {
	select {
	case <-c.semaphore:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// releaseSemaphore releases the semaphore after operation completes
func (c *Client) releaseSemaphore() {
	c.semaphore <- struct{}{}
}

// getHTTPClient returns the HTTP client instance
func (c *Client) getHTTPClient() HTTPClient {
	return c.httpClient
}

// getSession returns the session instance
func (c *Client) getSession() Authenticator {
	return c.session
}

// getVStoreID returns the VStore ID from the session
func (c *Client) getVStoreID() string {
	return c.session.GetVStoreID()
}

// Close releases session resources by calling session's Close method
func (c *Client) Close(ctx context.Context) error {
	if err := c.session.Close(ctx); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}
	return nil
}

func (c *Client) GetUrl(path string) string {
	return c.session.GetBaseURL() + path
}

// HTTPCallFunc defines the function type for HTTP calls.
// The ret parameter is a pointer to the response structure that will be filled by the HTTP call.
type HTTPCallFunc func(ret interface{}) error

// doRequest executes a complete business request flow.
// Encapsulates: concurrency control + authentication + HTTP call + response parsing + error conversion.
//
// This is a generic function (not a method) because Go does not support generic methods.
//
// Parameters:
//   - ctx: context
//   - client: *Client instance
//   - httpFn: HTTP call function, provided by business method, responsible for calling HTTPClient
//
// Returns:
//   - HttpResponse[T]: complete response including data and error information
//   - error: HTTP or business error
func doRequest[T any](ctx context.Context, client *Client, httpFn HTTPCallFunc) (HttpResponse[T], error) {
	var httpResponse HttpResponse[T]

	// 1. Concurrency control (semaphore)
	if err := client.acquireSemaphore(ctx); err != nil {
		return httpResponse, fmt.Errorf("failed to acquire semaphore: %w", err)
	}
	defer client.releaseSemaphore()

	// 2. Ensure authentication (lazy loading)
	if err := client.ensureAuthenticated(ctx); err != nil {
		return httpResponse, fmt.Errorf("authentication failed: %w", err)
	}

	// 3. Execute HTTP call (fills httpResponse)
	if err := httpFn(&httpResponse); err != nil {
		return httpResponse, err
	}

	// 4. Check business error code
	if httpResponse.Error.Code != 0 {
		return httpResponse, fmt.Errorf("API error: code=%d, description=%s",
			httpResponse.Error.Code, httpResponse.Error.Description)
	}

	// 5. Return complete response
	return httpResponse, nil
}
