package chat

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"kseli/common"
)

type Room struct {
	mu                 sync.RWMutex
	nextParticipantID  uint8
	roomID             string
	maxParticipants    uint8
	secretKey          string
	participants       map[string]*Participant // key sessionID
	bannedParticipants map[string]struct{}     // key sessionID
	onClose            func(roomID string)
	onExpire           *time.Timer
	expiresAt          int64
}

type Storage interface {
	AddRoom(roomID string, room *Room)
	GetRoom(roomID string) (*Room, bool)
	DeleteRoom(roomID string)
	RoomCleanupFunc() func(roomID string)
}

// make sure caller locks room for reading
func (r *Room) getParticipantByID(ID uint8) (*Participant, bool) {
	for _, p := range r.participants {
		if ID == p.id {
			return p, true
		}
	}

	return nil, false
}

// make sure caller locks room for reading
func (r *Room) getParticipantByUsername(username string) (*Participant, bool) {
	for _, p := range r.participants {
		if username == p.username {
			return p, true
		}
	}

	return nil, false
}

// make sure caller locks room for reading
func (r *Room) getParticipantsAsSlice() []ParticipantView {
	pSlice := make([]ParticipantView, 0, len(r.participants))

	for _, p := range r.participants {
		pView := ParticipantView{
			ID:       p.id,
			Username: p.username,
			Role:     p.role,
		}
		pSlice = append(pSlice, pView)
	}

	return pSlice
}

// make sure caller locks room for reading
func (r *Room) isUsernameTaken(username string) bool {
	for _, p := range r.participants {
		if username == p.username {
			return true
		}
	}

	return false
}

// make sure caller locks room for rw
func (r *Room) join(p *Participant) {
	r.participants[p.sessionID] = p

	// Start 10s timeout to wait for WebSocket connection
	p.wsTimeout = time.AfterFunc(10*time.Second, func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		// If WS is still not connected, remove the participant
		if p.wsConn == nil {
			delete(r.participants, p.sessionID)
		}
	})
}

func (r *Room) kick(pID uint8) error {
	r.mu.RLock()
	p, exists := r.getParticipantByID(pID)
	r.mu.RUnlock()
	if !exists {
		return fmt.Errorf("Participant with ID '%d' not found in room", pID)
	}

	r.mu.Lock()
	delete(r.participants, p.sessionID)
	r.mu.Unlock()

	go func() {
		p.cleanupWSConn("kick", true)
		r.broadcastLeave(p.id)
	}()

	return nil
}

func (r *Room) ban(pID uint8) error {
	r.mu.RLock()
	p, exists := r.getParticipantByID(pID)
	r.mu.RUnlock()
	if !exists {
		return fmt.Errorf("Participant with ID '%d' not found in room", pID)
	}

	r.mu.Lock()
	r.bannedParticipants[p.sessionID] = struct{}{}
	delete(r.participants, p.sessionID)
	r.mu.Unlock()

	go func() {
		p.cleanupWSConn("ban", true)
		r.broadcastLeave(p.id)
	}()

	return nil
}

func (r *Room) Close(isScheduled bool) {
	r.mu.Lock()
	participants := make([]*Participant, 0, len(r.participants))

	for _, p := range r.participants {
		participants = append(participants, p)
	}

	r.participants = nil
	r.bannedParticipants = nil

	r.onExpire.Stop()
	r.onExpire = nil

	onClose := r.onClose
	r.onClose = nil

	roomID := r.roomID
	r.mu.Unlock()

	for _, p := range participants {
		var reason string

		if isScheduled {
			reason = "close"
		} else {
			if p.role == common.Admin {
				reason = "close-admin"
			} else {
				reason = "close-user"
			}
		}

		go p.cleanupWSConn(reason, true)
	}

	if onClose != nil {
		go onClose(roomID)
	}
}

func generateUniqueRoomID(s Storage) string {
	for {
		roomID := generateRoomID()

		_, exists := s.GetRoom(roomID)

		if !exists {
			return roomID
		}
	}
}

func generateRoomID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateSecretKey() string {
	b := make([]byte, 10)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
