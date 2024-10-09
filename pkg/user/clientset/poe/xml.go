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

// Package poe provides poe client and poe apis
package poe

import "encoding/xml"

type user struct {
	XMLName    xml.Name `xml:"User"`
	UserName   string   `xml:"UserName"`
	Path       string   `xml:"Path"`
	UserID     string   `xml:"UserId"`
	Arn        string   `xml:"Arn"`
	CreateDate string   `xml:"CreateDate"`
}

type responseMetadata struct {
	XMLName   xml.Name `xml:"ResponseMetadata"`
	RequestId string   `xml:"RequestId"`
}

type createUserResponse struct {
	XMLName          xml.Name         `xml:"CreateUserResponse"`
	CreateUserResult createUserResult `xml:"CreateUserResult"`
	ResponseMetadata responseMetadata `xml:"ResponseMetadata"`
}

type createUserResult struct {
	XMLName xml.Name `xml:"CreateUserResult"`
	User    user     `xml:"User"`
}

type getUserResponse struct {
	XMLName          xml.Name         `xml:"GetUserResponse"`
	GetUserResult    getUserResult    `xml:"GetUserResult"`
	ResponseMetadata responseMetadata `xml:"ResponseMetadata"`
}

type getUserResult struct {
	XMLName xml.Name `xml:"GetUserResult"`
	User    user     `xml:"User"`
}

type deleteUserResponse struct {
	XMLName          xml.Name         `xml:"DeleteUserResponse"`
	ResponseMetadata responseMetadata `xml:"ResponseMetadata"`
}

type accessKey struct {
	XMLName         xml.Name `xml:"AccessKey"`
	AccountId       string   `xml:"AccountId"`
	AccessKeyId     string   `xml:"AccessKeyId"`
	Status          string   `xml:"Status"`
	SecretAccessKey string   `xml:"SecretAccessKey"`
	CreateDate      string   `xml:"CreateDate"`
	UserName        string   `xml:"UserName"`
}

type createUserAccessResponse struct {
	XMLName          xml.Name              `xml:"CreateAccessKeyResponse"`
	CreateUserResult createAccessKeyResult `xml:"CreateAccessKeyResult"`
	ResponseMetadata responseMetadata      `xml:"ResponseMetadata"`
}

type createAccessKeyResult struct {
	XMLName   xml.Name  `xml:"CreateAccessKeyResult"`
	AccessKey accessKey `xml:"AccessKey"`
}

type deleteUserAccessResponse struct {
	XMLName          xml.Name         `xml:"DeleteAccessKeyResponse"`
	ResponseMetadata responseMetadata `xml:"ResponseMetadata"`
}

type codeError struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

type errorResponse struct {
	XMLName   xml.Name  `xml:"ErrorResponse"`
	CodeError codeError `xml:"Error"`
	RequestId string    `xml:"RequestId"`
}

type listAccessKeysResponse struct {
	XMLName              xml.Name             `xml:"ListAccessKeysResponse"`
	ListAccessKeysResult listAccessKeysResult `xml:"ListAccessKeysResult"`
	ResponseMetadata     responseMetadata     `xml:"ResponseMetadata"`
}

type listAccessKeysResult struct {
	XMLName           xml.Name          `xml:"ListAccessKeysResult"`
	AccessKeyMetadata accessKeyMetadata `xml:"AccessKeyMetadata"`
}

type accessKeyMetadata struct {
	XMLName xml.Name `xml:"AccessKeyMetadata"`
	Members []member `xml:"member"`
}

type member struct {
	XMLName     xml.Name `xml:"member"`
	AccountId   string   `xml:"AccountId"`
	AccessKeyId string   `xml:"AccessKeyId"`
	Status      string   `xml:"Status"`
	CreateDate  string   `xml:"CreateDate"`
	UserName    string   `xml:"UserName"`
}
