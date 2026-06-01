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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClientWhenNoRootCAThenCreatesClientWithInsecureTLS(t *testing.T) {
	// Arrange
	baseURL := "https://xxxx.com"

	// Act
	client := NewHTTPClient(baseURL, nil)

	// Assert
	assert.NotNil(t, client, "client should not be nil")
}

func TestNewHTTPClientWhenRootCAProvidedThenCreatesClientWithSecureTLS(t *testing.T) {
	// Arrange
	baseURL := "https://xxxx.com"
	rootCA := []byte("certificate-data")

	// Act
	client := NewHTTPClient(baseURL, rootCA)

	// Assert
	assert.NotNil(t, client, "client should not be nil")
}

func TestHTTPClientSetHeadersWhenCalledThenStoresHeaders(t *testing.T) {
	// Arrange
	client, ok := NewHTTPClient("https://xxxx.com", nil).(*httpClient)
	assert.True(t, ok, "client should be of type *httpClient")

	headers := map[string]string{
		"Authorization":   "Bearer token123",
		"X-Custom-Header": "custom-value",
	}

	// Act
	client.SetHeaders(headers)

	// Assert
	assert.Equal(t, "Bearer token123", client.headers["Authorization"])
	assert.Equal(t, "custom-value", client.headers["X-Custom-Header"])
}

func TestHTTPClientGetWhenValidURLAndNoQueryParamsThenReturnsResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()

	// Act
	var resp interface{}
	err := client.GET(ctx, server.URL, nil, &resp)

	// Assert
	assert.NoError(t, err, "GET request should not fail")
}

func TestHTTPClientGetWhenQueryParamsProvidedThenAppendsToURL(t *testing.T) {
	// Arrange
	var receivedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()
	queryParams := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// Act
	var resp interface{}
	err := client.GET(ctx, server.URL, queryParams, &resp)

	// Assert
	assert.NoError(t, err, "GET request should not fail")
	assert.Contains(t, receivedURL, "key1=value1", "URL should contain key1")
	assert.Contains(t, receivedURL, "key2=value2", "URL should contain key2")
}

func TestHTTPClientGetWhenURLAlreadyHasQueryParamsThenAppendsWithAmpersand(t *testing.T) {
	// Arrange
	var receivedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()
	queryParams := map[string]string{
		"key2": "value2",
	}
	baseURL := server.URL + "?key1=value1"

	// Act
	var resp interface{}
	err := client.GET(ctx, baseURL, queryParams, &resp)

	// Assert
	assert.NoError(t, err, "GET request should not fail")
	assert.Contains(t, receivedURL, "?key1=value1", "URL should contain original param")
	assert.Contains(t, receivedURL, "&key2=value2", "URL should contain appended param with &")
}

const testDelay = 100 * time.Millisecond

func TestHTTPClientGetWhenContextCancelledThenReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(testDelay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	var resp interface{}
	err := client.GET(ctx, server.URL, nil, &resp)

	// Assert
	assert.Error(t, err, "GET request should fail with cancelled context")
}

func TestHTTPClientPostWhenValidRequestThenReturnsResponse(t *testing.T) {
	// Arrange
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "method should be POST")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "content-type should be application/json")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()
	body := map[string]string{"name": "test"}

	// Act
	var resp interface{}
	err := client.POST(ctx, server.URL, body, &resp)

	// Assert
	assert.NoError(t, err, "POST request should not fail")
	assert.Contains(t, string(receivedBody), "test", "body should contain test")
}

func TestHTTPClientPostWhenNilBodyThenSucceedsWithoutBody(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "method should be POST")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()

	// Act
	var resp interface{}
	err := client.POST(ctx, server.URL, nil, &resp)

	// Assert
	assert.NoError(t, err, "POST request should not fail")
}

func TestHTTPClientPutWhenValidRequestThenReturnsResponse(t *testing.T) {
	// Arrange
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "method should be PUT")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "content-type should be application/json")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()
	body := map[string]interface{}{"id": "123", "name": "updated"}

	// Act
	var resp interface{}
	err := client.PUT(ctx, server.URL, body, &resp)

	// Assert
	assert.NoError(t, err, "PUT request should not fail")
	assert.Contains(t, string(receivedBody), "123", "body should contain id")
}

