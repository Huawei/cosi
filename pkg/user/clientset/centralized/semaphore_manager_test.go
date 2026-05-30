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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSemaphoreManagerReturnsSameInstance(t *testing.T) {
	// Arrange
	var wg sync.WaitGroup
	var instances []*SemaphoreManager
	var mu sync.Mutex
	totals := 10
	wg.Add(totals)

	// Act
	for i := 0; i < totals; i++ {
		go func() {
			defer wg.Done()
			instance := GetSemaphoreManager()
			mu.Lock()
			instances = append(instances, instance)
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Assert
	first := instances[0]
	for _, instance := range instances[1:] {
		assert.Same(t, first, instance, "all instances should be the same")
	}
}

func TestSemaphoreManagerGetSemaphoreCreatesSemaphore(t *testing.T) {
	// Arrange
	manager := &SemaphoreManager{
		semaphores: make(map[string]chan struct{}),
	}
	endpoint := "https://endpoint1:8088"
	maxSize := 30

	// Act
	semaphore := manager.GetSemaphore(endpoint, maxSize)

	// Assert
	assert.NotNil(t, semaphore, "semaphore should not be nil")
	assert.Equal(t, maxSize, len(semaphore), "semaphore size should match maxSize")
}

func TestSemaphoreManagerGetSemaphoreReturnsSameForSameEndpoint(t *testing.T) {
	// Arrange
	manager := &SemaphoreManager{
		semaphores: make(map[string]chan struct{}),
	}
	endpoint := "https://endpoint1:8088"
	maxSize := 30

	// Act
	semaphore1 := manager.GetSemaphore(endpoint, maxSize)
	semaphore2 := manager.GetSemaphore(endpoint, maxSize)

	// Assert
	assert.Equal(t, semaphore1, semaphore2, "should return same semaphore instance")
}

func TestGetMaxConcurrentWithSecret(t *testing.T) {
	// Arrange
	secretMaxConcurrent := 50

	// Act
	result := GetMaxConcurrent(secretMaxConcurrent)

	// Assert
	assert.Equal(t, secretMaxConcurrent, result, "should return secret value")
}

func TestGetMaxConcurrentReturnsDefault(t *testing.T) {
	// Arrange
	secretMaxConcurrent := 0

	// Act
	result := GetMaxConcurrent(secretMaxConcurrent)

	// Assert
	assert.Equal(t, defaultMaxConcurrent, result, "should return default value")
}

func TestSemaphoreManagerGetSemaphoreConcurrent(t *testing.T) {
	// Arrange
	manager := &SemaphoreManager{
		semaphores: make(map[string]chan struct{}),
	}
	endpoint := "https://endpoint1:8088"
	maxSize := 30
	var wg sync.WaitGroup

	// Act
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.GetSemaphore(endpoint, maxSize)
		}()
	}
	wg.Wait()

	// Assert
	assert.NotNil(t, manager.semaphores[endpoint], "semaphore should exist")
}
