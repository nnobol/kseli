package chat_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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
	createResp, _ := createRoom(t, true, 0, mux, 3, "admin", "http://kseli.app", config.APIKey, "admin")

	// 2) Fetch invite token via get room as an admin
	inviteToken, _, _ := getRoom(t, true, true, 0, mux, createResp.RoomID, "http://kseli.app", createResp.Token)

	server := httptest.NewServer(mux)
	serverAddr := server.Listener.Addr().String()

	return &roomWSEnv{
		token:       createResp.Token,
		inviteToken: inviteToken,
		roomID:      createResp.RoomID,
		server:      server,
		serverAddr:  serverAddr,
		mux:         mux,
	}
}

func Test_RoomWS_Success_JoinMsgReceived(t *testing.T) {
	env := newRoomWSEnv(t)

	// join request to add a new user to the room
	joinResp, _ := joinRoom(t, true, 0, env.mux, "user", "http://kseli.app", env.inviteToken, "user")

	wsURL1 := "ws://" + env.serverAddr + "/ws/room?token=" + env.token
	wsURL2 := "ws://" + env.serverAddr + "/ws/room?token=" + joinResp.Token

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
	joinResp, _ := joinRoom(t, true, 0, env.mux, "user", "http://kseli.app", env.inviteToken, "user")

	wsURL1 := "ws://" + env.serverAddr + "/ws/room?token=" + env.token
	wsURL2 := "ws://" + env.serverAddr + "/ws/room?token=" + joinResp.Token

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

	// two join requests to add 2 other users to the room
	joinResp1, _ := joinRoom(t, true, 0, env.mux, "user1", "http://kseli.app", env.inviteToken, "user1")
	joinResp2, _ := joinRoom(t, true, 0, env.mux, "user2", "http://kseli.app", env.inviteToken, "user2")

	wsURL1 := "ws://" + env.serverAddr + "/ws/room?token=" + env.token
	wsURL2 := "ws://" + env.serverAddr + "/ws/room?token=" + joinResp1.Token
	wsURL3 := "ws://" + env.serverAddr + "/ws/room?token=" + joinResp2.Token

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
	deleteRoom(t, true, 0, env.mux, env.roomID, "http://kseli.app", env.token)

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
	joinResp, _ := joinRoom(t, true, 0, env.mux, "user", "http://kseli.app", env.inviteToken, "user")

	// 2) kick the user from the room
	kickOrBanUser(t, true, 0, env.mux, 2, "kick", env.roomID, "http://kseli.app", env.token)

	// 2) try connecting to the ws
	wsURL := "ws://" + env.serverAddr + "/ws/room?token=" + joinResp.Token

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
