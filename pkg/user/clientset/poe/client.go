/*
 Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

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

// Package poe provides poe client and poe apis
package poe

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/huawei/cosi-driver/pkg/utils"
	"github.com/huawei/cosi-driver/pkg/utils/log"
)

const (
	connectionTimeout   = time.Second * 60
	actionKey           = "Action"
	aWSAccessKeyIdKey   = "AWSAccessKeyId"
	signatureMethodKey  = "SignatureMethod"
	signatureVersionKey = "SignatureVersion"
	signatureKey        = "Signature"
	timestampKey        = "Timestamp"
	userNameKey         = "UserName"
	accessKeyIdKey      = "AccessKeyId"

	hmacSHA256           = "HmacSHA256"
	signatureVersionFour = "4"

	poeURI       = "/poe/rest"
	poePort      = "9443"
	portSeprator = ":"
	pathSeprator = "//"
)

// HttpClient interface that conforms to that of the http package's Client
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client struct for poe client
type Client struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	RootCA     []byte
	HttpClient HttpClient
}

// NewPoeClient returns Client for Poe
func NewPoeClient(endpoint, accessKey, secretKey string, rootCA []byte) (*Client, error) {
	// Validate endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is empty")
	}

	// Check endpoint format, endpoint format likes 'https://x.xx.xx.xx:9443'.
	// Modify Poe client port, its port must be '9443'.
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("url parse endpoint [%s] failed, error is [%v]", endpoint, err)
	}

	// reassemble endpoint with '9443' port
	var eb strings.Builder
	eb.WriteString(u.Scheme)
	eb.WriteString(portSeprator)
	eb.WriteString(pathSeprator)
	eb.WriteString(u.Hostname())
	eb.WriteString(portSeprator)
	eb.WriteString(poePort)
	endpoint = eb.String()

	// Validate access key
	if accessKey == "" {
		return nil, fmt.Errorf("access key is empty")
	}

	// Validate secret key
	if secretKey == "" {
		return nil, fmt.Errorf("secret key is empty")
	}

	tlsConfig, err := utils.BuildTLSConfig(rootCA)
	if err != nil {
		return nil, fmt.Errorf("build tls config failed, error is [%v]", err)
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{
		Timeout:   connectionTimeout,
		Transport: tr,
	}

	return &Client{
		Endpoint:   endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		RootCA:     rootCA,
		HttpClient: httpClient,
	}, nil
}

// Call makes request to the storage by poe Client.
// poe Client call only support Get httpMethod.
func (pec *Client) Call(ctx context.Context, param map[string]string) (body []byte, err error) {
	if param == nil {
		return nil, fmt.Errorf("enter param is empty")
	}

	// add public auth parameters
	param[aWSAccessKeyIdKey] = pec.AccessKey
	param[signatureMethodKey] = hmacSHA256
	param[signatureVersionKey] = signatureVersionFour
	param[timestampKey] = time.Now().UTC().Format("2006-01-02T15:04:05.0Z")

	// get signature
	stringToSign := pec.getStringToSign(http.MethodGet, strings.Split(pec.Endpoint, "//")[1],
		poeURI, utils.GetSortedUrlQueryString(param))

	signature, err := pec.getSignature(stringToSign)
	if err != nil {
		return nil, err
	}
	param[signatureKey] = signature

	// build request
	var urlBuilder strings.Builder
	urlBuilder.WriteString(pec.Endpoint)
	urlBuilder.WriteString(poeURI)
	urlBuilder.WriteString("?")
	urlBuilder.WriteString(utils.GetSortedUrlQueryString(param))
	request, err := http.NewRequest(http.MethodGet, urlBuilder.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build poe client http request failed, error is [%v]", err)
	}

	// send http request
	resp, err := pec.HttpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("poe client http do failed, error is [%v]", err)
	}
	defer resp.Body.Close()

	// get resp body
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body failed, error is [%v]", err)
	}

	// get err response when code is not success code
	if resp.StatusCode != http.StatusOK {
		log.AddContext(ctx).Errorf("poe client http call not success, http status code is [%v]", resp.StatusCode)
		return nil, handleErrorResponse(body)
	}

	return body, nil
}

func (pec *Client) getStringToSign(method, host, uri, paramString string) string {
	return method + "\n" + host + "\n" + uri + "\n" + paramString
}

func (pec *Client) getSignature(stringToSign string) (string, error) {
	b, err := utils.HmacSha256([]byte(pec.SecretKey), []byte(stringToSign))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
