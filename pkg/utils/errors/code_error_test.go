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

// Package errors provides customize error
package errors

import (
	"fmt"
	"testing"
)

func TestIsResourceNotExistErr_True(t *testing.T) {
	// arrange
	err := NewResourceNotExistErr("mock-err")

	// act
	got := IsResourceNotExistErr(err)

	// assert
	if got != true {
		t.Errorf("TestIsResourceNotExistErr_True failed, got= [%v], want= true", got)
	}
}

func TestIsResourceNotExistErr_False(t *testing.T) {
	// arrange
	err := fmt.Errorf("mock-err")

	// act
	got := IsResourceNotExistErr(err)

	// assert
	if got != false {
		t.Errorf("TestIsResourceNotExistErr_False failed, got= [%v], want= false", got)
	}
}
