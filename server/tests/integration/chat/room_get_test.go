package chat_test

import (
	"bytes"
	"encoding/json"
	"net/http"
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

func Test_GetRoom_Success_Admin(t *testing.T) {
	env := newGetEnv(t)

	_, resp, _ := getRoom(t, true, true, 0, env.mux, env.roomID, "http://kseli.app", env.adminToken)

	if resp.UserRole != common.Admin {
		t.Fatalf("expected UserRole %v, got %v", common.Admin, resp.UserRole)
	}

	if resp.InviteLink == "" {
		t.Fatal("expected non-empty InviteLink for admin user")
	}

	if !strings.Contains(resp.InviteLink, "?invite=") {
		t.Fatalf("unexpected InviteLink format: %s", resp.InviteLink)
	}
}

func Test_GetRoom_Success_User(t *testing.T) {
	env := newGetEnv(t)

	_, resp, _ := getRoom(t, true, false, 0, env.mux, env.roomID, "http://kseli.app", env.regularToken)

	if resp.UserRole != common.Member {
		t.Fatalf("expected UserRole %v, got %v", common.Member, resp.UserRole)
	}

	if resp.InviteLink != "" {
		t.Fatal("expected empty InviteLink for regular user")
	}
}

func Test_GetRoom_RoomNotFound(t *testing.T) {
	env := newGetEnv(t)

	_, _, errResp := getRoom(t, false, false, 404, env.mux, "invalid-room-id", "http://kseli.app", env.regularToken)

	expectedErrMsg := "Chat Room not found."
	if errResp.Message != expectedErrMsg {
		t.Fatalf("expected error message %q, got %q", expectedErrMsg, errResp.Message)
	}
}

func Test_GetRoom_AccessForbidden(t *testing.T) {
	env := newGetEnv(t)

	// 1) Create a new room (to get different claims and try to join using that)
	newRoom, _ := createRoom(t, true, 0, env.mux, 2, "admin", "http://kseli.app", config.APIKey, "admin")

	// 2) Try to get the original room using the new token
	_, _, errResp := getRoom(t, false, false, 403, env.mux, env.roomID, "http://kseli.app", newRoom.Token)

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
	_, _, errResp := getRoom(t, false, false, 403, env.mux, env.roomID, "http://kseli.app", env.regularToken)

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
			if tc.expectedStatus != http.StatusOK {
				_, _, errResp := getRoom(t, false, false, tc.expectedStatus, env.mux, tc.roomID, "http://kseli.app", env.regularToken)

				if errResp.Message != tc.expectedRoomIDError {
					t.Fatalf("[%s] expected error message %q, got %q", tc.name, tc.expectedRoomIDError, errResp.Message)
				}
			} else {
				getRoom(t, true, false, 0, env.mux, tc.roomID, "http://kseli.app", env.regularToken)
			}
		})
	}
}
