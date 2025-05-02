package chat_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"kseli/auth"
	"kseli/common"
	"kseli/config"
	"kseli/features/chat"
	"kseli/router"
)

type joinEnv struct {
	roomID      string
	invitetoken string
	adminToken  string
	mux         *http.ServeMux
}

func newJoinEnv(t *testing.T) *joinEnv {
	config.APIKey = "test-api-key"
	mux := router.New()

	// 1) Create the room
	createReqBody, _ := json.Marshal(chat.CreateRoomRequest{
		Username:        "admin",
		MaxParticipants: 2,
	})
	createReqHeaders := map[string]string{
		"Origin":                   "http://kseli.app",
		"X-Api-Key":                config.APIKey,
		"X-Participant-Session-Id": "admin-session-id",
	}
	createStatus, createRespBody := sendRequest(mux, http.MethodPost, "/api/rooms", bytes.NewReader(createReqBody), createReqHeaders)
	if createStatus != http.StatusCreated {
		t.Fatalf("newJoinEnv - expected create 201, got %d, body: %s", createStatus, string(createRespBody))
	}
	var createRespStruct chat.CreateRoomResponse
	if err := json.Unmarshal(createRespBody, &createRespStruct); err != nil {
		t.Fatalf("newJoinEnv - failed to unmarshal: %v", err)
	}

	// 2) Fetch invite link via GetRoomHandler as an admin
	getReqHeaders := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": createRespStruct.Token,
	}
	getStatus, getRespBody := sendRequest(mux, http.MethodGet, "/api/rooms/"+createRespStruct.RoomID, nil, getReqHeaders)
	if getStatus != http.StatusOK {
		t.Fatalf("newJoinEnv - expected get 200, got %d, body: %s", getStatus, string(getRespBody))
	}
	var gerRespStruct chat.GetRoomResponse
	if err := json.Unmarshal(getRespBody, &gerRespStruct); err != nil {
		t.Fatalf("newJoinEnv - failed to unmarshal: %v", err)
	}

	parts := strings.Split(gerRespStruct.InviteLink, "?invite=")
	if len(parts) != 2 {
		t.Fatalf("newJoinEnv - bad invite link %q", gerRespStruct.InviteLink)
	}

	return &joinEnv{
		roomID:      createRespStruct.RoomID,
		invitetoken: parts[1],
		adminToken:  createRespStruct.Token,
		mux:         mux,
	}
}

func Test_JoinRoom_Success(t *testing.T) {
	env := newJoinEnv(t)

	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	headers := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.invitetoken,
		"X-Participant-Session-Id": "user-session-id",
	}

	status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	expectedStatus := http.StatusCreated
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var respStruct chat.JoinRoomResponse
	if err := json.Unmarshal(respBody, &respStruct); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if respStruct.RoomID == "" || respStruct.Token == "" {
		t.Fatalf("got empty RoomID or Token: %+v", respStruct)
	}
}

