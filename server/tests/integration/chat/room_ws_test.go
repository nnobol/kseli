package chat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kseli/config"
	"kseli/features/chat"
	"kseli/router"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type roomWSEnv struct {
	token       string
	inviteToken string
	roomID      string
	server      *httptest.Server
	serverAddr  string
	mux         *http.ServeMux
}

func newRoomWSEnv(t *testing.T) *roomWSEnv {
	config.APIKey = "test-api-key"
	mux := router.New()

	createReqBody, _ := json.Marshal(chat.CreateRoomRequest{
		Username:        "admin",
		MaxParticipants: 3,
	})
	createReqHeaders := map[string]string{
		"Origin":                   "http://kseli.app",
		"X-Api-Key":                config.APIKey,
		"X-Participant-Session-Id": "admin-session-id",
	}
	createStatus, createRespBody := sendRequest(mux, http.MethodPost, "/api/rooms", bytes.NewReader(createReqBody), createReqHeaders)
	if createStatus != http.StatusCreated {
		t.Fatalf("newRoomWSEnv - expected create 201, got %d, body: %s", createStatus, string(createRespBody))
	}
	var createRespStruct chat.CreateRoomResponse
	if err := json.Unmarshal(createRespBody, &createRespStruct); err != nil {
		t.Fatalf("newRoomWSEnv - failed to unmarshal: %v", err)
	}

	getReqHeaders := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": createRespStruct.Token,
	}
	getStatus, getRespBody := sendRequest(mux, http.MethodGet, "/api/rooms/"+createRespStruct.RoomID, nil, getReqHeaders)
	if getStatus != http.StatusOK {
		t.Fatalf("newKickBanEnv - expected get 200, got %d, body: %s", getStatus, string(getRespBody))
	}
	var gerRespStruct chat.GetRoomResponse
	if err := json.Unmarshal(getRespBody, &gerRespStruct); err != nil {
		t.Fatalf("newKickBanEnv - failed to unmarshal: %v", err)
	}

	parts := strings.Split(gerRespStruct.InviteLink, "?invite=")
	if len(parts) != 2 {
		t.Fatalf("newKickBanEnv - bad invite link %q", gerRespStruct.InviteLink)
	}

	server := httptest.NewServer(mux)
	serverAddr := server.Listener.Addr().String()

	return &roomWSEnv{
		token:       createRespStruct.Token,
		inviteToken: parts[1],
		roomID:      createRespStruct.RoomID,
		server:      server,
		serverAddr:  serverAddr,
		mux:         mux,
	}
}

func Test_RoomWS_Success(t *testing.T) {
	env := newRoomWSEnv(t)

	wsURL := "ws://" + env.serverAddr + "/ws/room?token=" + env.token

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn, _, _, err := dialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	msg, op, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("failed to read text frame: %v", err)
	}
	if op != ws.OpText {
		t.Fatalf("expected text frame, got op=%v", op)
	}

	var respMsg chat.WSMsg
	if err := json.Unmarshal(msg, &respMsg); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}

	if respMsg.MsgType != "join" {
		t.Errorf("expected type=join, got %q", respMsg.MsgType)
	}
}

func Test_RoomWS_InvalidOrigin(t *testing.T) {
	chat.WaitForTestClientWSSetup = make(chan struct{})
	env := newRoomWSEnv(t)

	wsURL := "ws://" + env.serverAddr + "/ws/room?token=" + env.token

	conn, _, _, err := ws.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	close(chat.WaitForTestClientWSSetup)

	frame, err := ws.ReadFrame(conn)
	if err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}

	if frame.Header.OpCode != ws.OpClose {
		t.Fatalf("Expected OpClose, got Op=%v", frame.Header.OpCode)
	}

	code, reason := ws.ParseCloseFrameData(frame.Payload)

	if code != ws.StatusNormalClosure {
		t.Errorf("Expected status 1000, got %d", code)
	}
	if reason != "invalid-origin" {
		t.Errorf("Expected reason `invalid-origin`, got %q", reason)
	}

	if _, _, err = wsutil.ReadServerData(conn); err == nil {
		t.Fatalf("expected read failure after close, got nil")
	}

	if _, ok := err.(wsutil.ClosedError); ok {
		// good
	} else if err == io.EOF {
		// good
	} else {
		t.Fatalf("expected closed/EOF error, got %v", err)
	}

}