func TestHTTPClientDeleteWhenValidRequestThenReturnsResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "method should be DELETE")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()

	// Act
	var resp interface{}
	err := client.DELETE(ctx, server.URL, nil, &resp)

	// Assert
	assert.NoError(t, err, "DELETE request should not fail")
}

func TestHTTPClientDeleteWhenQueryParamsProvidedThenAppendsToURL(t *testing.T) {
	// Arrange
	var receivedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "method should be DELETE")
		receivedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()
	queryParams := map[string]string{"id": "123", "ownerType": "1"}

	// Act
	var resp interface{}
	err := client.DELETE(ctx, server.URL, queryParams, &resp)

	// Assert
	assert.NoError(t, err, "DELETE request should not fail")
	assert.Contains(t, receivedURL, "id=123", "URL should contain id parameter")
	assert.Contains(t, receivedURL, "ownerType=1", "URL should contain ownerType parameter")
}

func TestHTTPClientSendRequestWhenSetHeadersCalledThenIncludesHeadersInRequest(t *testing.T) {
	// Arrange
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	customHeaders := map[string]string{
		"X-Custom-Auth": "secret-token",
		"X-Request-ID":  "req-123",
	}

	client.SetHeaders(customHeaders)
	ctx := context.Background()

	// Act
	var resp interface{}
	err := client.GET(ctx, server.URL, nil, &resp)

	// Assert
	assert.NoError(t, err, "GET request should not fail")
	assert.Equal(t, "secret-token", receivedHeaders.Get("X-Custom-Auth"), "custom auth header should match")
	assert.Equal(t, "req-123", receivedHeaders.Get("X-Request-ID"), "request ID header should match")
}

func TestAppendQueryParamsWhenEmptyQueryParamsThenReturnsOriginalURL(t *testing.T) {
	// Arrange
	originalURL := "https://xxxx.com/api/resource"
	queryParams := make(map[string]string)

	// Act
	result := appendQueryParams(originalURL, queryParams)

	// Assert
	assert.Equal(t, originalURL, result, "URL should remain unchanged")
}

func TestAppendQueryParamsWhenNoExistingQueryThenAppendsWithQuestionMark(t *testing.T) {
	// Arrange
	originalURL := "https://xxxx.com/api/resource"
	queryParams := map[string]string{
		"key": "value",
	}

	// Act
	result := appendQueryParams(originalURL, queryParams)

	// Assert
	assert.Contains(t, result, "?key=value", "should append with ?")
}

func TestAppendQueryParamsWhenExistingQueryThenAppendsWithAmpersand(t *testing.T) {
	// Arrange
	originalURL := "https://xxxx.com/api/resource?existing=param"
	queryParams := map[string]string{
		"key": "value",
	}

	// Act
	result := appendQueryParams(originalURL, queryParams)

	// Assert
	assert.Contains(t, result, "?existing=param", "should contain original param")
	assert.Contains(t, result, "&key=value", "should append with &")
}

func TestAppendQueryParamsWhenMultipleParamsThenEncodesAll(t *testing.T) {
	// Arrange
	originalURL := "https://xxxx.com/api/resource"
	queryParams := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// Act
	result := appendQueryParams(originalURL, queryParams)

	// Assert
	assert.Contains(t, result, "key1=value1", "should contain key1")
	assert.Contains(t, result, "key2=value2", "should contain key2")
	assert.Contains(t, result, "key3=value3", "should contain key3")
}

func TestHTTPClientWhenMultipleRequestsThenMaintainsSession(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":{},"error":{"code":0}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, nil)

	// Act
	var resp1, resp2 interface{}
	err := client.GET(context.Background(), server.URL+"/test1", nil, &resp1)
	assert.NoError(t, err)

	err = client.GET(context.Background(), server.URL+"/test2", nil, &resp2)
	assert.NoError(t, err)
}

