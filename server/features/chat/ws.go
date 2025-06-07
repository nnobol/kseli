package chat

import (
	"encoding/json"
	"io"
	"net"
	"time"

	"kseli/common"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WSMsg struct {
	MsgType string      `json:"type"`
	Data    interface{} `json:"data"`
}

type ChatMsg struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

type JoinMsg struct {
	ID       uint8       `json:"id"`
	Username string      `json:"username"`
	Role     common.Role `json:"role"`
}

type LeaveMsg struct {
	ID uint8 `json:"id"`
}

func (r *Room) addWSConn(conn net.Conn, username string) {
	r.mu.RLock()
	p, exists := r.getParticipantByUsername(username)
	if !exists {
		r.mu.RUnlock()
		wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "user-not-exists"))
		conn.Close()
		return
	}

	id := p.id
	role := p.role
	r.mu.RUnlock()

	r.mu.Lock()
	// WS connection established, we stop the timeout timer
	if p.wsTimeout != nil {
		p.wsTimeout.Stop()
		p.wsTimeout = nil
	}
	if p.wsConn != nil {
		p.wsConn.Close()
		p.wsConn = nil
	}
	if p.msgQueue != nil {
		close(p.msgQueue)
		p.msgQueue = nil
	}

	p.wsConn = conn
	p.msgQueue = make(chan []byte, 20)
	r.mu.Unlock()

	pongChan := make(chan struct{}, 1)

	go r.broadcastJoin(id, username, role)
	go r.handleRead(conn, username, pongChan)
	go r.handleWrite(conn, username, p.msgQueue, pongChan)
}

func (r *Room) handleRead(conn net.Conn, username string, pongChan chan<- struct{}) {
	var cleanup bool
	const maxMsgSize = 1024

	defer func() {
		if cleanup {
			close(pongChan)
			r.rmParticipantFromRoom(username)
		}
	}()

	msgReader := wsutil.NewReader(conn, ws.StateServerSide)
	for {

		hdr, err := msgReader.NextFrame()
		if err != nil {
			return
		}

		lr := io.LimitReader(msgReader, maxMsgSize+1)
		buf := make([]byte, maxMsgSize+1)
		n, _ := io.ReadFull(lr, buf)

		if n > maxMsgSize {
			wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "message-too-large"))
			cleanup = true
			return
		}

		switch hdr.OpCode {
		case ws.OpText:
			r.broadcastChatMsg(username, string(buf[:n]))

		case ws.OpClose:
			_, reason := ws.ParseCloseFrameData(buf[:n])
			if reason == "leave" {
				// If "leave" is not received, the action on the client side might be a refresh.
				// In case of refresh, WS conn will be reestablished and that is why we don't clean up.
				// We do however clean up in case "leave" is received.
				cleanup = true
			}
			return

		case ws.OpBinary:
			select {
			case pongChan <- struct{}{}:
			default:
			}

		default:
		}
	}
}

func (r *Room) handleWrite(conn net.Conn, username string, msgQueue <-chan []byte, pongChan chan struct{}) {
	var cleanup bool

	defer func() {
		if cleanup {
			close(pongChan)
			r.rmParticipantFromRoom(username)
		}
	}()

	pingInterval := 10 * time.Second
	timeoutDuration := 30 * time.Second

	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

	lastPong := time.Now()
	hasSentFirstPing := false

	for {
		select {
		case msg, ok := <-msgQueue:
			if !ok {
				return
			}

			if err := wsutil.WriteServerMessage(conn, ws.OpText, msg); err != nil {
				cleanup = true
				return
			}

		case <-pongChan:
			lastPong = time.Now()

		case <-pingTicker.C:
			if hasSentFirstPing && time.Since(lastPong) > timeoutDuration {
				cleanup = true
				return
			}

			wsutil.WriteServerMessage(conn, ws.OpBinary, []byte{0})
			hasSentFirstPing = true
		}
	}
}

func (p *Participant) cleanupWSConn(reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.wsConn != nil {
		wsutil.WriteServerMessage(p.wsConn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, reason))
		p.wsConn.Close()
		p.wsConn = nil
	}

	if p.msgQueue != nil {
		close(p.msgQueue)
		p.msgQueue = nil
	}
}

func (r *Room) rmParticipantFromRoom(username string) {
	r.mu.RLock()
	p, exists := r.getParticipantByUsername(username)
	if !exists {
		r.mu.RUnlock()
		return
	}
	role := p.role
	r.mu.RUnlock()

	if role == common.Admin {
		// Admin disconnects -> room shuts down
		// The participant will be an Admin only in case of ping-pong failure
		// Otherwise r.Close will be called from a handler
		r.Close(false)
	} else {
		r.mu.Lock()
		delete(r.participants, p.sessionID)
		r.mu.Unlock()

		p.cleanupWSConn("")
		r.broadcastLeave(p.id)
	}
}

func (r *Room) broadcastChatMsg(username, content string) {
	msg := encodeWSMessage("msg", ChatMsg{
		Username: username,
		Content:  content,
	})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastJoin(id uint8, uname string, role common.Role) {
	msg := encodeWSMessage("join", JoinMsg{
		ID:       id,
		Username: uname,
		Role:     role,
	})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastLeave(pID uint8) {
	msg := encodeWSMessage("leave", LeaveMsg{ID: pID})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastMessage(msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.participants {
		select {
		case p.msgQueue <- msg:
		default:
		}
	}
}

func encodeWSMessage(msgType string, data interface{}) []byte {
	msg := WSMsg{
		MsgType: msgType,
		Data:    data,
	}
	result, _ := json.Marshal(msg)
	return result
}
