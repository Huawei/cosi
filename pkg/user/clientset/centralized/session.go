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
	"fmt"
	"strings"
	"sync"
)

const (
	systemVStoreID = "0"
	initDeviceID   = "xxx"
)

// Session manages authentication session state.
type Session struct {
	username      string
	password      string
	endpoint      string
	authenticated bool
	deviceID      string
	ibaToken      string
	vstoreId      string
	mu            sync.Mutex
	httpClient    HTTPClient
}

// NewSession creates a new Session instance.
func NewSession(endpoint, username, password string) *Session {
	return &Session{
		endpoint: endpoint,
		username: username,
		password: password,
		vstoreId: systemVStoreID,
		deviceID: initDeviceID,
	}
}

// SetHTTPClient sets the HTTP client for the session.
func (s *Session) SetHTTPClient(client HTTPClient) {
	s.httpClient = client
}

// Authenticate performs login and returns the device ID.
func (s *Session) Authenticate(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.authenticated {
		return s.deviceID, nil
	}

	data, err := s.doLogin(ctx)
	if err != nil {
		return "", err
	}

	s.updateSessionState(data)
	return s.deviceID, nil
}

// doLogin performs the login request and returns login response data.
// Caller must hold s.mu lock.
func (s *Session) doLogin(ctx context.Context) (LoginResponse, error) {
	loginReq := &LoginRequest{
		Username: s.username,
		Password: s.password,
		Scope:    "0",
	}

	var result HttpResponse[LoginResponse]
	if err := s.httpClient.POST(ctx, s.GetBaseURL()+"/sessions", loginReq, &result); err != nil {
		return LoginResponse{}, err
	}

	if result.Error.Code != 0 {
		return LoginResponse{}, fmt.Errorf("login failed: code=%d, description=%s", result.Error.Code, result.Error.Description)
	}
	s.password = ""
	return result.Data, nil
}

// updateSessionState updates session state with login response data.
// Caller must hold s.mu lock.
func (s *Session) updateSessionState(data LoginResponse) {
	s.deviceID = data.DeviceId
	s.ibaToken = data.IbaseToken
	s.authenticated = true
	if data.VStoreId != "" {
		s.vstoreId = data.VStoreId
	}
	if s.httpClient != nil {
		s.httpClient.SetHeaders(s.GetAuthHeaders())
	}
}

// GetDeviceID returns the device ID from the current session.
func (s *Session) GetDeviceID() string {
	return s.deviceID
}

// GetBaseURL returns the base URL for API requests.
func (s *Session) GetBaseURL() string {
	endpoint := strings.TrimSuffix(s.endpoint, "/")
	return endpoint + "/deviceManager/rest/" + s.deviceID
}

// GetAuthHeaders returns the authentication headers for API requests.
func (s *Session) GetAuthHeaders() map[string]string {
	return map[string]string{
		"iBaseToken": s.ibaToken,
	}
}

// IsAuthenticated returns true if the session is authenticated.
func (s *Session) IsAuthenticated() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.authenticated
}

// GetVStoreID returns the VStore ID from the current session.
func (s *Session) GetVStoreID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vstoreId
}

// Close performs logout and cleans up session resources.
func (s *Session) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.authenticated {
		s.clearState()
		return nil
	}

	if s.httpClient == nil {
		return nil
	}

	var resp interface{}
	if err := s.httpClient.DELETE(ctx, s.GetBaseURL()+"/sessions", nil, &resp); err != nil {
		s.clearState()
		return fmt.Errorf("logout failed: %w", err)
	}
	s.clearState()
	return nil
}

func (s *Session) clearState() {
	s.authenticated = false
	s.deviceID = ""
	s.ibaToken = ""
	s.vstoreId = ""
}
