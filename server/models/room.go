package models

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Room struct {
	Mu              sync.RWMutex        `json:"-"`
	NextUserId      uint8               `json:"-"`
	RoomID          string              `json:"id"`
	MaxParticipants uint8               `json:"maxParticipants"`
	SecretKey       *string             `json:"secretKey,omitempty"`
	Participants    map[string]*User    `json:"participants"` // sessionID -> User
	BannedUsers     map[string]struct{} `json:"-"`            // sessionID
	OnClose         func(roomID string) `json:"-"`
	OnExpire        *time.Timer         `json:"-"`
}

type WSMessage struct {
	MsgType string      `json:"type"`
	Data    interface{} `json:"data"`
}

type ChatMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

type CloseMessage struct {
	Reason string `json:"reason"`
}

type LeaveMessage struct {
	ID uint8 `json:"id"`
}

// make sure caller locks room for writing
func (r *Room) Join(user *User) {
	r.Participants[user.SessionId] = user
}

func (r *Room) Kick(targetUserID uint8) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	targetUser, exists := r.GetUserByID(targetUserID)
	if !exists {
		return fmt.Errorf("User with ID '%d' not found in room", targetUserID)
	}

	if targetUser.WSConnection != nil {

		wsutil.WriteServerMessage(targetUser.WSConnection, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "kick"))

		targetUser.WSConnection.Close()
		targetUser.WSConnection = nil
	}

	if targetUser.MessageQueue != nil {
		close(targetUser.MessageQueue)
		targetUser.MessageQueue = nil
	}

	delete(r.Participants, targetUser.SessionId)

	go r.broadcastLeave(targetUser.ID)

	return nil
}

func (r *Room) Ban(targetUserID uint8) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	targetUser, exists := r.GetUserByID(targetUserID)
	if !exists {
		return fmt.Errorf("User with ID '%d' not found in room", targetUserID)
	}

	_, alreadyBanned := r.BannedUsers[targetUser.SessionId]
	if alreadyBanned {
		return fmt.Errorf("User with ID '%d' is already banned", targetUserID)
	}

	if targetUser.WSConnection != nil {

		wsutil.WriteServerMessage(targetUser.WSConnection, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "ban"))

		targetUser.WSConnection.Close()
		targetUser.WSConnection = nil
	}

	if targetUser.MessageQueue != nil {
		close(targetUser.MessageQueue)
		targetUser.MessageQueue = nil
	}

	r.BannedUsers[targetUser.SessionId] = struct{}{}

	delete(r.Participants, targetUser.SessionId)

	go r.broadcastLeave(targetUser.ID)

	return nil
}

func (r *Room) Close(roomID string, isScheduledCleanup bool) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	for _, user := range r.Participants {
		if user.WSConnection != nil {
			if isScheduledCleanup {
				wsutil.WriteServerMessage(user.WSConnection, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "close"))
			} else {
				if user.Role == Admin {
					wsutil.WriteServerMessage(user.WSConnection, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "close-admin"))
				} else {
					wsutil.WriteServerMessage(user.WSConnection, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "close-user"))
				}
			}
			user.WSConnection.Close()
			user.WSConnection = nil
		}

		if user.MessageQueue != nil {
			close(user.MessageQueue)
			user.MessageQueue = nil
		}

		delete(r.Participants, user.SessionId)
	}

	r.OnClose(r.RoomID)
}

// make sure caller locks room for reading
func (r *Room) GetParticipantsAsSlice() []User {
	pSlice := make([]User, 0, len(r.Participants))

	for _, user := range r.Participants {
		pSlice = append(pSlice, *user)
	}

	return pSlice
}

// make sure caller locks room for reading
func (r *Room) GetUserByUsername(username string) (*User, bool) {
	for _, user := range r.Participants {
		if username == user.Username {
			return user, true
		}
	}

	return nil, false
}

