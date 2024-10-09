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

func TestClient_CreateUser_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.CreateUserInput{}
	body := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<CreateUserResponse>\n" +
			"<CreateUserResult>\n" +
			"<User>\n" +
			"<Path>/</Path>\n" +
			"<UserName>zf-test-14</UserName>\n" +
			"<UserId>00000191224B7D1F3A893E889C135BBA</UserId>\n" +
			"<Arn>arn:aws:iam::3059394579:user/zf-test-14</Arn>\n" +
			"<CreateDate>2024-08-05T11:27:38.271Z</CreateDate>\n" +
			"</User>\n" +
			"</CreateUserResult>\n" +
			"<ResponseMetadata>\n" +
			"<RequestId>86a0f4ff-fc13-4263-b1a4-8c80124f57e9</RequestId>\n" +
			"</ResponseMetadata>\n</CreateUserResponse>")

	want := &api.CreateUserOutput{
		UserName: "zf-test-14",
		UserID:   "00000191224B7D1F3A893E889C135BBA",
		Arn:      "arn:aws:iam::3059394579:user/zf-test-14",
	}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return body, nil
		})

	// act
	got, gotErr := c.CreateUser(ctx, in)

	// assert
	if !reflect.DeepEqual(want, got) || gotErr != nil {
		t.Errorf("TestClient_CreateUser_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_DeleteUser_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.DeleteUserInput{}
	body := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<DeleteUserResponse>\n" +
			"<ResponseMetadata>\n" +
			"<RequestId>38284051-12e5-4a0d-bba3-c2fdca506405</RequestId>\n" +
			"</ResponseMetadata>\n" +
			"</DeleteUserResponse>")

	want := &api.DeleteUserOutput{}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return body, nil
		})

	// act
	got, gotErr := c.DeleteUser(ctx, in)

	// assert
	if !reflect.DeepEqual(want, got) || gotErr != nil {
		t.Errorf("TestClient_DeleteUser_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_GetUser_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.GetUserInput{}
	body := []byte(
		"<?xml version=\"1.0\"?>\n" +
			"<GetUserResponse>\n" +
			"<GetUserResult>\n" +
			"<User>\n" +
			"<Path>/</Path>\n" +
			"<UserName>zf-test-12</UserName>\n" +
			"<UserId>00000191224988653A8931B8E451B937</UserId>\n" +
			"<Arn>arn:aws:iam::3059394579:user/zf-test-12</Arn>\n" +
			"<CreateDate>2024-08-05T11:25:30.085Z</CreateDate>\n" +
			"</User>\n" +
			"</GetUserResult>\n" +
			"<ResponseMetadata>\n" +
			"<RequestId>79155879-c9b0-4728-8b23-4c2be50d7a73</RequestId>\n" +
			"</ResponseMetadata>\n</GetUserResponse>")

	want := &api.GetUserOutput{
		UserName: "zf-test-12",
		UserID:   "00000191224988653A8931B8E451B937",
		Arn:      "arn:aws:iam::3059394579:user/zf-test-12",
	}

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return body, nil
		})

	// act
	got, gotErr := c.GetUser(ctx, in)

	// assert
	if !reflect.DeepEqual(want, got) || gotErr != nil {
		t.Errorf("TestClient_GetUser_Success failed, got= [%v], want= [%v], "+
			"gotErr= [%v], wantErr= nil", got, want, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_GetUser_NotExist(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.GetUserInput{UserName: "user-demo"}
	errBody := []byte(
		"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n" +
			"<ErrorResponse>\n" +
			"<Error>\n" +
			"<Code>NoSuchEntity</Code>\n" +
			"<Message>The request was rejected because it referenced a user that does not exist.</Message>\n" +
			"</Error>\n" +
			"<RequestId>5e8141b8-601d-460b-af2d-dea105442f26</RequestId>\n" +
			"</ErrorResponse>\n")

	mockError := handleErrorResponse(errBody)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return nil, mockError
		})

	// act
	got, gotErr := c.GetUser(ctx, in)

	// assert
	if got != nil || gotErr != nil {
		t.Errorf("TestClient_GetUser_NotExist failed, got= [%v], want= nil, "+
			"gotErr= [%v], wantErr= nil", got, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_DeleteUser_NotExist(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.DeleteUserInput{UserName: "user-demo"}
	errBody := []byte(
		"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n" +
			"<ErrorResponse>\n" +
			"<Error>\n" +
			"<Code>NoSuchEntity</Code>\n" +
			"<Message>The request was rejected because it referenced a user that does not exist.</Message>\n" +
			"</Error>\n" +
			"<RequestId>5e8141b8-601d-460b-af2d-dea105442f26</RequestId>\n" +
			"</ErrorResponse>\n")
	mockError := handleErrorResponse(errBody)

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return nil, mockError
		})

	// act
	_, gotErr := c.DeleteUser(ctx, in)

	// assert
	if gotErr != nil {
		t.Errorf("TestClient_DeleteUser_NotExist failed, gotErr= [%v], wantErr= [%v]", gotErr, nil)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestClient_DeleteUser_InternalFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	c := &Client{}
	in := &api.DeleteUserInput{UserName: "user-demo"}
	errBody := []byte(
		"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n" +
			"<ErrorResponse>\n" +
			"<Error>\n" +
			"<Code>InternalFailed</Code>\n" +
			"<Message>Internal occurs error.</Message>\n" +
			"</Error>\n" +
			"<RequestId>5e8141b8-601d-460b-af2d-dea105442f26</RequestId>\n" +
			"</ErrorResponse>\n")
	mockError := handleErrorResponse(errBody)
	wantErr := fmt.Errorf("error Response: code is [InternalFailed], msg is [Internal occurs error.], " +
		"requestId is [5e8141b8-601d-460b-af2d-dea105442f26]")

	// mock
	mock := gomonkey.ApplyMethod(reflect.TypeOf(c), "Call",
		func(_ *Client, ctx context.Context, param map[string]string) ([]byte, error) {
			return nil, mockError
		})

	// act
	_, gotErr := c.DeleteUser(ctx, in)

	// assert
	if gotErr.Error() != wantErr.Error() {
		t.Errorf("TestClient_DeleteUser_InternalFailed failed, gotErr= [%v], wantErr= [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
