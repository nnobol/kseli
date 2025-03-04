package models

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/coder/websocket"
)

type Room struct {
	Mu              sync.RWMutex     `json:"-"`
	NextUserId      uint8            `json:"-"`
	RoomID          string           `json:"id"`
	MaxParticipants uint8            `json:"maxParticipants"`
	SecretKey       *string          `json:"secretKey,omitempty"`
	Participants    map[string]*User `json:"participants"`
}

type WSMessage struct {
	MsgType string      `json:"type"`
	Data    interface{} `json:"data"`
}

type ChatMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

// make sure caller locks room for writing
func (r *Room) Join(user *User) {
	r.Participants[user.Username] = user
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
func (r *Room) GetUsernameByID(userID uint8) string {

	for username, user := range r.Participants {
		if userID == user.ID {
			return username
		}
	}

	return ""
}

func (r *Room) AddWSConnection(ctx context.Context, username string, conn *websocket.Conn) {
	r.Mu.Lock()
	user, exists := r.Participants[username]
	if !exists {
		r.Mu.Unlock()
		return
	}

	if user.WSConnection != nil {
		user.WSConnection.CloseNow()
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

	go r.handleRead(username, ctx, conn)
	go r.handleWrite(username, ctx, conn, user.MessageQueue)
}

func (r *Room) handleRead(username string, ctx context.Context, conn *websocket.Conn) {
	defer r.removeWSConnection(username)

	for {
		msgType, msg, err := conn.Read(ctx)
		if err != nil {
			break
		}

		if msgType != websocket.MessageText {
			continue
		}

		wsMessage, _ := json.Marshal(WSMessage{
			MsgType: "msg",
			Data: ChatMessage{
				Username: username,
				Content:  string(msg),
			},
		})

		r.broadcastMessage(wsMessage)
	}
}

func (r *Room) handleWrite(username string, ctx context.Context, conn *websocket.Conn, queue <-chan []byte) {
	defer r.removeWSConnection(username)

	for {
		select {
		case msg, ok := <-queue:
			if !ok {
				return
			}

			if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
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

func (r *Room) removeWSConnection(username string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	user, exists := r.Participants[username]
	if !exists {
		return
	}

	if user.WSConnection != nil {
		user.WSConnection.CloseNow()
		user.WSConnection = nil
	}

	if user.MessageQueue != nil {
		close(user.MessageQueue)
		user.MessageQueue = nil
	}
}
