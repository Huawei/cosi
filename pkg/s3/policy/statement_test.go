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

// Package policy helps to process the data structure of bucket policy
package policy

import (
	"reflect"
	"testing"
)

func Test_Statement_Build_Case(t *testing.T) {
	// arrange
	sid := "sid-test"
	e := EffectAllow
	userArn := "arn:aws:iam::domain-id:user/user-name"
	pricipal := map[string][]string{awsPrinciple: {"arn:aws:iam::domain-id:user/user-name"}}
	ac := AllowedReadActions
	bucketName := "bucket-name"
	resources := []string{"arn:aws:s3:::bucket-name", "arn:aws:s3:::bucket-name/*"}

	wantStatement := &Statement{
		Sid:       sid,
		Effect:    e,
		Principal: pricipal,
		Action:    ac,
		Resource:  resources,
	}

	// act
	gotStatement := NewStatementBuilder().
		WithSID(sid).
		WithEffect(e).
		WithPrincipals(userArn).
		WithActions(ac).
		WithResources(bucketName).
		WithSubResources(bucketName).
		Build()

	// assert
	if !reflect.DeepEqual(gotStatement, wantStatement) {
		t.Errorf("Test_Statement_Build_Case failed, gotStatement= [%v], "+
			"wantStatement= [%v]", gotStatement, wantStatement)
	}
}
