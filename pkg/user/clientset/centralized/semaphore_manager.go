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
	"os"
	"strconv"
	"sync"
)

const (
	defaultMaxConcurrent        = 30
	maxConcurrentPerEndpointEnv = "MAX_CONCURRENT_PER_ENDPOINT"
)

// SemaphoreManager provides global semaphore management for concurrent access control.
// It ensures that each endpoint has an independent semaphore shared by all clients
// using that endpoint.
type SemaphoreManager struct {
	semaphores map[string]chan struct{}
	mutex      sync.RWMutex
}

var (
	semaphoreManagerInstance *SemaphoreManager
	semaphoreManagerOnce     sync.Once
)

// GetSemaphoreManager returns the global singleton SemaphoreManager instance.
func GetSemaphoreManager() *SemaphoreManager {
	semaphoreManagerOnce.Do(func() {
		semaphoreManagerInstance = &SemaphoreManager{
			semaphores: make(map[string]chan struct{}),
		}
	})
	return semaphoreManagerInstance
}

// GetSemaphore retrieves or creates a semaphore for the given endpoint.
// If the endpoint already has a semaphore, it returns the existing one.
// If the endpoint is new, it creates a semaphore with the specified max size.
// This method is thread-safe.
func (m *SemaphoreManager) GetSemaphore(endpoint string, maxSize int) chan struct{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if semaphore, exists := m.semaphores[endpoint]; exists {
		return semaphore
	}

	semaphore := make(chan struct{}, maxSize)
	for i := 0; i < maxSize; i++ {
		semaphore <- struct{}{}
	}

	m.semaphores[endpoint] = semaphore
	return semaphore
}

// GetMaxConcurrent returns the maximum concurrent connections per endpoint.
// It follows the priority order:
// 1. Secret-injected MaxConcurrent value
// 2. Environment variable MAX_CONCURRENT_PER_ENDPOINT
// 3. Default value (30)
func GetMaxConcurrent(secretMaxConcurrent int) int {
	if secretMaxConcurrent > 0 {
		return secretMaxConcurrent
	}

	envValue := os.Getenv(maxConcurrentPerEndpointEnv)
	if envValue != "" {
		if value, err := strconv.Atoi(envValue); err == nil && value > 0 {
			return value
		}
	}

	return defaultMaxConcurrent
}
