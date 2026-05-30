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
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

const emptyFlag = "<empty>"

// LogMarshaler is an interface for customizing log serialization.
// Types implementing this interface can control their own log output to omit sensitive fields.
type LogMarshaler interface {
	// LogString returns the string for logging, with sensitive fields omitted
	LogString() string
}

// httpClient implements HTTPClient interface with cookie jar support.
type httpClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

const defaultTimeout = 60 * time.Second

// NewHTTPClient creates a new HTTP client with the given base URL and root CA certificate.
func NewHTTPClient(baseURL string, rootCA []byte) HTTPClient {
	// cookiejar.New only fails with unsupported options, nil always succeeds
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}

	client := &http.Client{
		Timeout: defaultTimeout,
		Transport: &http.Transport{
			TLSClientConfig: utils.BuildTLSConfig(rootCA),
		},
		Jar: jar,
	}

	return &httpClient{
		client:  client,
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// SetHeaders sets the default headers for all requests.
func (c *httpClient) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		c.headers[k] = v
	}
}

// GET sends a GET request to the specified URL with optional query parameters.
func (c *httpClient) GET(ctx context.Context, url string, queryParams map[string]string, ret interface{}) error {
	url = appendQueryParams(url, queryParams)
	return c.sendRequest(ctx, http.MethodGet, url, nil, ret)
}

// POST sends a POST request with JSON body to the specified URL.
func (c *httpClient) POST(ctx context.Context, url string, body interface{}, ret interface{}) error {
	return c.sendRequest(ctx, http.MethodPost, url, body, ret)
}

// PUT sends a PUT request with JSON body to the specified URL.
func (c *httpClient) PUT(ctx context.Context, url string, body interface{}, ret interface{}) error {
	return c.sendRequest(ctx, http.MethodPut, url, body, ret)
}

// DELETE sends a DELETE request to the specified URL with query parameters.
// Query params are appended to the URL, not sent as JSON body.
func (c *httpClient) DELETE(ctx context.Context, url string, queryParams map[string]string, ret interface{}) error {
	url = appendQueryParams(url, queryParams)
	return c.sendRequest(ctx, http.MethodDelete, url, nil, ret)
}

// setHeaders sets headers on the request.
func (c *httpClient) setHeaders(req *http.Request) {
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}
}

// appendQueryParams appends query parameters to the URL.
func appendQueryParams(rawURL string, queryParams map[string]string) string {
	if len(queryParams) == 0 {
		return rawURL
	}

	q := make(url.Values)
	for k, v := range queryParams {
		if k != "" && v != "" {
			q.Add(k, v)
		}
	}

	if separator := "?"; strings.Contains(rawURL, separator) {
		return rawURL + "&" + q.Encode()
	}
	return rawURL + "?" + q.Encode()
}

// getLogBody returns the log string for the request body, with sensitive fields omitted
func getLogBody(body interface{}, jsonData []byte) string {
	if body == nil {
		return ""
	}
	marshaler, ok := body.(LogMarshaler)
	if ok {
		mask := marshaler.LogString()
		if mask != emptyFlag {
			return mask
		}
	}
	return string(jsonData)
}

// sendRequest sends HTTP requests with optional JSON body.
func (c *httpClient) sendRequest(ctx context.Context, method, url string, body interface{}, ret interface{}) error {
	var req *http.Request
	var err error

	if body != nil {
		// Serialize request body
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}

		log.AddContext(ctx).Infof("%s %s %s", method, url, getLogBody(body, jsonData))
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		// Log request without body
		log.AddContext(ctx).Infof("%s %s", method, url)
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return err
		}
	}

	c.setHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.handleResponse(ctx, method, url, resp, ret)
}

// handleResponse reads response body, parses JSON to ret, and logs the response
func (c *httpClient) handleResponse(ctx context.Context, method, url string, resp *http.Response,
	ret interface{}) error {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 5. Parse JSON to ret
	if ret != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, ret); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	// 6. Log response
	log.AddContext(ctx).Infof("%s %s -> %d %s", method, url, resp.StatusCode, getLogBody(ret, respBody))
	return nil
}
