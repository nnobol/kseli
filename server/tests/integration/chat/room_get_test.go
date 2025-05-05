package chat_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"kseli/common"
	"kseli/config"
	"kseli/features/chat"
	"kseli/router"
)

type getEnv struct {
	roomID       string
	adminToken   string
	regularToken string
	mux          *http.ServeMux
}

func newGetEnv(t *testing.T) *getEnv {
	config.APIKey = "test-api-key"
	mux := router.New()

	// 1) Create the room to get the admin token
	createResp, _ := createRoom(t, true, 0, mux, 2, "http://kseli.app", config.APIKey, "admin")

	// 2) Fetch invite link via GetRoomHandler as an admin
	getReqHeaders := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": createResp.Token,
	}
	getStatus, getRespBody := sendRequest(mux, http.MethodGet, "/api/rooms/"+createResp.RoomID, nil, getReqHeaders)
	if getStatus != http.StatusOK {
		t.Fatalf("newGetEnv - expected get 200, got %d, body: %s", getStatus, string(getRespBody))
	}
	var gerRespStruct chat.GetRoomResponse
	if err := json.Unmarshal(getRespBody, &gerRespStruct); err != nil {
		t.Fatalf("newGetEnv - failed to unmarshal: %v", err)
	}

	parts := strings.Split(gerRespStruct.InviteLink, "?invite=")
	if len(parts) != 2 {
		t.Fatalf("newGetEnv - bad invite link %q", gerRespStruct.InviteLink)
	}

	// 3) Join the room to get the regular token
	joinReqBody, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	joinReqHeaders := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            parts[1],
		"X-Participant-Session-Id": "user-session-id",
	}
	joinStatus, joinRespBody := sendRequest(mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(joinReqBody), joinReqHeaders)
	if joinStatus != http.StatusCreated {
		t.Fatalf("newGetEnv - expected join 201, got %d, body: %s", joinStatus, string(joinRespBody))
	}
	var joinRespStruct chat.JoinRoomResponse
	if err := json.Unmarshal(joinRespBody, &joinRespStruct); err != nil {
		t.Fatalf("newGetEnv - failed to unmarshal: %v", err)
	}

	return &getEnv{
		roomID:       createResp.RoomID,
		adminToken:   createResp.Token,
		regularToken: joinRespStruct.Token,
		mux:          mux,
	}
}

func Test_GetRoom_Success_Admin(t *testing.T) {
	env := newGetEnv(t)

	headers := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": env.adminToken,
	}

	status, respBody := sendRequest(env.mux, http.MethodGet, "/api/rooms/"+env.roomID, nil, headers)

	expectedStatus := http.StatusOK
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var respStruct chat.GetRoomResponse
	if err := json.Unmarshal(respBody, &respStruct); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if respStruct.UserRole != common.Admin {
		t.Fatalf("expected UserRole %v, got %v", common.Admin, respStruct.UserRole)
	}

	if respStruct.InviteLink == "" {
		t.Fatal("expected non-empty InviteLink for admin user")
	}

	if !strings.Contains(respStruct.InviteLink, "?invite=") {
		t.Fatalf("unexpected InviteLink format: %s", respStruct.InviteLink)
	}
}

func Test_GetRoom_Success_User(t *testing.T) {
	env := newGetEnv(t)

	headers := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": env.regularToken,
	}

	status, respBody := sendRequest(env.mux, http.MethodGet, "/api/rooms/"+env.roomID, nil, headers)

	expectedStatus := http.StatusOK
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var respStruct chat.GetRoomResponse
	if err := json.Unmarshal(respBody, &respStruct); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if respStruct.UserRole != common.Member {
		t.Fatalf("expected UserRole %v, got %v", common.Member, respStruct.UserRole)
	}

	if respStruct.InviteLink != "" {
		t.Fatal("expected empty InviteLink for regular user")
	}
}

func Test_GetRoom_RoomNotFound(t *testing.T) {
	env := newGetEnv(t)

	headers := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": env.regularToken,
	}

	status, respBody := sendRequest(env.mux, http.MethodGet, "/api/rooms/non-existent-id", nil, headers)

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

func Test_GetRoom_AccessForbidden(t *testing.T) {
	env := newGetEnv(t)

	// 1) Create a new room (to get different claims and try to join using that)
	createResp, _ := createRoom(t, true, 0, env.mux, 2, "http://kseli.app", config.APIKey, "admin")

	// 2) Try to get the original room using the new token
	headers := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": createResp.Token,
	}

	status, respBody := sendRequest(env.mux, http.MethodGet, "/api/rooms/"+env.roomID, nil, headers)

	expectedStatus := http.StatusForbidden
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "You do not have access to this room."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_GetRoom_UserNotInRoom(t *testing.T) {
	env := newGetEnv(t)

	// 1) get kicked out of the room
	kickBody, _ := json.Marshal(chat.UserRequest{
		TargetUserID: 2,
	})
	kickHeaders := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": env.adminToken,
	}

	kickStatus, kickRespBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/kick", bytes.NewReader(kickBody), kickHeaders)

	kickExpectedStatus := http.StatusNoContent
	if kickStatus != kickExpectedStatus {
		t.Fatalf("Kick: expected %d, got %d, body: %s", kickExpectedStatus, kickStatus, string(kickRespBody))
	}

	// 2) try to get the room without rejoining
	headers := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": env.regularToken,
	}

	status, respBody := sendRequest(env.mux, http.MethodGet, "/api/rooms/"+env.roomID, nil, headers)

	expectedStatus := http.StatusForbidden
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "You are not in this room and can't retrieve the details. Try joining again."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_GetRoom_RoomIDValidation(t *testing.T) {
	env := newGetEnv(t)

	type testCase struct {
		name                string
		roomID              string
		expectedStatus      int
		expectedRoomIDError string
	}

	tests := []testCase{
		{
			name:                "Invalid roomID: spaces",
			roomID:              "a b",
			expectedStatus:      http.StatusBadRequest,
			expectedRoomIDError: "Chat Room Id cannot contain spaces.",
		},
		{
			name:                "Invalid roomID: too long",
			roomID:              "asdfghjklasdfghjk",
			expectedStatus:      http.StatusBadRequest,
			expectedRoomIDError: "Incorrect Chat Room Id, it is too long.",
		},
		{
			name:                "Valid roomID",
			roomID:              env.roomID,
			expectedStatus:      http.StatusOK,
			expectedRoomIDError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{
				"X-Origin":      "http://kseli.app",
				"Authorization": env.regularToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodGet, "/api/rooms/"+url.PathEscape(tc.roomID), nil, headers)

			if status != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d, body: %s", tc.name, tc.expectedStatus, status, string(respBody))
			}

			if tc.expectedStatus != http.StatusOK {
				var errResp common.ErrorResponse
				if err := json.Unmarshal(respBody, &errResp); err != nil {
					t.Fatalf("[%s] failed to unmarshal: %v", tc.name, err)
				}

				if errResp.Message != tc.expectedRoomIDError {
					t.Fatalf("[%s] expected error message %q, got %q", tc.name, tc.expectedRoomIDError, errResp.Message)
				}
			}
		})
	}
}
