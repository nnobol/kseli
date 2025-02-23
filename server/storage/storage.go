package storage

import (
	"crypto/rand"
	"encoding/base64"
	"sync"

	"kseli-server/models"
)

type MainStorage struct {
	mu    sync.RWMutex
	rooms map[string]*models.Room
}

func InitializeStorage() *MainStorage {
	return &MainStorage{
		rooms: make(map[string]*models.Room),
	}
}

func (s *MainStorage) CreateRoom(adminUser *models.User, maxParticipants uint8) string {
	roomID := generateUniqueRoomID(s)
	roomSecretKey := generateSecretKey()

	room := &models.Room{
		NextUserId:      2,
		RoomID:          roomID,
		MaxParticipants: maxParticipants,
		SecretKey:       &roomSecretKey,
		Participants:    make(map[string]*models.User),
	}

	room.Participants[adminUser.Username] = adminUser

	s.mu.Lock()
	s.rooms[roomID] = room
	s.mu.Unlock()

	return roomID
}

func (s *MainStorage) GetRoom(roomID string) (*models.Room, bool) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	return room, exists
}

func generateUniqueRoomID(s *MainStorage) string {
	for {
		roomID := generateRoomID()

		_, exists := s.GetRoom(roomID)

		if !exists {
			return roomID
		}
	}
}

func generateRoomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateSecretKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
