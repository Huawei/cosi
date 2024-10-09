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
	"fmt"
)

type action string

const (
	// S3 Object Operation
	getObject                action = "s3:GetObject"
	putObject                action = "s3:PutObject"
	getObjectVersion         action = "s3:GetObjectVersion"
	deleteObjectVersion      action = "s3:DeleteObjectVersion"
	deleteObject             action = "s3:DeleteObject"
	listMultipartUploadParts action = "s3:ListMultipartUploadParts"
	getObjectAcl             action = "s3:GetObjectAcl"
	getObjectVersionAcl      action = "s3:GetObjectVersionAcl"
	putObjectAcl             action = "s3:PutObjectAcl"
	putObjectVersionAcl      action = "s3:PutObjectVersionAcl"
	abortMultipartUpload     action = "s3:AbortMultipartUpload"

	// S3 Bucket Operation
	listBucketMultiPartUploads action = "s3:ListBucketMultiPartUploads"
	listBucket                 action = "s3:ListBucket"
	listBucketVersions         action = "s3:ListBucketVersions"
)

// AllowedReadActions is a lenient default list of read actions
var AllowedReadActions = []action{
	getObject,
	getObjectVersion,
	listMultipartUploadParts,
	getObjectAcl,
	getObjectVersionAcl,
	listBucketVersions,
	listBucket,
	listBucketMultiPartUploads,
}

// AllowedReadWriteActions is a lenient default list of read and write actions
var AllowedReadWriteActions = []action{
	getObject,
	getObjectVersion,
	listMultipartUploadParts,
	getObjectAcl,
	getObjectVersionAcl,
	listBucketVersions,
	listBucket,
	listBucketMultiPartUploads,
	abortMultipartUpload,
	putObjectAcl,
	deleteObjectVersion,
	putObjectVersionAcl,
	putObject,
	deleteObject,
}

type effect string

const (
	// EffectAllow values are expected by the S3 API to be 'Allow' explicitly
	EffectAllow effect = "Allow"

	// EffectDeny values are expected by the S3 API to be 'Deny' explicitly
	EffectDeny effect = "Deny"

	// the version of the BucketPolicy json structure
	version = "2012-10-17"

	// aws principle
	awsPrinciple = "AWS"

	// arn resource format
	arnResourceFormat = "arn:aws:s3:::%s"
)

// Statement is the Go representation of a bucket policy statement json struct,
// it defines relevant permission controls on a Resource
type Statement struct {
	// Sid is the policy statement's unique identifier, optional
	Sid string `json:"Sid"`

	// Effect determines whether the Action type is 'Allow' or 'Deny'
	Effect effect `json:"Effect"`

	// Principle is the user of arn format affected by this policy statement
	// the format likes 'arn:aws:iam::{accountId}:{userName}'
	Principal map[string][]string `json:"Principal"`

	// Action is a list of s3 actions
	Action []action `json:"Action"`

	// Resource is the ARN identifier for the S3 bucket
	// the format likes 'arn:aws:s3:::{bucket-name}'
	Resource []string `json:"Resource"`
}

// NewStatementBuilder generates a new Policy statement builder.
func NewStatementBuilder() *Statement {
	return &Statement{
		Sid:       "",
		Effect:    "",
		Principal: map[string][]string{},
		Action:    []action{},
		Resource:  []string{},
	}
}

// WithSID add sid to policy statement
func (ps *Statement) WithSID(sid string) *Statement {
	ps.Sid = sid
	return ps
}

// WithPrincipals adds arn format users to the policy statement,
// arn format likes 'arn:aws:iam::%s:user/{user-name}'
func (ps *Statement) WithPrincipals(userArns ...string) *Statement {
	principals := ps.Principal[awsPrinciple]
	for _, u := range userArns {
		principals = append(principals, u)
	}

	ps.Principal[awsPrinciple] = principals
	return ps
}

// WithResources adds arn format buckets to the policy statement,
// arn format likes 'arn:aws:s3:::{bucket-name}'
func (ps *Statement) WithResources(resourceNames ...string) *Statement {
	for _, v := range resourceNames {
		ps.Resource = append(ps.Resource, fmt.Sprintf(arnResourceFormat, v))
	}
	return ps
}

// WithSubResources add all objects inside the bucket to the policy statement,
// arn format likes 'arn:aws:s3:::{bucket-name}/*'
func (ps *Statement) WithSubResources(resourceNames ...string) *Statement {
	var subresource string
	for _, v := range resourceNames {
		subresource = fmt.Sprintf("%s/*", v)
		ps.Resource = append(ps.Resource, fmt.Sprintf(arnResourceFormat, subresource))
	}
	return ps
}

// WithEffect sets the effect to policy statement's Actions
func (ps *Statement) WithEffect(e effect) *Statement {
	ps.Effect = e
	return ps
}

// WithActions sets s3 actions for the policy statement
func (ps *Statement) WithActions(actions []action) *Statement {
	ps.Action = actions
	return ps
}

// Build return assembled statement
func (ps *Statement) Build() *Statement {
	return ps
}
