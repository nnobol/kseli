package chat_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"kseli/auth"
	"kseli/common"
	"kseli/config"
	"kseli/router"
)

func newKickBanEnv(t *testing.T) *getEnv {
	config.APIKey = "test-api-key"
	mux := router.New()

	// 1) Create the room to get the admin token
	createResp, _ := createRoom(t, true, 0, mux, 2, "admin", "http://kseli.app", config.APIKey, "admin")

	// 2) Fetch invite token via get room as an admin
	inviteToken, _, _ := getRoom(t, true, true, 0, mux, createResp.RoomID, "http://kseli.app", createResp.Token)

	// 3) Join the room to get the regular token
	joinResp, _ := joinRoom(t, true, 0, mux, "user", "http://kseli.app", inviteToken, "user")

	return &getEnv{
		roomID:       createResp.RoomID,
		adminToken:   createResp.Token,
		regularToken: joinResp.Token,
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

			kickOrBanUser(t, true, 0, env.mux, 2, action, env.roomID, "http://kseli.app", env.adminToken)
		})
	}
}

func Test_KickAndBan_OriginValidation(t *testing.T) {
	env := newKickBanEnv(t)

	actions := []string{
		"kick", "ban",
	}

	type testCase struct {
		name           string
		origin         string
		expectedStatus int
		expectedErrMsg string
	}

	for _, action := range actions {
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
			tcName := action + ": " + tc.name
			t.Run(tcName, func(t *testing.T) {
				errResp := kickOrBanUser(t, false, tc.expectedStatus, env.mux, 2, action, env.roomID, tc.origin, env.adminToken)

				if errResp.Message != tc.expectedErrMsg {
					t.Fatalf("[%s] expected error message %q, got %q", tcName, tc.expectedErrMsg, errResp.Message)
				}
			})
		}
	}
}

func Test_KickAndBan_TokenValidation(t *testing.T) {
	env := newKickBanEnv(t)

	actions := []string{
		"kick", "ban",
	}

	type testCase struct {
		name           string
		token          string
		expectedStatus int
		expectedErrMsg string
	}

	for _, action := range actions {
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
			tcName := action + ": " + tc.name
			t.Run(tcName, func(t *testing.T) {
				errResp := kickOrBanUser(t, false, tc.expectedStatus, env.mux, 2, action, env.roomID, "http://kseli.app", tc.token)

				if errResp.Message != tc.expectedErrMsg {
					t.Fatalf("[%s] expected error message %q, got %q", tcName, tc.expectedErrMsg, errResp.Message)
				}
			})
		}
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
				if tc.expectedStatus != http.StatusNoContent {
					errResp := kickOrBanUser(t, false, tc.expectedStatus, env.mux, 2, action, tc.roomID, "http://kseli.app", env.adminToken)

					if errResp.Message != tc.expectedRoomIDError {
						t.Fatalf("[%s] expected error message %q, got %q", tcName, tc.expectedRoomIDError, errResp.Message)
					}
				} else {
					kickOrBanUser(t, true, 0, env.mux, 2, action, tc.roomID, "http://kseli.app", env.adminToken)
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

			errResp := kickOrBanUser(t, false, 400, env.mux, 0, action, env.roomID, "http://kseli.app", env.adminToken)

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

			errResp := kickOrBanUser(t, false, 404, env.mux, 2, action, "invalid-room-id", "http://kseli.app", env.adminToken)

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
			createResp, _ := createRoom(t, true, 0, env.mux, 2, "admin", "http://kseli.app", config.APIKey, "admin")

			// 2) kick or ban the user using the different room id
			errResp := kickOrBanUser(t, false, 403, env.mux, 2, action, createResp.RoomID, "http://kseli.app", env.adminToken)

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

			errResp := kickOrBanUser(t, false, 403, env.mux, 2, action, env.roomID, "http://kseli.app", env.regularToken)

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

			errResp := kickOrBanUser(t, false, 400, env.mux, 1, action, env.roomID, "http://kseli.app", env.adminToken)

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
			errResp := kickOrBanUser(t, false, 404, env.mux, userID, action, env.roomID, "http://kseli.app", env.adminToken)

			expectedErrMsg := fmt.Sprintf("Participant with ID '%d' not found in room", userID)
			if errResp.Message != expectedErrMsg {
				t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
			}
		})
	}
}
