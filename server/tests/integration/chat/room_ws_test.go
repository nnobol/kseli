package chat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kseli/common"
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

	// 1) Create the room to get the admin token
	createResp, _ := createRoom(t, true, 0, mux, 3, "http://kseli.app", config.APIKey, "admin")

	getReqHeaders := map[string]string{
		"X-Origin":      "http://kseli.app",
		"Authorization": createResp.Token,
	}
	getStatus, getRespBody := sendRequest(mux, http.MethodGet, "/api/rooms/"+createResp.RoomID, nil, getReqHeaders)
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
		token:       createResp.Token,
		inviteToken: parts[1],
		roomID:      createResp.RoomID,
		server:      server,
		serverAddr:  serverAddr,
		mux:         mux,
	}
}

func Test_RoomWS_Success_JoinMsgReceived(t *testing.T) {
	env := newRoomWSEnv(t)

	// join request to add a new user to the room
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

	wsURL1 := "ws://" + env.serverAddr + "/ws/room?token=" + env.token
	wsURL2 := "ws://" + env.serverAddr + "/ws/room?token=" + joinRespStruct.Token

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn1, _, _, err := dialer.Dial(context.Background(), wsURL1)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	conn2, _, _, err := dialer.Dial(context.Background(), wsURL2)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	// conn1: receive admin (self) join
	msg := mustReadWSJoin(t, conn1)
	assertJoinMsg(t, msg, 1, "admin", common.Admin)

	// conn1: receive user join
	msg = mustReadWSJoin(t, conn1)
	assertJoinMsg(t, msg, 2, "user", common.Member)

	// conn2: receive user join
	msg = mustReadWSJoin(t, conn2)
	assertJoinMsg(t, msg, 2, "user", common.Member)
}

func Test_RoomWS_Success_ChatMsgReceived(t *testing.T) {
	env := newRoomWSEnv(t)

	// join request to add a new user to the room
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

	wsURL1 := "ws://" + env.serverAddr + "/ws/room?token=" + env.token
	wsURL2 := "ws://" + env.serverAddr + "/ws/room?token=" + joinRespStruct.Token

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn1, _, _, err := dialer.Dial(context.Background(), wsURL1)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	conn2, _, _, err := dialer.Dial(context.Background(), wsURL2)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	// drain join messages
	mustReadWSJoin(t, conn1)
	mustReadWSJoin(t, conn1)
	mustReadWSJoin(t, conn2)

	adminMsg := "Hello from admin"
	if err := wsutil.WriteClientText(conn1, []byte(adminMsg)); err != nil {
		t.Fatalf("admin failed to send message: %v", err)
	}

	got1 := mustReadWSChat(t, conn1)
	got2 := mustReadWSChat(t, conn2)

	assertChatMsg(t, got1, "admin", adminMsg)
	assertChatMsg(t, got2, "admin", adminMsg)

	userMsg := "Hey admin"
	if err := wsutil.WriteClientText(conn2, []byte(userMsg)); err != nil {
		t.Fatalf("user failed to send message: %v", err)
	}

	got1 = mustReadWSChat(t, conn1)
	got2 = mustReadWSChat(t, conn2)

	assertChatMsg(t, got1, "user", userMsg)
	assertChatMsg(t, got2, "user", userMsg)
}

func Test_RoomWS_Success_LeaveMsgReceived(t *testing.T) {
	env := newRoomWSEnv(t)

	// 2 join requests to add a new users to the room
	joinReqHeaders1 := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.inviteToken,
		"X-Participant-Session-Id": "user-session-id",
	}
	joinReqHeaders2 := map[string]string{
		"Origin":                   "http://kseli.app",
		"Authorization":            env.inviteToken,
		"X-Participant-Session-Id": "user2-session-id",
	}
	joinReqBody1, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user",
	})
	joinReqBody2, _ := json.Marshal(chat.JoinRoomRequest{
		Username: "user2",
	})
	joinStatus1, joinRespBody1 := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(joinReqBody1), joinReqHeaders1)
	if joinStatus1 != http.StatusCreated {
		t.Fatalf("Join: expected join 201, got %d, body: %s", joinStatus1, string(joinRespBody1))
	}
	joinStatus2, joinRespBody2 := sendRequest(env.mux, http.MethodPost, "/api/rooms/join", bytes.NewReader(joinReqBody2), joinReqHeaders2)
	if joinStatus2 != http.StatusCreated {
		t.Fatalf("Join: expected join 201, got %d, body: %s", joinStatus1, string(joinRespBody2))
	}
	var joinRespStruct1 chat.JoinRoomResponse
	if err := json.Unmarshal(joinRespBody1, &joinRespStruct1); err != nil {
		t.Fatalf("Join: failed to unmarshal: %v", err)
	}
	var joinRespStruct2 chat.JoinRoomResponse
	if err := json.Unmarshal(joinRespBody2, &joinRespStruct2); err != nil {
		t.Fatalf("Join: failed to unmarshal: %v", err)
	}

	wsURL1 := "ws://" + env.serverAddr + "/ws/room?token=" + env.token
	wsURL2 := "ws://" + env.serverAddr + "/ws/room?token=" + joinRespStruct1.Token
	wsURL3 := "ws://" + env.serverAddr + "/ws/room?token=" + joinRespStruct2.Token

	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{
			"Origin": []string{"http://kseli.app"},
		},
	}

	conn1, _, _, err := dialer.Dial(context.Background(), wsURL1)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	conn2, _, _, err := dialer.Dial(context.Background(), wsURL2)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	conn3, _, _, err := dialer.Dial(context.Background(), wsURL3)
	if err != nil {
		t.Fatalf("ws handshake failed: %v", err)
	}

	// drain join messages
	mustReadWSJoin(t, conn1)
	mustReadWSJoin(t, conn1)
	mustReadWSJoin(t, conn1)
	mustReadWSJoin(t, conn2)
	mustReadWSJoin(t, conn2)
	mustReadWSJoin(t, conn3)

	err = wsutil.WriteClientMessage(conn2, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "leave"))
	if err != nil {
		t.Fatalf("user1 failed to send close: %v", err)
	}

	gotAdmin := mustReadWSLeave(t, conn1)
	gotUser2 := mustReadWSLeave(t, conn3)

	assertLeaveMsg(t, gotAdmin.ID, 2)
	assertLeaveMsg(t, gotUser2.ID, 2)
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

