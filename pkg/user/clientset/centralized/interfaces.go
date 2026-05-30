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
)

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	// GET sends a GET request to the specified URL with optional query parameters.
	// The ret parameter is a pointer to the response structure that will be filled.
	GET(ctx context.Context, url string, params map[string]string, ret interface{}) error

	// POST sends a POST request with the given body to the specified URL.
	// The ret parameter is a pointer to the response structure that will be filled.
	POST(ctx context.Context, url string, body interface{}, ret interface{}) error

	// PUT sends a PUT request with the given body to the specified URL.
	// The ret parameter is a pointer to the response structure that will be filled.
	PUT(ctx context.Context, url string, body interface{}, ret interface{}) error

	// DELETE sends a DELETE request to the specified URL with query parameters.
	// The ret parameter is a pointer to the response structure that will be filled.
	DELETE(ctx context.Context, url string, params map[string]string, ret interface{}) error

	// SetHeaders sets the default headers for all requests.
	SetHeaders(headers map[string]string)
}

// Authenticator defines the interface for authentication operations.
type Authenticator interface {
	// Authenticate performs login and returns the device ID.
	Authenticate(ctx context.Context) (string, error)

	// GetDeviceID returns the device ID from the current session.
	GetDeviceID() string

	// GetVStoreID returns the VStore ID from the current session.
	GetVStoreID() string

	// GetBaseURL returns the base URL for API requests.
	GetBaseURL() string

	// GetAuthHeaders returns the authentication headers for API requests.
	GetAuthHeaders() map[string]string

	// IsAuthenticated returns true if the session is authenticated.
	IsAuthenticated() bool

	// Close performs logout and cleans up session resources.
	Close(ctx context.Context) error
}