func Test_JoinRoom_BadRequests(t *testing.T) {
	env := newJoinEnv(t)

	const maxBody = 128

	type testCase struct {
		name           string
		body           io.Reader
		expectedStatus int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "Bad request: malformed JSON",
			body:           strings.NewReader("{ not json"),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "Bad request: empty body",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "Bad request: type mismatch",
			body:           strings.NewReader(`{"username":123}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name: "Bad request: body too large",
			body: func() io.Reader {
				bigUsername := strings.Repeat("a", maxBody)
				payload := fmt.Sprintf(`{"username":%q}`, bigUsername)
				return strings.NewReader(payload)
			}(),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "Valid request: extra fields allowed",
			body:           strings.NewReader(`{"username":"user","extra":"ignored"}`),
			expectedStatus: http.StatusCreated,
			expectedErrMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{
				"Origin":                   "http://kseli.app",
				"Authorization":            env.invitetoken,
				"X-Participant-Session-Id": "user-session-id",
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", tc.body, headers)

			if status != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d, body: %s", tc.name, tc.expectedStatus, status, string(respBody))
			}

			if tc.expectedStatus != http.StatusCreated {
				var errResp common.ErrorResponse
				if err := json.Unmarshal(respBody, &errResp); err != nil {
					t.Fatalf("[%s] failed to unmarshal: %v", tc.name, err)
				}

				if errResp.Message != tc.expectedErrMsg {
					t.Fatalf("[%s] expected error message %q, got %q", tc.name, tc.expectedErrMsg, errResp.Message)
				}
			}
		})
	}
}

func Test_JoinRoom_RoomNotFound(t *testing.T) {
	env := newJoinEnv(t)

	fakeClaims := auth.InviteClaims{
		RoomID:    "does-not-exist",
		SecretKey: "irrelevant",
		Exp:       time.Now().Add(time.Hour).Unix(),
	}

	fakeToken, _ := auth.CreateToken(fakeClaims)

	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	headers := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            fakeToken,
		"X-Participant-Session-Id": "user-session-id",
	}

	status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	expectedStatus := http.StatusNotFound
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "Chat Room not found."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_InvalidInvite(t *testing.T) {
	env := newJoinEnv(t)

	fakeClaims := auth.InviteClaims{
		RoomID:    env.roomID,
		SecretKey: "invalid",
		Exp:       time.Now().Add(time.Hour).Unix(),
	}

	fakeToken, _ := auth.CreateToken(fakeClaims)

	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	headers := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            fakeToken,
		"X-Participant-Session-Id": "user-session-id",
	}

	status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	expectedStatus := http.StatusForbidden
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "Invalid invite link."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_AlreadyInRoom(t *testing.T) {
	env := newJoinEnv(t)

	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	headers := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.invitetoken,
		"X-Participant-Session-Id": "admin-session-id",
	}

	status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	expectedStatus := http.StatusBadRequest
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "You can not join a room you are already in."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_BannedFromRoom(t *testing.T) {
	env := newJoinEnv(t)

	// 1) Join the room
	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	headers := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.invitetoken,
		"X-Participant-Session-Id": "user-session-id",
	}

	joinStatus1, joinRespBody1 := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	joinExpectedStatus1 := http.StatusCreated
	if joinStatus1 != joinExpectedStatus1 {
		t.Fatalf("Join 1: expected %d, got %d, body: %s", joinExpectedStatus1, joinStatus1, string(joinRespBody1))
	}

	// 2) Get banned from the room
	banBody, _ := json.Marshal(chat.UserRequest{
		TargetUserID: 2,
	})
	banHeaders := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": env.adminToken,
	}

	banStatus, banRespBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/ban", bytes.NewReader(banBody), banHeaders)

	banExpectedStatus := http.StatusNoContent
	if banStatus != banExpectedStatus {
		t.Fatalf("Ban: expected %d, got %d, body: %s", banExpectedStatus, banStatus, string(banRespBody))
	}

	// 3) Try to join the room again
	joinStatus2, joinRespBody2 := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	joinExpectedStatus2 := http.StatusForbidden
	if joinStatus2 != joinExpectedStatus2 {
		t.Fatalf("expected %d, got %d, body: %s", joinExpectedStatus2, joinStatus2, string(joinRespBody2))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(joinRespBody2, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "You are banned from this room."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_RoomFull(t *testing.T) {
	env := newJoinEnv(t)

	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	headers1 := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.invitetoken,
		"X-Participant-Session-Id": "user1-session-id",
	}

	joinStatus1, joinRespBody1 := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers1)

	joinExpectedStatus1 := http.StatusCreated
	if joinStatus1 != joinExpectedStatus1 {
		t.Fatalf("Join 1: expected %d, got %d, body: %s", joinExpectedStatus1, joinStatus1, string(joinRespBody1))
	}

	headers2 := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.invitetoken,
		"X-Participant-Session-Id": "user2-session-id",
	}

	joinStatus2, joinRespBody2 := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers2)

	joinExpectedStatus2 := http.StatusConflict
	if joinStatus2 != joinExpectedStatus2 {
		t.Fatalf("expected %d, got %d, body: %s", joinExpectedStatus2, joinStatus2, string(joinRespBody2))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(joinRespBody2, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "Chat Room is full."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_UsernameTaken(t *testing.T) {
	env := newJoinEnv(t)

	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "admin",
	})
	headers := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.invitetoken,
		"X-Participant-Session-Id": "user-session-id",
	}

	status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	expectedStatus := http.StatusBadRequest
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	errMsg, ok := errResp.FieldErrors["username"]
	if !ok {
		t.Fatal("expected field error for 'username', but none found")
	}

	expectedErrMsg := "This username is taken."
	if errMsg != expectedErrMsg {
		t.Fatalf("expected field error message %q, got %q", expectedErrMsg, errMsg)
	}
}

func Test_JoinRoom_UsernameValidation(t *testing.T) {
	type testCase struct {
		name                  string
		username              string
		expectedStatus        int
		expectedUsernameError string
	}

	tests := []testCase{
		{
			name:                  "Valid username: ASCII 3 chars",
			username:              "abc",
			expectedStatus:        http.StatusCreated,
			expectedUsernameError: "",
		},
		{
			name:                  "Valid username: ASCII 15 chars",
			username:              "abcdefghijklmno",
			expectedStatus:        http.StatusCreated,
			expectedUsernameError: "",
		},
		{
			name:                  "Valid username: UTF-8 3 chars",
			username:              "აბგ",
			expectedStatus:        http.StatusCreated,
			expectedUsernameError: "",
		},
		{
			name:                  "Valid username: UTF-8 15 chars",
			username:              "აბგდევზთიკლმნოპ",
			expectedStatus:        http.StatusCreated,
			expectedUsernameError: "",
		},
		{
			name:                  "Invalid username: less than 3 chars",
			username:              "ab",
			expectedStatus:        http.StatusBadRequest,
			expectedUsernameError: "Username must be between 3 and 15 characters.",
		},
		{
			name:                  "Invalid username: more than 15 chars",
			username:              "abcdefghijklmnop",
			expectedStatus:        http.StatusBadRequest,
			expectedUsernameError: "Username must be between 3 and 15 characters.",
		},
		{
			name:                  "Invalid username: empty",
			username:              "",
			expectedStatus:        http.StatusBadRequest,
			expectedUsernameError: "Username cannot be empty.",
		},
		{
			name:                  "Invalid username: contains space",
			username:              "a bc",
			expectedStatus:        http.StatusBadRequest,
			expectedUsernameError: "Username cannot contain spaces.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newJoinEnv(t)

			body, _ := json.Marshal(chat.JoinRoomRequest{
				Username: tc.username,
			})
			headers := map[string]string{
				"Origin":                   "http://kseli.app",
				"Authorization":            env.invitetoken,
				"X-Participant-Session-Id": "user-session-id",
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

			if status != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d, body: %s", tc.name, tc.expectedStatus, status, string(respBody))
			}

			if tc.expectedStatus != http.StatusCreated {
				var errResp common.ErrorResponse
				if err := json.Unmarshal(respBody, &errResp); err != nil {
					t.Fatalf("[%s] failed to parse error JSON: %v", tc.name, err)
				}

				errMsg, ok := errResp.FieldErrors["username"]
				if !ok {
					t.Fatalf("[%s] expected field error for 'username', but none found", tc.name)
				}

				if errMsg != tc.expectedUsernameError {
					t.Fatalf("[%s] expected field error message %q, got %q", tc.name, tc.expectedUsernameError, errMsg)
				}
			}
		})
	}
}