func Test_RoomWS_TokenMissing(t *testing.T) {
	chat.WaitForTestClientWSSetup = make(chan struct{})
	env := newRoomWSEnv(t)

	wsURL := "ws://" + env.serverAddr + "/ws/room"

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn, _, _, err := dialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	close(chat.WaitForTestClientWSSetup)

	frame, err := ws.ReadFrame(conn)
	if err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}

	if frame.Header.OpCode != ws.OpClose {
		t.Fatalf("Expected OpClose, got Op=%v", frame.Header.OpCode)
	}

	code, reason := ws.ParseCloseFrameData(frame.Payload)

	if code != ws.StatusNormalClosure {
		t.Errorf("Expected status 1000, got %d", code)
	}
	if reason != "token-missing" {
		t.Errorf("Expected reason `token-missing`, got %q", reason)
	}

	if _, _, err = wsutil.ReadServerData(conn); err == nil {
		t.Fatalf("expected read failure after close, got nil")
	}

	if _, ok := err.(wsutil.ClosedError); ok {
		// good
	} else if err == io.EOF {
		// good
	} else {
		t.Fatalf("expected closed/EOF error, got %v", err)
	}
}

func Test_RoomWS_TokenInvalid(t *testing.T) {
	chat.WaitForTestClientWSSetup = make(chan struct{})
	env := newRoomWSEnv(t)

	wsURL := "ws://" + env.serverAddr + "/ws/room?token=invalid-token"

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn, _, _, err := dialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	close(chat.WaitForTestClientWSSetup)

	frame, err := ws.ReadFrame(conn)
	if err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}

	if frame.Header.OpCode != ws.OpClose {
		t.Fatalf("Expected OpClose, got Op=%v", frame.Header.OpCode)
	}

	code, reason := ws.ParseCloseFrameData(frame.Payload)

	if code != ws.StatusNormalClosure {
		t.Errorf("Expected status 1000, got %d", code)
	}
	if reason != "token-invalid" {
		t.Errorf("Expected reason `token-invalid`, got %q", reason)
	}

	if _, _, err = wsutil.ReadServerData(conn); err == nil {
		t.Fatalf("expected read failure after close, got nil")
	}

	if _, ok := err.(wsutil.ClosedError); ok {
		// good
	} else if err == io.EOF {
		// good
	} else {
		t.Fatalf("expected closed/EOF error, got %v", err)
	}
}

func Test_RoomWS_RoomNotExists(t *testing.T) {
	chat.WaitForTestClientWSSetup = make(chan struct{})
	env := newRoomWSEnv(t)

	// 1) close the room with delete request
	headers := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": env.token,
	}

	status, respBody := sendRequest(env.mux, http.MethodDelete, "/api/rooms/"+env.roomID, nil, headers)

	expectedStatus := http.StatusNoContent
	if status != expectedStatus {
		t.Fatalf("Delete: expected %d, got %d, body: %s", expectedStatus, status, string(respBody))
	}

	// 2) try connecting to the ws
	wsURL := "ws://" + env.serverAddr + "/ws/room?token=" + env.token

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn, _, _, err := dialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	close(chat.WaitForTestClientWSSetup)

	frame, err := ws.ReadFrame(conn)
	if err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}

	if frame.Header.OpCode != ws.OpClose {
		t.Fatalf("Expected OpClose, got Op=%v", frame.Header.OpCode)
	}

	code, reason := ws.ParseCloseFrameData(frame.Payload)

	if code != ws.StatusNormalClosure {
		t.Errorf("Expected status 1000, got %d", code)
	}
	if reason != "room-not-exists" {
		t.Errorf("Expected reason `room-not-exists`, got %q", reason)
	}

	if _, _, err = wsutil.ReadServerData(conn); err == nil {
		t.Fatalf("expected read failure after close, got nil")
	}

	if _, ok := err.(wsutil.ClosedError); ok {
		// good
	} else if err == io.EOF {
		// good
	} else {
		t.Fatalf("expected closed/EOF error, got %v", err)
	}
}

