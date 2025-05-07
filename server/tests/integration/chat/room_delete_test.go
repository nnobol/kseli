package chat_test

import (
	"net/http"
	"testing"

	"kseli/config"
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
	createResp, _ := createRoom(t, true, 0, mux, 2, "admin", "http://kseli.app", config.APIKey, "admin")

	// 2) Fetch invite token via get room as an admin
	inviteToken, _, _ := getRoom(t, true, true, 0, mux, createResp.RoomID, "http://kseli.app", createResp.Token)

	// 3) Join the room to get the regular token
	joinResp, _ := joinRoom(t, true, 0, mux, "user", "http://kseli.app", inviteToken, "user")

	return &delEnv{
		roomID:       createResp.RoomID,
		adminToken:   createResp.Token,
		regularToken: joinResp.Token,
		mux:          mux,
	}
}

func Test_DeleteRoom_Success(t *testing.T) {
	env := newDelEnv(t)

	deleteRoom(t, true, 0, env.mux, env.roomID, "http://kseli.app", env.adminToken)
}

func Test_DeleteRoom_RoomNotFound(t *testing.T) {
	env := newDelEnv(t)

	errResp := deleteRoom(t, false, 404, env.mux, "invalid-room-id", "http://kseli.app", env.adminToken)

	expectedErrMsg := "Chat Room not found."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_DeleteRoom_AccessForbidden(t *testing.T) {
	env := newDelEnv(t)

	// 1) Create a new room (to get different claims and try to delete using that)
	createResp, _ := createRoom(t, true, 0, env.mux, 2, "admin", "http://kseli.app", config.APIKey, "admin")

	// 2) Try to delete the original room using the new token
	errResp := deleteRoom(t, false, 403, env.mux, env.roomID, "http://kseli.app", createResp.Token)

	expectedErrMsg := "You do not have access to this room."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_DeleteRoom_NotAdmin(t *testing.T) {
	env := newDelEnv(t)

	errResp := deleteRoom(t, false, 403, env.mux, env.roomID, "http://kseli.app", env.regularToken)

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
			if tc.expectedStatus != http.StatusNoContent {
				errResp := deleteRoom(t, false, tc.expectedStatus, env.mux, tc.roomID, "http://kseli.app", env.adminToken)

				if errResp.Message != tc.expectedRoomIDError {
					t.Fatalf("[%s] expected error message %q, got %q", tc.name, tc.expectedRoomIDError, errResp.Message)
				}
			} else {
				deleteRoom(t, true, 0, env.mux, tc.roomID, "http://kseli.app", env.adminToken)
			}
		})
	}
}
