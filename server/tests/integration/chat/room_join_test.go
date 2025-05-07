package chat_test

import (
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
	createResp, _ := createRoom(t, true, 0, mux, 2, "admin", "http://kseli.app", config.APIKey, "admin")

	// 2) Fetch invite token via get room as an admin
	inviteToken, _, _ := getRoom(t, true, true, 0, mux, createResp.RoomID, "http://kseli.app", createResp.Token)

	return &joinEnv{
		roomID:      createResp.RoomID,
		invitetoken: inviteToken,
		adminToken:  createResp.Token,
		mux:         mux,
	}
}

func Test_JoinRoom_Success(t *testing.T) {
	env := newJoinEnv(t)

	resp, _ := joinRoom(t, true, 0, env.mux, "user", "http://kseli.app", env.invitetoken, "user")

	if resp.RoomID == "" || resp.Token == "" {
		t.Fatalf("got empty RoomID or Token: %+v", resp)
	}
}

func Test_JoinRoom_OriginValidation(t *testing.T) {
	env := newJoinEnv(t)

	type testCase struct {
		name           string
		origin         string
		expectedStatus int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "Origin missing",
			origin:         "",
			expectedStatus: http.StatusForbidden,
			expectedErrMsg: "Missing Origin header.",
		},
		{
			name:           "Invalid origin",
			origin:         "invalid-origin",
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "Invalid Origin header.",
		},
		{
			name:           "Origin not allowed",
			origin:         "http://kseli.apps",
			expectedStatus: http.StatusForbidden,
			expectedErrMsg: "Origin not allowed. Access from this origin is restricted.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errResp := joinRoom(t, false, tc.expectedStatus, env.mux, "user", tc.origin, env.invitetoken, "user")

			if errResp.Message != tc.expectedErrMsg {
				t.Fatalf("[%s] expected error message %q, got %q", tc.name, tc.expectedErrMsg, errResp.Message)
			}
		})
	}
}

func Test_JoinRoom_TokenValidation(t *testing.T) {
	env := newJoinEnv(t)

	type testCase struct {
		name           string
		token          string
		expectedStatus int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "Token missing",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedErrMsg: "Missing Authorization token.",
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedErrMsg: "Invalid or expired token.",
		},
		{
			name: "Expired token",
			token: func() string {
				expiredClaims := auth.Claims{
					UserID:   1,
					Username: "does-not-matter",
					Role:     common.Admin,
					RoomID:   "does-not-matter",
					Exp:      time.Now().Add(-1 * time.Minute).Unix(),
				}
				token, _ := auth.CreateToken(expiredClaims)
				return token
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedErrMsg: "Invalid or expired token.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errResp := joinRoom(t, false, tc.expectedStatus, env.mux, "user", "http://kseli.app", tc.token, "user")

			if errResp.Message != tc.expectedErrMsg {
				t.Fatalf("[%s] expected error message %q, got %q", tc.name, tc.expectedErrMsg, errResp.Message)
			}
		})
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

	_, errResp := joinRoom(t, false, 404, env.mux, "user", "http://kseli.app", fakeToken, "user")

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

	_, errResp := joinRoom(t, false, 403, env.mux, "user", "http://kseli.app", fakeToken, "user")

	expectedErrMsg := "Invalid invite link."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_AlreadyInRoom(t *testing.T) {
	env := newJoinEnv(t)

	_, errResp := joinRoom(t, false, 400, env.mux, "user", "http://kseli.app", env.invitetoken, "admin")

	expectedErrMsg := "You can not join a room you are already in."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_BannedFromRoom(t *testing.T) {
	env := newJoinEnv(t)

	// 1) Join the room
	joinRoom(t, true, 0, env.mux, "user", "http://kseli.app", env.invitetoken, "user")

	// 2) Get banned from the room
	kickOrBanUser(t, true, 0, env.mux, 2, "ban", env.roomID, "http://kseli.app", env.adminToken)

	// 3) Try to join the room again
	_, errResp := joinRoom(t, false, 403, env.mux, "user", "http://kseli.app", env.invitetoken, "user")

	expectedErrMsg := "You are banned from this room."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_RoomFull(t *testing.T) {
	env := newJoinEnv(t)

	// 1) Join to fill uo the room fully
	joinRoom(t, true, 0, env.mux, "user", "http://kseli.app", env.invitetoken, "user")

	// 2) Try to join with another user
	_, errResp := joinRoom(t, false, 409, env.mux, "user2", "http://kseli.app", env.invitetoken, "user2")

	expectedErrMsg := "Chat Room is full."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_JoinRoom_UsernameTaken(t *testing.T) {
	env := newJoinEnv(t)

	_, errResp := joinRoom(t, false, 400, env.mux, "admin", "http://kseli.app", env.invitetoken, "user")

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

			if tc.expectedStatus != http.StatusCreated {
				_, errResp := joinRoom(t, false, tc.expectedStatus, env.mux, tc.username, "http://kseli.app", env.invitetoken, "user")

				errMsg, ok := errResp.FieldErrors["username"]
				if !ok {
					t.Fatalf("[%s] expected field error for 'username', but none found", tc.name)
				}

				if errMsg != tc.expectedUsernameError {
					t.Fatalf("[%s] expected field error message %q, got %q", tc.name, tc.expectedUsernameError, errMsg)
				}
			} else {
				joinRoom(t, true, 0, env.mux, tc.username, "http://kseli.app", env.invitetoken, "user")
			}
		})
	}
}
