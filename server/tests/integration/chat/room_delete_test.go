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

type delEnv struct {
	roomID       string
	adminToken   string
	regularToken string
	mux          *http.ServeMux
}

func newDelEnv(t *testing.T) *delEnv {
	config.APIKey = "test-api-key"
	mux := router.New()

	// 1) Create the room to get the admin
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
		t.Fatalf("newDelEnv - expected create 201, got %d, body: %s", createStatus, string(createRespBody))
	}
	var createRespStruct chat.CreateRoomResponse
	if err := json.Unmarshal(createRespBody, &createRespStruct); err != nil {
		t.Fatalf("newDelEnv - failed to unmarshal: %v", err)
	}

	// 2) Fetch invite link via GetRoomHandler as an admin
	getReqHeaders := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": createRespStruct.Token,
	}
	getStatus, getRespBody := sendRequest(mux, http.MethodGet, "/api/rooms/"+createRespStruct.RoomID, nil, getReqHeaders)
	if getStatus != http.StatusOK {
		t.Fatalf("newDelEnv - expected get 200, got %d, body: %s", getStatus, string(getRespBody))
	}
	var gerRespStruct chat.GetRoomResponse
	if err := json.Unmarshal(getRespBody, &gerRespStruct); err != nil {
		t.Fatalf("newDelEnv - failed to unmarshal: %v", err)
	}

	parts := strings.Split(gerRespStruct.InviteLink, "?invite=")
	if len(parts) != 2 {
		t.Fatalf("newDelEnv - bad invite link %q", gerRespStruct.InviteLink)
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
		t.Fatalf("newDelEnv - expected join 201, got %d, body: %s", joinStatus, string(joinRespBody))
	}
	var joinRespStruct chat.JoinRoomResponse
	if err := json.Unmarshal(joinRespBody, &joinRespStruct); err != nil {
		t.Fatalf("newDelEnv - failed to unmarshal: %v", err)
	}

	return &delEnv{
		roomID:       createRespStruct.RoomID,
		adminToken:   createRespStruct.Token,
		regularToken: joinRespStruct.Token,
		mux:          mux,
	}
}

func Test_DeleteRoom_Success(t *testing.T) {
	env := newDelEnv(t)

	headers := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": env.adminToken,
	}

	status, respBody := sendRequest(env.mux, http.MethodDelete, "/api/rooms/"+env.roomID, nil, headers)

	expectedStatus := http.StatusNoContent
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}
}

func Test_DeleteRoom_RoomNotFound(t *testing.T) {
	env := newDelEnv(t)

	headers := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": env.adminToken,
	}

	status, respBody := sendRequest(env.mux, http.MethodDelete, "/api/rooms/non-existent-id", nil, headers)

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

func Test_DeleteRoom_AccessForbidden(t *testing.T) {
	env := newDelEnv(t)

	// 1) Create a new room (to get different claims and try to delete using that)
	createReqBody, _ := json.Marshal(chat.CreateRoomRequest{
		Username:        "admin",
		MaxParticipants: 2,
	})

	createReqHeaders := map[string]string{
		"Origin":                   "http://kseli.app",
		"X-Api-Key":                config.APIKey,
		"X-Participant-Session-Id": "admin-session-id",
	}

	createStatus, createRespBody := sendRequest(env.mux, http.MethodPost, "/api/rooms", bytes.NewReader(createReqBody), createReqHeaders)

	if createStatus != http.StatusCreated {
		t.Fatalf("newGetEnv - expected create 201, got %d, body: %s", createStatus, string(createRespBody))
	}

	var createRespStruct chat.CreateRoomResponse
	if err := json.Unmarshal(createRespBody, &createRespStruct); err != nil {
		t.Fatalf("newGetEnv - failed to unmarshal: %v", err)
	}

	// 2) Try to delete the original room using the new token
	headers := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": createRespStruct.Token,
	}

	status, respBody := sendRequest(env.mux, http.MethodDelete, "/api/rooms/"+env.roomID, nil, headers)

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

func Test_DeleteRoom_NotAdmin(t *testing.T) {
	env := newDelEnv(t)

	headers := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": env.regularToken,
	}

	status, respBody := sendRequest(env.mux, http.MethodDelete, "/api/rooms/"+env.roomID, nil, headers)

	expectedStatus := http.StatusForbidden
	if status != expectedStatus {
		t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	var errResp common.ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedErrMsg := "You are not an admin and can't close this room."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_DeleteRoom_RoomIDValidation(t *testing.T) {
	env := newDelEnv(t)

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
			expectedStatus:      http.StatusNoContent,
			expectedRoomIDError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.adminToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodDelete, "/api/rooms/"+url.PathEscape(tc.roomID), nil, headers)

			if status != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d, body: %s", tc.name, tc.expectedStatus, status, string(respBody))
			}

			if tc.expectedStatus != http.StatusNoContent {
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
