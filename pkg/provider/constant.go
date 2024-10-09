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

// Package provider providers cosi standard interface
package provider

const (
	// these keys are used in account secret data
	ak       = "accessKey"
	sk       = "secretKey"
	endpoint = "endpoint"
	rootCA   = "rootCA"

	// these keys are used in access/cred secret data
	accessAk = "accessKeyID"
	accessSk = "accessSecretKey"

	// these keys are used in bucketClass/bucketAccessClass parameters
	accountSecretName      = "accountSecretName"
	accountSecretNamespace = "accountSecretNamespace"
	bucketPolicyModel      = "bucketPolicyModel"
	bucketPolicyModelRW    = "rw"
	bucketPolicyModelRO    = "ro"
	bucketACL              = "bucketACL"
	bucketLocation         = "bucketLocation"

	// these keys are protocols
	s3Protocol = "s3"

	// lock related parameters
	keyLockSize = 100
)
