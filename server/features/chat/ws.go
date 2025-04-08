package chat

import (
	"encoding/json"
	"net"
	"time"

	"kseli/common"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type wsMsg struct {
	MsgType string      `json:"type"`
	Data    interface{} `json:"data"`
}

type chatMsg struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

type joinMsg struct {
	ID       uint8       `json:"id"`
	Username string      `json:"username"`
	Role     common.Role `json:"role"`
}

type leaveMsg struct {
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
	// WS connection established, we stop the timer
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

	defer func() {
		if cleanup {
			close(pongChan)
			r.rmParticipantFromRoom(username)
		}
	}()

	for {
		// If "leave" is not received, the action on the client side might be a refresh and a tab or browser close
		// On tab or browser close, we won't receive a ping so we clean up when ping fails in handleWrite
		// In case of refresh, WS conn will be reastablished and ping won't fail so that is why we just break in that case
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			if closeErr, ok := err.(wsutil.ClosedError); ok {
				if closeErr.Reason == "leave" {
					cleanup = true
				}
			}
			return
		}

		switch op {
		case ws.OpText:
			r.broadcastChatMsg(username, string(msg))

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

	pingInterval := 15 * time.Second
	timeoutDuration := 20 * time.Second

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

			if err := wsutil.WriteServerMessage(conn, ws.OpBinary, []byte{0}); err != nil {
				cleanup = true
				return
			}
			hasSentFirstPing = true
		}
	}
}

func (p *Participant) cleanupWSConn(reason string, sendCloseFrame bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.wsConn != nil {
		if sendCloseFrame {
			wsutil.WriteServerMessage(p.wsConn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, reason))
		}
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

		p.cleanupWSConn("", false)
		r.broadcastLeave(p.id)
	}
}

func (r *Room) broadcastChatMsg(username, content string) {
	msg := encodeWSMessage("msg", chatMsg{
		Username: username,
		Content:  content,
	})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastJoin(id uint8, uname string, role common.Role) {
	msg := encodeWSMessage("join", joinMsg{
		ID:       id,
		Username: uname,
		Role:     role,
	})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastLeave(pID uint8) {
	msg := encodeWSMessage("leave", leaveMsg{ID: pID})
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
	msg := wsMsg{
		MsgType: msgType,
		Data:    data,
	}
	result, _ := json.Marshal(msg)
	return result
}
