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

// Package user provides user clients and apis
package user

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/cosi-driver/pkg/user/clientset/poe"
)

func Test_NewUserClient_PoeType(t *testing.T) {
	// arrange
	c := Config{ClientType: PoeType}
	userApi := &poe.Client{}

	// mock
	patches := gomonkey.ApplyFuncReturn(poe.NewPoeClient, userApi, nil)

	// act
	gotApi, gotErr := NewUserClient(c)

	// assert
	if !reflect.DeepEqual(gotApi, userApi) || gotErr != nil {
		t.Errorf("Test_NewUserClient_PoeType failed, gotApi= [%v], wantApi= [%v], "+
			"gotErr= [%v], wantErr= [%v]", gotApi, userApi, gotErr, nil)
	}

	// cleanup
	t.Cleanup(func() {
		patches.Reset()
	})
}

func Test_NewUserClient_Default(t *testing.T) {
	// arrange
	c := Config{ClientType: "default"}
	wantErr := fmt.Errorf("unknown user client type [%s]", c.ClientType)

	// act
	gotApi, gotErr := NewUserClient(c)

	// assert
	if gotApi != nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("Test_NewUserClient_Default failed, gotApi= [%v], wantApi= [%v], "+
			"gotErr= [%v], wantErr= [%v]", gotApi, nil, gotErr, wantErr)
	}
}
