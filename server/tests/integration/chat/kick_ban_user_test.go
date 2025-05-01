package chat_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"kseli/common"
	"kseli/config"
	"kseli/features/chat"
	"kseli/router"
)

func newKickBanEnv(t *testing.T) *getEnv {
	config.APIKey = "test-api-key"
	mux := router.New()

	// 1) Create the room to get the admin token
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
		t.Fatalf("newGetEnv - expected create 201, got %d, body: %s", createStatus, string(createRespBody))
	}
	var createRespStruct chat.CreateRoomResponse
	if err := json.Unmarshal(createRespBody, &createRespStruct); err != nil {
		t.Fatalf("newGetEnv - failed to unmarshal: %v", err)
	}

	// 2) Fetch invite link via GetRoomHandler as an admin
	getReqHeaders := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": createRespStruct.Token,
	}
	getStatus, getRespBody := sendRequest(mux, http.MethodGet, "/api/rooms/"+createRespStruct.RoomID, nil, getReqHeaders)
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
		roomID:       createRespStruct.RoomID,
		adminToken:   createRespStruct.Token,
		regularToken: joinRespStruct.Token,
		mux:          mux,
	}
}

func Test_KickAndBan_Success(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	for _, action := range actions {
		tcName := action + ": success"
		t.Run(tcName, func(t *testing.T) {
			env := newKickBanEnv(t)

			body, _ := json.Marshal(chat.UserRequest{
				TargetUserID: 2,
			})

			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.adminToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/"+action, bytes.NewReader(body), headers)

			if status != http.StatusNoContent {
				t.Fatalf("[%s] expected %d, got %d, body: %s", tcName, http.StatusNoContent, status, string(respBody))
			}
		})
	}
}

