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

// Package utils provides a lot of utility function
package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"runtime/debug"
	"sort"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/huawei/cosi-driver/pkg/utils/log"
)

// HmacSha256 gets hmac sha256 value of input
func HmacSha256(key, value []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(value)
	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}

// GetSortedUrlQueryString is used to get url query str after being sorted
func GetSortedUrlQueryString(param map[string]string) string {
	var list []string
	var keys []string

	for k, _ := range param {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := param[key]
		list = append(list, key+"="+url.QueryEscape(value))
	}

	return strings.Join(list, "&")
}

// GetKubeConfig is used to get kube config by path, if path is "", then return inCluster config
func GetKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	var kubeConfig *rest.Config
	var err error

	if kubeConfigPath != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	} else {
		kubeConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	return kubeConfig, nil
}

// RecoverPanic used to recover panic
func RecoverPanic(ctx context.Context) {
	if re := recover(); re != nil {
		log.AddContext(ctx).Errorf("panic message is [%s], panic stack is [%s]", re, debug.Stack())
	}
}

// ContainsElement is used to determine whether slice contains specified element
func ContainsElement(elements []string, target string) bool {
	for _, element := range elements {
		if element == target {
			return true
		}
	}

	return false
}

// BuildTLSConfig is used to build tls config
func BuildTLSConfig(rootCA []byte) (*tls.Config, error) {
	var tlsConfig *tls.Config

	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	if len(rootCA) != 0 {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(rootCA)
		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            caCertPool,
			MinVersion:         tls.VersionTLS12,
		}
	}

	return tlsConfig, nil
}