func mustReadWSJoin(t *testing.T, conn net.Conn) chat.JoinMsg {
	t.Helper()

	raw, op, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("ReadServerData failed: %v", err)
	}
	if op != ws.OpText {
		t.Fatalf("Expected OpText, got %v", op)
	}

	var wsMsg chat.WSMsg
	if err := json.Unmarshal(raw, &wsMsg); err != nil {
		t.Fatalf("Unmarshal WSMsg failed: %v", err)
	}
	if wsMsg.MsgType != "join" {
		t.Fatalf("Expected WSMsg type 'join', got %q", wsMsg.MsgType)
	}

	// Marshal then unmarshal to get proper JoinMsg type from `Data`
	jsonData, err := json.Marshal(wsMsg.Data)
	if err != nil {
		t.Fatalf("Marshal inner data failed: %v", err)
	}

	var joinMsg chat.JoinMsg
	if err := json.Unmarshal(jsonData, &joinMsg); err != nil {
		t.Fatalf("Unmarshal JoinMsg failed: %v", err)
	}
	return joinMsg
}

func assertJoinMsg(t *testing.T, msg chat.JoinMsg, wantID uint8, wantUser string, wantRole common.Role) {
	t.Helper()
	if msg.ID != wantID {
		t.Errorf("Expected ID %d, got %d", wantID, msg.ID)
	}
	if msg.Username != wantUser {
		t.Errorf("Expected Username %q, got %q", wantUser, msg.Username)
	}
	if msg.Role != wantRole {
		t.Errorf("Expected Role %q, got %q", wantRole, msg.Role)
	}
}

func mustReadWSChat(t *testing.T, conn net.Conn) chat.ChatMsg {
	t.Helper()

	raw, op, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("ReadServerData failed: %v", err)
	}
	if op != ws.OpText {
		t.Fatalf("Expected OpText, got %v", op)
	}

	var wsMsg chat.WSMsg
	if err := json.Unmarshal(raw, &wsMsg); err != nil {
		t.Fatalf("Unmarshal WSMsg failed: %v", err)
	}
	if wsMsg.MsgType != "msg" {
		t.Fatalf("Expected WSMsg type 'msg', got %q", wsMsg.MsgType)
	}

	jsonData, err := json.Marshal(wsMsg.Data)
	if err != nil {
		t.Fatalf("Marshal inner data failed: %v", err)
	}

	var chatMsg chat.ChatMsg
	if err := json.Unmarshal(jsonData, &chatMsg); err != nil {
		t.Fatalf("Unmarshal ChatMsg failed: %v", err)
	}
	return chatMsg
}

func assertChatMsg(t *testing.T, msg chat.ChatMsg, wantUser, wantContent string) {
	t.Helper()
	if msg.Username != wantUser {
		t.Errorf("Expected username %q, got %q", wantUser, msg.Username)
	}
	if msg.Content != wantContent {
		t.Errorf("Expected content %q, got %q", wantContent, msg.Content)
	}
}

func mustReadWSLeave(t *testing.T, conn net.Conn) chat.LeaveMsg {
	t.Helper()
	raw, op, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("ReadServerData failed: %v", err)
	}
	if op != ws.OpText {
		t.Fatalf("Expected OpText, got %v", op)
	}

	var wsMsg chat.WSMsg
	if err := json.Unmarshal(raw, &wsMsg); err != nil {
		t.Fatalf("Unmarshal WSMsg failed: %v", err)
	}
	if wsMsg.MsgType != "leave" {
		t.Fatalf("Expected WSMsg type 'leave', got %q", wsMsg.MsgType)
	}

	jsonData, err := json.Marshal(wsMsg.Data)
	if err != nil {
		t.Fatalf("Marshal inner data failed: %v", err)
	}

	var leaveMsg chat.LeaveMsg
	if err := json.Unmarshal(jsonData, &leaveMsg); err != nil {
		t.Fatalf("Unmarshal LeaveMsg failed: %v", err)
	}
	return leaveMsg
}

func assertLeaveMsg(t *testing.T, gotID, expectedID uint8) {
	t.Helper()
	if gotID != expectedID {
		t.Errorf("Expected leave ID %d, got %d", expectedID, gotID)
	}
}