func Test_KickAndBan_BadRequests(t *testing.T) {
	const maxBody = 128

	actions := []string{
		"kick", "ban",
	}

	type testCase struct {
		name           string
		body           []byte
		expectedStatus int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "bad request - malformed JSON",
			body:           []byte("{ not json"),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "bad request - empty body",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "bad request - overflow uint8",
			body:           []byte(`{"userId":555555}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "bad request - negative number",
			body:           []byte(`{"userId":-1}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "bad request - float in uint8",
			body:           []byte(`{"userId":2.5}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name: "bad request - body too large",
			body: func() []byte {
				largeUserID := strings.Repeat("1", maxBody)
				payload := fmt.Sprintf(`{"userId":%s}`, largeUserID)
				return []byte(payload)
			}(),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "valid request - extra fields allowed",
			body:           []byte(`{"userId":2,"extra":"ignored"}`),
			expectedStatus: http.StatusNoContent,
			expectedErrMsg: "",
		},
	}

	for _, action := range actions {
		for _, tc := range tests {
			tcName := action + ": " + tc.name
			t.Run(tcName, func(t *testing.T) {
				env := newKickBanEnv(t)

				headers := map[string]string{
					"Origin":        "http://kseli.app",
					"Authorization": env.adminToken,
				}

				status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/"+action, bytes.NewReader(tc.body), headers)

				if status != tc.expectedStatus {
					t.Fatalf("[%s] expected %d, got %d, body: %s", tcName, tc.expectedStatus, status, string(respBody))
				}

				if tc.expectedStatus != http.StatusNoContent {
					var errResp common.ErrorResponse
					if err := json.Unmarshal(respBody, &errResp); err != nil {
						t.Fatalf("[%s] failed to unmarshal: %v", tcName, err)
					}

					errMsg := errResp.Message
					if errMsg != tc.expectedErrMsg {
						t.Fatalf("[%s] expected error message %q, got %q", tcName, tc.expectedErrMsg, errMsg)
					}
				}
			})
		}
	}
}

func Test_KickAndBan_RoomIDValidation(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	type testCase struct {
		name                string
		roomID              string
		expectedStatus      int
		expectedRoomIDError string
	}

	for _, action := range actions {
		env := newKickBanEnv(t)

		tests := []testCase{
			{
				name:                "invalid roomID: spaces",
				roomID:              "a b",
				expectedStatus:      http.StatusBadRequest,
				expectedRoomIDError: "Chat Room Id cannot contain spaces.",
			},
			{
				name:                "invalid roomID: too long",
				roomID:              "asdfghjklasdfghjk",
				expectedStatus:      http.StatusBadRequest,
				expectedRoomIDError: "Incorrect Chat Room Id, it is too long.",
			},
			{
				name:                "valid roomID",
				roomID:              env.roomID,
				expectedStatus:      http.StatusNoContent,
				expectedRoomIDError: "",
			},
		}

		for _, tc := range tests {
			tcName := action + ": " + tc.name
			t.Run(tcName, func(t *testing.T) {
				body, _ := json.Marshal(chat.UserRequest{
					TargetUserID: 2,
				})

				headers := map[string]string{
					"Origin":        "http://kseli.app",
					"Authorization": env.adminToken,
				}

				status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+url.PathEscape(tc.roomID)+"/"+action, bytes.NewReader(body), headers)

				if status != tc.expectedStatus {
					t.Fatalf("[%s] expected %d, got %d, body: %s", tcName, tc.expectedStatus, status, string(respBody))
				}

				if tc.expectedStatus != http.StatusNoContent {
					var errResp common.ErrorResponse
					if err := json.Unmarshal(respBody, &errResp); err != nil {
						t.Fatalf("[%s] failed to unmarshal: %v", tcName, err)
					}

					if errResp.Message != tc.expectedRoomIDError {
						t.Fatalf("[%s] expected error message %q, got %q", tcName, tc.expectedRoomIDError, errResp.Message)
					}
				}
			})
		}
	}
}

func Test_KickAndBan_UserIDEmpty(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	for _, action := range actions {
		tcName := action + ": userId empty"
		t.Run(tcName, func(t *testing.T) {
			env := newKickBanEnv(t)

			body, _ := json.Marshal(chat.UserRequest{
				TargetUserID: 0,
			})

			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.adminToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/"+action, bytes.NewReader(body), headers)

			expectedStatus := http.StatusBadRequest
			if status != expectedStatus {
				t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
			}

			var errResp common.ErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			expectedErrMsg := "User Id is required in the request."
			if errResp.Message != expectedErrMsg {
				t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
			}
		})
	}
}

func Test_KickAndBan_RoomNotFound(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	for _, action := range actions {
		tcName := action + ": userId empty"
		t.Run(tcName, func(t *testing.T) {
			env := newKickBanEnv(t)

			body, _ := json.Marshal(chat.UserRequest{
				TargetUserID: 2,
			})

			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.adminToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/non-existent-id/"+action, bytes.NewReader(body), headers)

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
		})
	}
}

func Test_KickAndBan_AccessForbidden(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	for _, action := range actions {
		tcName := action + ": userId empty"
		t.Run(tcName, func(t *testing.T) {
			env := newKickBanEnv(t)

			// 1) Create a new room (to get different claims and try to kick and ban using that)
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

			// 2) kick or ban the user using the different room id
			body, _ := json.Marshal(chat.UserRequest{
				TargetUserID: 2,
			})

			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.adminToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+createRespStruct.RoomID+"/"+action, bytes.NewReader(body), headers)

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
		})
	}
}

func Test_KickAndBan_NotAdmin(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	for _, action := range actions {
		tcName := action + ": userId empty"
		t.Run(tcName, func(t *testing.T) {
			env := newKickBanEnv(t)

			body, _ := json.Marshal(chat.UserRequest{
				TargetUserID: 2,
			})

			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.regularToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/"+action, bytes.NewReader(body), headers)

			expectedStatus := http.StatusForbidden
			if status != expectedStatus {
				t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
			}

			var errResp common.ErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			expectedErrMsg := fmt.Sprintf("You are not an admin and can't %s anyone from this room.", action)
			if errResp.Message != expectedErrMsg {
				t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
			}
		})
	}
}

func Test_KickAndBan_ActionOnAdmin(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	for _, action := range actions {
		tcName := action + ": userId empty"
		t.Run(tcName, func(t *testing.T) {
			env := newKickBanEnv(t)

			body, _ := json.Marshal(chat.UserRequest{
				TargetUserID: 1,
			})

			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.adminToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/"+action, bytes.NewReader(body), headers)

			expectedStatus := http.StatusForbidden
			if status != expectedStatus {
				t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
			}

			var errResp common.ErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			expectedErrMsg := fmt.Sprintf("You can't %s yourself from the room.", action)
			if errResp.Message != expectedErrMsg {
				t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
			}
		})
	}
}

func Test_KickAndBan_UserNotFound(t *testing.T) {
	actions := []string{
		"kick", "ban",
	}

	for _, action := range actions {
		tcName := action + ": userId empty"
		t.Run(tcName, func(t *testing.T) {
			env := newKickBanEnv(t)

			var userID uint8 = 3

			body, _ := json.Marshal(chat.UserRequest{
				TargetUserID: userID,
			})

			headers := map[string]string{
				"Origin":        "http://kseli.app",
				"Authorization": env.adminToken,
			}

			status, respBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/"+action, bytes.NewReader(body), headers)

			expectedStatus := http.StatusNotFound
			if status != expectedStatus {
				t.Fatalf("expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
			}

			var errResp common.ErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			expectedErrMsg := fmt.Sprintf("Participant with ID '%d' not found in room", userID)
			if errResp.Message != expectedErrMsg {
				t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
			}
		})
	}
}
