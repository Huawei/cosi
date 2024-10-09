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

// Package errors provides s3 errors
package errors

import "github.com/aws/aws-sdk-go/service/s3"

const (
	// ErrNoSuchBucketPolicy is the s3 err about bucket policy not exist
	ErrNoSuchBucketPolicy = "NoSuchBucketPolicy"

	// ErrNoSuchBucket is the s3 err about bucket not exist
	ErrNoSuchBucket = s3.ErrCodeNoSuchBucket
)

var (
	// EmptyExceptionalErrCodes return empty exceptional error codes
	EmptyExceptionalErrCodes = []string{}
)

// NewExceptionalErrCodes return exceptional error codes
func NewExceptionalErrCodes(codes ...string) []string {
	return codes
}