// make sure caller locks room for reading
func (r *Room) GetUserByID(userID uint8) (*User, bool) {
	for _, user := range r.Participants {
		if userID == user.ID {
			return user, true
		}
	}

	return nil, false
}

// make sure caller locks room for reading
func (r *Room) IsUsernameTaken(username string) bool {
	for _, user := range r.Participants {
		if username == user.Username {
			return true
		}
	}

	return false
}

func (r *Room) AddWSConnection(conn net.Conn, username string) {
	r.Mu.Lock()
	// User should most definitely exist at this point, look into if this is needed.
	user, exists := r.GetUserByUsername(username)
	if !exists {
		r.Mu.Unlock()
		wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "user-not-exists"))
		conn.Close()
		return
	}

	// Both of these properties should be nil, look into this as well
	if user.WSConnection != nil {
		user.WSConnection.Close()
	}
	if user.MessageQueue != nil {
		close(user.MessageQueue)
	}

	user.WSConnection = conn
	user.MessageQueue = make(chan []byte, 20)
	r.Mu.Unlock()

	// edge case: user or room might get deleted by some go routine and panic on user member access
	wsMessage, _ := json.Marshal(WSMessage{
		MsgType: "join",
		Data: User{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		},
	})

	r.broadcastMessage(wsMessage)

	pongChan := make(chan struct{}, 1)
	go r.handleRead(conn, username, pongChan)
	go r.handleWrite(conn, username, user.MessageQueue, pongChan)
}

func (r *Room) handleRead(conn net.Conn, username string, pongChan chan<- struct{}) {
	var cleanup bool

	defer func() {
		if cleanup {
			close(pongChan)
			r.removeUserFromRoom(username)
		}
	}()

	for {
		// if "leave" is not received (tab close, browser close), we break and then clean up when ping fails from handleWrite
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			if closeErr, ok := err.(wsutil.ClosedError); ok {
				if closeErr.Reason == "leave" {
					cleanup = true
				}
			}
			break
		}

		switch op {
		case ws.OpText:
			wsMessage, _ := json.Marshal(WSMessage{
				MsgType: "msg",
				Data: ChatMessage{
					Username: username,
					Content:  string(msg),
				},
			})

			r.broadcastMessage(wsMessage)

		case ws.OpBinary:
			select {
			case pongChan <- struct{}{}:
			default:
			}

		default:
		}
	}
}

func (r *Room) handleWrite(conn net.Conn, username string, queue <-chan []byte, pongChan chan struct{}) {
	var cleanup bool

	defer func() {
		if cleanup {
			close(pongChan)
			r.removeUserFromRoom(username)
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
		case msg, ok := <-queue:
			if !ok {
				return
			}

			// what happens if this fails, we don't clean up here so need to make sure to clean up resources
			if err := wsutil.WriteServerMessage(conn, ws.OpText, msg); err != nil {
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

func (r *Room) broadcastMessage(msg []byte) {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	for _, user := range r.Participants {
		select {
		case user.MessageQueue <- msg:
		default:
		}
	}
}

func (r *Room) broadcastLeave(userID uint8) {
	leaveMsg, _ := json.Marshal(WSMessage{
		MsgType: "leave",
		Data: LeaveMessage{
			ID: userID,
		},
	})

	r.broadcastMessage(leaveMsg)
}

func (r *Room) removeUserFromRoom(username string) {
	r.Mu.RLock()
	user, exists := r.GetUserByUsername(username)
	if !exists {
		r.Mu.RUnlock()
		return
	}
	r.Mu.RUnlock()

	if user.Role == Admin {
		r.Close(r.RoomID, false)
	} else {
		r.Mu.Lock()
		defer r.Mu.Unlock()

		if user.WSConnection != nil {
			user.WSConnection.Close()
			user.WSConnection = nil
		}

		if user.MessageQueue != nil {
			close(user.MessageQueue)
			user.MessageQueue = nil
		}

		delete(r.Participants, user.SessionId)

		go r.broadcastLeave(user.ID)
	}
}
