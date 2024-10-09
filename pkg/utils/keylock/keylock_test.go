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

// Package keylock provides locks based on key values
package keylock

import (
	"testing"
)

func Test_KeyLock_Concurrent_Scenario_Success(t *testing.T) {
	// arrange
	keyLock := NewKeyLock(100)
	sign := make(chan struct{}, 10)
	gotRes := 0

	adder := func(value int, key string) {
		defer func() {
			sign <- struct{}{}
		}()

		keyLock.Lock(key)
		defer keyLock.Unlock(key)

		gotRes += value
	}

	subtractor := func(value int, key string) {
		defer func() {
			sign <- struct{}{}
		}()

		keyLock.Lock(key)
		defer keyLock.Unlock(key)

		gotRes -= value
	}
	wantRes := 0
	num := 100000
	lockKeyValue := "key-value"

	// act
	for i := 0; i < num/2; i++ {
		go adder(1, lockKeyValue)
		go subtractor(1, lockKeyValue)
	}
	for i := 0; i < num; i++ {
		<-sign
	}

	// assert
	if gotRes != wantRes {
		t.Errorf("Test_KeyLock_Concurrent_Scenario_Success failed, gotRes= [%v], wantRes= [%v]", gotRes, wantRes)
	}
}