func TestHTTPClientWhenMultipleRequestsThenMaintainsSessionCookie(t *testing.T) {
	// Arrange
	cookieName := "sessionid"
	cookieValue := "test123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Cookies()
		if len(cookies) > 0 {
			w.Header().Set("X-Has-Cookie", "true")
		} else {
			// Set cookie on first request
			http.SetCookie(w, &http.Cookie{
				Name:  cookieName,
				Value: cookieValue,
				Path:  "/",
			})
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()

	// First request - should receive cookie
	var resp1 interface{}
	err := client.GET(ctx, server.URL, nil, &resp1)
	assert.NoError(t, err)

	// Second request - should send cookie
	var resp2 interface{}
	err = client.GET(ctx, server.URL, nil, &resp2)

	// Assert
	assert.NoError(t, err, "second GET request should not fail")
}

type logMarshalerStruct struct {
	Secret   string `json:"secret"`
	Username string `json:"username"`
}

func (l *logMarshalerStruct) LogString() string {
	return fmt.Sprintf(`{"username":"%s"}`, l.Username)
}

func TestHTTPClientSendRequestWhenNilBodyThenSendsRequestWithoutBody(t *testing.T) {
	// Arrange
	var receivedMethod string
	var receivedBodyLength int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedBodyLength = r.ContentLength
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":{},"error":{"code":0}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()

	// Act
	var resp interface{}
	err := client.POST(ctx, server.URL, nil, &resp)

	// Assert
	assert.NoError(t, err, "POST request with nil body should not fail")
	assert.Equal(t, http.MethodPost, receivedMethod, "method should be POST")
	assert.Equal(t, int64(0), receivedBodyLength, "body length should be 0")
}

func TestHTTPClientSendRequestWhenValidBodyThenSendsRequestWithBody(t *testing.T) {
	// Arrange
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "content-type should be application/json")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{},"error":{"code":0}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()
	body := map[string]string{"key": "value"}

	// Act
	var resp interface{}
	err := client.POST(ctx, server.URL, body, &resp)

	// Assert
	assert.NoError(t, err, "POST request with body should not fail")
	assert.Contains(t, string(receivedBody), "key", "body should contain key")
	assert.Contains(t, string(receivedBody), "value", "body should contain value")
}

func TestHTTPClientSendRequestWhenLogMarshalerThenUsesLogString(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":{},"error":{"code":0}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()
	body := &logMarshalerStruct{
		Secret:   "secret-123",
		Username: "admin",
	}

	// Act
	var resp interface{}
	err := client.POST(ctx, server.URL, body, &resp)

	// Assert
	assert.NoError(t, err, "POST request with LogMarshaler body should not fail")
}

func TestHTTPClientSendRequestWhenGETWithNilBodyThenSendsWithoutBody(t *testing.T) {
	// Arrange
	var receivedMethod string
	var receivedBodyLength int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedBodyLength = r.ContentLength
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":{},"error":{"code":0}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()

	// Act
	var resp interface{}
	err := client.GET(ctx, server.URL, nil, &resp)

	// Assert
	assert.NoError(t, err, "GET request should not fail")
	assert.Equal(t, http.MethodGet, receivedMethod, "method should be GET")
	assert.Equal(t, int64(0), receivedBodyLength, "body length should be 0 (no body)")
}

func TestHTTPClientSendRequestWhenDELETEWithNilBodyThenSendsWithoutBody(t *testing.T) {
	// Arrange
	var receivedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":{},"error":{"code":0}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient("", nil)
	ctx := context.Background()

	// Act
	var resp interface{}
	err := client.DELETE(ctx, server.URL, nil, &resp)

	// Assert
	assert.NoError(t, err, "DELETE request should not fail")
	assert.Equal(t, http.MethodDelete, receivedMethod, "method should be DELETE")
}

// mockHTTPClient is a mock implementation of HTTPClient for testing
type mockHTTPClient struct {
	responseFunc     func() *http.Response
	errFunc          func() error
	setHeadersCalled map[string]string
}

func (m *mockHTTPClient) GET(ctx context.Context, url string, queryParams map[string]string, ret interface{}) error {
	return m.mock(ret)
}

func (m *mockHTTPClient) POST(ctx context.Context, url string, body interface{}, ret interface{}) error {
	return m.mock(ret)
}

func (m *mockHTTPClient) PUT(ctx context.Context, url string, body interface{}, ret interface{}) error {
	return m.mock(ret)
}

func (m *mockHTTPClient) DELETE(ctx context.Context, url string, params map[string]string, ret interface{}) error {
	return m.mock(ret)
}

func (m *mockHTTPClient) mock(ret interface{}) error {
	if m.errFunc != nil {
		return m.errFunc()
	}
	resp := m.responseFunc()
	if resp != nil && ret != nil {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(ret)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mockHTTPClient) SetHeaders(headers map[string]string) {
	if m.setHeadersCalled == nil {
		m.setHeadersCalled = make(map[string]string)
	}
	for k, v := range headers {
		m.setHeadersCalled[k] = v
	}
}
