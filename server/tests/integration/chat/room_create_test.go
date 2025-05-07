package chat_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"kseli/common"
	"kseli/config"
	"kseli/router"
)

func newCreateEnv() *http.ServeMux {
	config.APIKey = "test-api-key"
	mux := router.New()

	return mux
}

func Test_CreateRoom_Success(t *testing.T) {
	mux := newCreateEnv()

	resp, _ := createRoom(t, true, 0, mux, 3, "admin", "http://kseli.app", config.APIKey, "admin")

	if resp.RoomID == "" || resp.Token == "" {
		t.Fatalf("got empty RoomID or Token: %+v", resp)
	}
}

func Test_CreateRoom_BadRequests(t *testing.T) {
	mux := newCreateEnv()

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
			body:           strings.NewReader(`{"username":123,"maxParticipants":"foo"}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "Bad request: negative number",
			body:           strings.NewReader(`{"username":"abc","maxParticipants":-1}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "Bad request: float in uint8",
			body:           strings.NewReader(`{"username":"abc","maxParticipants":2.5}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "Bad request: overflow uint8",
			body:           strings.NewReader(`{"username":"abc","maxParticipants":300}`),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name: "Bad request: body too large",
			body: func() io.Reader {
				bigUsername := strings.Repeat("a", maxBody)
				payload := fmt.Sprintf(`{"username":%q,"maxParticipants":3}`, bigUsername)
				return strings.NewReader(payload)
			}(),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid JSON request body.",
		},
		{
			name:           "Valid request: extra fields allowed",
			body:           strings.NewReader(`{"username":"user","maxParticipants":3,"extra":"ignored"}`),
			expectedStatus: http.StatusCreated,
			expectedErrMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{
				"Origin":                   "http://kseli.app",
				"X-Api-Key":                config.APIKey,
				"X-Participant-Session-Id": "admin-session-id",
			}
			status, respBody := sendRequest(mux, http.MethodPost, "/api/rooms", tc.body, headers)

			if status != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d, body: %s", tc.name, tc.expectedStatus, status, string(respBody))
			}

			if tc.expectedStatus != http.StatusCreated {
				var errResp common.ErrorResponse
				if err := json.Unmarshal(respBody, &errResp); err != nil {
					t.Fatalf("[%s] failed to unmarshal: %v", tc.name, err)
				}

				errMsg := errResp.Message
				if errMsg != tc.expectedErrMsg {
					t.Fatalf("[%s] expected error message %q, got %q", tc.name, tc.expectedErrMsg, errMsg)
				}
			}
		})
	}
}

func Test_CreateRoom_UsernameValidation(t *testing.T) {
	mux := newCreateEnv()

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
			if tc.expectedStatus != http.StatusCreated {
				_, errResp := createRoom(t, false, tc.expectedStatus, mux, 3, tc.username, "http://kseli.app", config.APIKey, "admin")

				errMsg, ok := errResp.FieldErrors["username"]
				if !ok {
					t.Fatalf("[%s] expected field error for 'username', but none found", tc.name)
				}

				if errMsg != tc.expectedUsernameError {
					t.Fatalf("[%s] expected field error message %q, got %q", tc.name, tc.expectedUsernameError, errMsg)
				}
			} else {
				createRoom(t, true, 0, mux, 3, tc.username, "http://kseli.app", config.APIKey, "admin")
			}
		})
	}
}

func Test_CreateRoom_MaxParticipantsValidation(t *testing.T) {
	mux := newCreateEnv()

	type testCase struct {
		name                         string
		maxParticipants              uint8
		expectedStatus               int
		expectedMaxParticipantsError string
	}

	tests := []testCase{
		{
			name:                         "Valid maxParticipants: equals 2",
			maxParticipants:              2,
			expectedStatus:               http.StatusCreated,
			expectedMaxParticipantsError: "",
		},
		{
			name:                         "Valid maxParticipants: equals 5",
			maxParticipants:              5,
			expectedStatus:               http.StatusCreated,
			expectedMaxParticipantsError: "",
		},
		{
			name:                         "Invalid maxParticipants: less than 2",
			maxParticipants:              1,
			expectedStatus:               http.StatusBadRequest,
			expectedMaxParticipantsError: "Max participants must be between 2 and 5.",
		},
		{
			name:                         "Invalid maxParticipants: more than 5",
			maxParticipants:              6,
			expectedStatus:               http.StatusBadRequest,
			expectedMaxParticipantsError: "Max participants must be between 2 and 5.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus != http.StatusCreated {
				_, errResp := createRoom(t, false, tc.expectedStatus, mux, tc.maxParticipants, "admin", "http://kseli.app", config.APIKey, "admin")

				errMsg, ok := errResp.FieldErrors["maxParticipants"]
				if !ok {
					t.Fatalf("[%s] expected field error for 'username', but none found", tc.name)
				}

				if errMsg != tc.expectedMaxParticipantsError {
					t.Fatalf("[%s] expected field error message %q, got %q", tc.name, tc.expectedMaxParticipantsError, errMsg)
				}
			} else {
				createRoom(t, true, 0, mux, tc.maxParticipants, "admin", "http://kseli.app", config.APIKey, "admin")
			}
		})
	}
}