func Test_RoomWS_UserNotExists(t *testing.T) {
	chat.WaitForTestClientWSSetup = make(chan struct{})
	env := newRoomWSEnv(t)

	// 1) join the room as a new user
	joinReqBody, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	joinReqHeaders := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.inviteToken,
		"X-Participant-Session-Id": "user-session-id",
	}
	joinStatus, joinRespBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(joinReqBody), joinReqHeaders)
	if joinStatus != http.StatusCreated {
		t.Fatalf("Join: expected join 201, got %d, body: %s", joinStatus, string(joinRespBody))
	}

	var joinRespStruct chat.JoinRoomResponse
	if err := json.Unmarshal(joinRespBody, &joinRespStruct); err != nil {
		t.Fatalf("Join: failed to unmarshal: %v", err)
	}

	// 2) kick the user from the room
	kickBody, _ := json.Marshal(chat.UserRequest{
		TargetUserID: 2,
	})
	kickHeaders := map[string]string{
		"Origin":        "http://kseli.app",
		"Authorization": env.token,
	}

	kickStatus, kickRespBody := sendRequest(env.mux, http.MethodPost, "/api/rooms/"+env.roomID+"/kick", bytes.NewReader(kickBody), kickHeaders)

	kickExpectedStatus := http.StatusNoContent
	if kickStatus != kickExpectedStatus {
		t.Fatalf("Kick: expected %d, got %d, body: %s", kickExpectedStatus, kickStatus, string(kickRespBody))
	}

	// 2) try connecting to the ws
	wsURL := "ws://" + env.serverAddr + "/ws/room?token=" + joinRespStruct.Token

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn, _, _, err := dialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	close(chat.WaitForTestClientWSSetup)

	frame, err := ws.ReadFrame(conn)
	if err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}

	if frame.Header.OpCode != ws.OpClose {
		t.Fatalf("Expected OpClose, got Op=%v", frame.Header.OpCode)
	}

	code, reason := ws.ParseCloseFrameData(frame.Payload)

	if code != ws.StatusNormalClosure {
		t.Errorf("Expected status 1000, got %d", code)
	}
	if reason != "user-not-exists" {
		t.Errorf("Expected reason `user-not-exists`, got %q", reason)
	}

	if _, _, err = wsutil.ReadServerData(conn); err == nil {
		t.Fatalf("expected read failure after close, got nil")
	}

	if _, ok := err.(wsutil.ClosedError); ok {
		// good
	} else if err == io.EOF {
		// good
	} else {
		t.Fatalf("expected closed/EOF error, got %v", err)
	}
}

func Test_RoomWS_MessageTooLarge(t *testing.T) {
	chat.WaitForTestClientWSSetup = make(chan struct{})
	env := newRoomWSEnv(t)

	wsURL := "ws://" + env.serverAddr + "/ws/room?token=" + env.token

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn, _, _, err := dialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	close(chat.WaitForTestClientWSSetup)

	_, op, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("failed to read join frame: %v", err)
	}
	if op != ws.OpText {
		t.Fatalf("expected join message (OpText), got Op=%v", op)
	}

	oversizedMsg := make([]byte, 1025)
	for i := range oversizedMsg {
		oversizedMsg[i] = 'a'
	}

	if err := wsutil.WriteClientText(conn, oversizedMsg); err != nil {
		t.Fatalf("failed to write oversized message: %v", err)
	}

	frame, err := ws.ReadFrame(conn)
	if err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}

	if frame.Header.OpCode != ws.OpClose {
		t.Fatalf("Expected OpClose, got Op=%v", frame.Header.OpCode)
	}

	code, reason := ws.ParseCloseFrameData(frame.Payload)

	if code != ws.StatusNormalClosure {
		t.Errorf("Expected status 1000, got %d", code)
	}
	if reason != "message-too-large" {
		t.Errorf("Expected reason `message-too-large`, got %q", reason)
	}

	if _, _, err = wsutil.ReadServerData(conn); err == nil {
		t.Fatalf("expected read failure after close, got nil")
	}

	if _, ok := err.(wsutil.ClosedError); ok {
		// good
	} else if err == io.EOF {
		// good
	} else {
		t.Fatalf("expected closed/EOF error, got %v", err)
	}
}
