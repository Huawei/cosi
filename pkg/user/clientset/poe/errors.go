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

import (
	"encoding/xml"
	"fmt"
)

const (
	// errNoSuchUser means user not exist
	errNoSuchUser errorReason = "NoSuchEntity"
	// errNoSuchUserAccess means user access not exist
	errNoSuchUserAccess errorReason = "NoSuchEntity"
)

// errorReason is the reason of the error
type errorReason string

func handleErrorResponse(errBody []byte) error {
	resp := errorResponse{}
	err := xml.Unmarshal(errBody, &resp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal poe http error response, "+
			"unmarshal error is [%v], errBody is [%s]", err, string(errBody))
	}

	return resp
}

func (e errorReason) Error() string {
	return string(e)
}

// Is determines whether the error is known to be reported
func (e errorResponse) Is(target error) bool {
	return target == errorReason(e.CodeError.Code)
}

// Error returns non-empty string if there was an error.
func (e errorResponse) Error() string {
	return fmt.Sprintf("error Response: code is [%s], msg is [%s], "+
		"requestId is [%s]", e.CodeError.Code, e.CodeError.Message, e.RequestId)
}
