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

import "encoding/json"

// BucketPolicy represents set of policy statements for a single bucket.
type BucketPolicy struct {
	// Id identifies the bucket policy, optional
	Id string `json:"Id"`

	// Version is the version of the BucketPolicy data structure
	// should always be '2012-10-17'
	Version string `json:"Version"`

	// Statement is the bucket policy statement
	Statement []Statement `json:"Statement"`
}

// NewBucketPolicy returns a new BucketPolicy with given Statement
func NewBucketPolicy(ps ...Statement) *BucketPolicy {
	return &BucketPolicy{
		Version:   version,
		Statement: append([]Statement{}, ps...),
	}
}

// ToJsonString is used to marshal bucket policy to json string format
func (bp *BucketPolicy) ToJsonString() (string, error) {
	b, err := json.Marshal(bp)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ModifyStatement is used to update targeted existing statement in the policy,
// if targeted statement not exist, add it.
// Match key is statement sid.
func (bp *BucketPolicy) ModifyStatement(newPs Statement) *BucketPolicy {
	var match bool
	for j, oldPs := range bp.Statement {
		if newPs.Sid == oldPs.Sid {
			bp.Statement[j] = newPs
			match = true
		}
	}

	if !match {
		bp.Statement = append(bp.Statement, newPs)
	}

	return bp
}

// RemoveStatement is used to remove targeted statement with specified sid.
// Sid is unique in statements.
// Return a new bucket policy.
func (bp *BucketPolicy) RemoveStatement(sid string) *BucketPolicy {
	newBp := NewBucketPolicy()
	for _, statement := range bp.Statement {
		if statement.Sid != sid {
			newBp.Statement = append(newBp.Statement, statement)
		}
	}

	return newBp
}
