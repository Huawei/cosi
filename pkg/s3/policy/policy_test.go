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

func Test_BucketPolicy_ModifyStatement_Replace(t *testing.T) {
	// arrange
	sid := "sid-test"
	efe := EffectAllow
	pricipal := map[string][]string{awsPrinciple: {"arn:aws:iam::domain-id:user/user-name-1"}}
	existStatement := Statement{
		Sid:       sid,
		Effect:    efe,
		Principal: pricipal,
	}
	bp := &BucketPolicy{Statement: []Statement{existStatement}}

	newPricipal := map[string][]string{awsPrinciple: {"arn:aws:iam::domain-id:user/user-name-2"}}
	newEfe := EffectDeny
	newStatement := Statement{
		Sid:       sid,
		Effect:    newEfe,
		Principal: newPricipal,
	}
	wantBp := &BucketPolicy{Statement: []Statement{newStatement}}

	// act
	gotBp := bp.ModifyStatement(newStatement)

	// assert
	if !reflect.DeepEqual(gotBp, wantBp) {
		t.Errorf("Test_BucketPolicy_ModifyStatement_Replace failed, gotBp= [%v], wantBp= [%v]", gotBp, wantBp)
	}
}

func Test_BucketPolicy_ModifyStatement_Add(t *testing.T) {
	// arrange
	sid := "sid-test-1"
	existStatement := Statement{
		Sid: sid,
	}
	bp := &BucketPolicy{Statement: []Statement{existStatement}}

	newSid := "sid-test-2"
	newStatement := Statement{
		Sid: newSid,
	}
	wantBp := &BucketPolicy{Statement: []Statement{existStatement, newStatement}}

	// act
	gotBp := bp.ModifyStatement(newStatement)

	// assert
	if !reflect.DeepEqual(gotBp, wantBp) {
		t.Errorf("Test_BucketPolicy_ModifyStatement_Add failed, gotBp= [%v], wantBp= [%v]", gotBp, wantBp)
	}
}

func Test_BucketPolicy_RemoveStatement_TargetRemoveSuccess(t *testing.T) {
	// arrange
	sid := "sid-test-1"
	existStatement := Statement{
		Sid: sid,
	}
	bp := &BucketPolicy{Statement: []Statement{existStatement}}
	wantBp := NewBucketPolicy()

	// act
	gotBp := bp.RemoveStatement(sid)

	// assert
	if !reflect.DeepEqual(gotBp, wantBp) {
		t.Errorf("Test_BucketPolicy_RemoveStatement_TargetRemoveSuccess failed, gotBp= [%v], wantBp= [%v]", gotBp, wantBp)
	}
}

func Test_BucketPolicy_RemoveStatement_TargetNotExist(t *testing.T) {
	// arrange
	sid := "sid-test-1"
	statement := Statement{
		Sid: sid,
	}
	bp := NewBucketPolicy(statement)
	wantBp := NewBucketPolicy(statement)
	notExistSid := "sid-not-exist"

	// act
	gotBp := bp.RemoveStatement(notExistSid)

	// assert
	if !reflect.DeepEqual(gotBp, wantBp) {
		t.Errorf("Test_BucketPolicy_RemoveStatement_TargetNotExist failed, gotBp= [%v], wantBp= [%v]", gotBp, wantBp)
	}
}
