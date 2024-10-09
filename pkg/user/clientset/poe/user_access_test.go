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
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/cosi-driver/pkg/user/api"
)

func TestClient_CreateUserAccess_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.CreateUserAccessInput{}
	body := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<CreateAccessKeyResponse>\n" +
			"<CreateAccessKeyResult>\n" +
			"<AccessKey>\n" +
			"<UserName>zf-test-12</UserName>\n" +
			"<AccountId>00000190BFBAEB103A8B923C8FD0CA09</AccountId>\n" +
			"<AccessKeyId>B2500CF822659A30C219</AccessKeyId>\n" +
			"<Status>Active</Status>\n" +
			"<SecretAccessKey>3uo0D4sAo8UHAUUKr9lDs/IYdssAAAGRImWaMCWR</SecretAccessKey>\n" +
			"<CreateDate>2024-08-05T11:56:09.648Z</CreateDate>\n" +
			"</AccessKey>\n" +
			"</CreateAccessKeyResult>\n" +
			"<ResponseMetadata>\n" +
			"<RequestId>d11fd236-8897-438d-903f-2fcdd2a6f738</RequestId>\n" +
			"</ResponseMetadata>\n" +
			"</CreateAccessKeyResponse>")

	want := &api.CreateUserAccessOutput{
		AccessKeyId:     "B2500CF822659A30C219",
		SecretAccessKey: "3uo0D4sAo8UHAUUKr9lDs/IYdssAAAGRImWaMCWR",
	}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return body, nil
		})

	// act
	got, gotErr := c.CreateUserAccess(ctx, in)

	// assert
	if !reflect.DeepEqual(want, got) || gotErr != nil {
		t.Errorf("TestClient_CreateUserAccess_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_DeleteUserAccess_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.DeleteUserAccessInput{}
	body := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<DeleteAccessKeyResponse>\n" +
			"<ResponseMetadata>\n" +
			"<RequestId>38284051-12e5-4a0d-bba3-c2fdca506405</RequestId>\n" +
			"</ResponseMetadata>\n" +
			"</DeleteAccessKeyResponse>")

	want := &api.DeleteUserAccessOutput{}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return body, nil
		})

	// act
	got, gotErr := c.DeleteUserAccess(ctx, in)

	// assert
	if !reflect.DeepEqual(want, got) || gotErr != nil {
		t.Errorf("TestClient_DeleteUserAccess_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_DeleteUserAccess_NotExist(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.DeleteUserAccessInput{UserName: "user-demo", AccessKeyId: "ak-demo"}
	errBody := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<ErrorResponse>\n" +
			"<Error>\n" +
			"<Code>NoSuchEntity</Code>\n" +
			"<Message>The specified credential does not exist, please config credential first.</Message>\n" +
			"</Error>\n" +
			"<RequestId>1ea36501-6917-49a8-8da9-3e9e1c595c5d</RequestId>\n" +
			"</ErrorResponse>")

	mockError := handleErrorResponse(errBody)
	want := &api.DeleteUserAccessOutput{}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return nil, mockError
		})

	// act
	got, gotErr := c.DeleteUserAccess(ctx, in)

	// assert
	if gotErr != nil || !reflect.DeepEqual(want, got) {
		t.Errorf("TestClient_DeleteUserAccess_NotExist failed, gotErr= [%v], wantErr= [%v], "+
			"got= [%v], want= [%v]", gotErr, nil, got, want)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_DeleteUserAccess_InternalFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.DeleteUserAccessInput{UserName: "user-demo", AccessKeyId: "ak-demo"}
	errBody := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<ErrorResponse>\n" +
			"<Error>\n" +
			"<Code>InternalFailed</Code>\n" +
			"<Message>Internal occurs error.</Message>\n" +
			"</Error>\n" +
			"<RequestId>1ea36501-6917-49a8-8da9-3e9e1c595c5d</RequestId>\n" +
			"</ErrorResponse>")

	mockError := handleErrorResponse(errBody)
	wantErr := fmt.Errorf("error Response: code is [InternalFailed], msg is [Internal occurs error.], " +
		"requestId is [1ea36501-6917-49a8-8da9-3e9e1c595c5d]")

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return nil, mockError
		})

	// act
	_, gotErr := c.DeleteUserAccess(ctx, in)

	// assert
	if gotErr.Error() != wantErr.Error() {
		t.Errorf("TestClient_DeleteUserAccess_InternalFailed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_ListUserAccessKeys_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.ListUserAccessKeysInput{}
	body := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<ListAccessKeysResponse>\n" +
			"<ListAccessKeysResult>\n" +
			"<AccessKeyMetadata>\n" +
			"<member>\n" +
			"<UserName>zf-test-12</UserName>\n" +
			"<AccountId>00000190BFBAEB103A8B923C8FD0CA09</AccountId>\n" +
			"<AccessKeyId>B07C0CC035F17D7F98AB</AccessKeyId>\n" +
			"<Status>Active</Status>\n" +
			"<CreateDate>2024-08-09T07:01:44.448Z</CreateDate>\n" +
			"</member>\n" +
			"<member>\n" +
			"<UserName>zf-test-12</UserName>\n" +
			"<AccountId>00000190BFBAEB103A8B923C8FD0CA09</AccountId>\n" +
			"<AccessKeyId>BD080C1B35CD9959A4CA</AccessKeyId>\n" +
			"<Status>Active</Status>\n" +
			"<CreateDate>2024-08-09T06:22:32.281Z</CreateDate>\n" +
			"</member>\n" +
			"</AccessKeyMetadata>\n" +
			"</ListAccessKeysResult>\n" +
			"<ResponseMetadata>\n" +
			"<RequestId>841af781-6da9-4cb1-acb3-dcaef2e3902f</RequestId>\n" +
			"</ResponseMetadata>\n</ListAccessKeysResponse>")

	want := &api.ListUserAccessKeysOutput{
		AccessKeys: []string{"B07C0CC035F17D7F98AB", "BD080C1B35CD9959A4CA"},
	}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return body, nil
		})

	// act
	got, gotErr := c.ListUserAccessKeys(ctx, in)

	// assert
	if !reflect.DeepEqual(want, got) || gotErr != nil {
		t.Errorf("TestClient_ListUserAccessKeys_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_ListUserAccessKeys_NoValues(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.ListUserAccessKeysInput{}
	body := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<ListAccessKeysResponse>\n" +
			"<ListAccessKeysResult>\n" +
			"<AccessKeyMetadata/>\n" +
			"</ListAccessKeysResult>\n" +
			"<ResponseMetadata>\n" +
			"<RequestId>5fc443be-d9e8-4a0e-86b4-2131a4d45402</RequestId>\n" +
			"</ResponseMetadata>\n" +
			"</ListAccessKeysResponse>")

	want := &api.ListUserAccessKeysOutput{
		AccessKeys: nil,
	}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return body, nil
		})

	// act
	got, gotErr := c.ListUserAccessKeys(ctx, in)

	// assert
	if !reflect.DeepEqual(want, got) || gotErr != nil {
		t.Errorf("TestClient_ListUserAccessKeys_NoValues failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_ListUserAccessKeys_NoUser(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.ListUserAccessKeysInput{}
	errBody := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<ErrorResponse>\n" +
			"<Error>\n" +
			"<Code>NoSuchEntity</Code>\n" +
			"<Message>The request was rejected because it referenced a user that does not exist.</Message>\n" +
			"</Error>\n" +
			"<RequestId>1ea36501-6917-49a8-8da9-3e9e1c595c5d</RequestId>\n" +
			"</ErrorResponse>")

	mockError := handleErrorResponse(errBody)
	want := &api.ListUserAccessKeysOutput{}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return nil, mockError
		})

	// act
	got, gotErr := c.ListUserAccessKeys(ctx, in)

	// assert
	if gotErr != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("TestClient_ListUserAccessKeys_NoUser failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
