package models

import "sync"

type Room struct {
	Mu              sync.RWMutex
	NextUserId      uint8
	RoomID          string           `json:"id"`
	MaxParticipants uint8            `json:"maxParticipants"`
	SecretKey       *string          `json:"secretKey,omitempty"`
	Participants    map[string]*User `json:"participants"`
}

func (r *Room) Join(user *User) {
	r.Mu.Lock()
	r.Participants[user.Username] = user
	r.Mu.Unlock()
}

// make sure caller locks room for reading
func (r *Room) GetParticipantsAsSlice() []User {
	pSlice := make([]User, 0, len(r.Participants))

	for _, user := range r.Participants {
		pSlice = append(pSlice, *user)
	}

	return pSlice
}
