package storage

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"kseli-server/models"
)

var cleanupChan = make(chan string, 50)

type MainStorage struct {
	mu    sync.RWMutex
	rooms map[string]*models.Room
}

func InitializeStorage() *MainStorage {
	storage := &MainStorage{
		rooms: make(map[string]*models.Room),
	}

	go storage.cleanupRooms()

	return storage
}

func (s *MainStorage) CreateRoom(adminUser *models.User, maxParticipants uint8) string {
	roomID := generateUniqueRoomID(s)
	roomSecretKey := generateSecretKey()

	room := &models.Room{
		NextUserId:      2,
		RoomID:          roomID,
		MaxParticipants: maxParticipants,
		SecretKey:       &roomSecretKey,
		Participants:    make(map[string]*models.User, maxParticipants),
		BannedUsers:     make(map[string]struct{}),
		OnClose: func(roomID string) {
			s.DeleteRoom(roomID)
		},
		OnExpire: time.AfterFunc(30*time.Minute, func() {
			cleanupChan <- roomID
		}),
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute).Unix(),
	}

	room.Participants[adminUser.SessionId] = adminUser

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

func (s *MainStorage) DeleteRoom(roomID string) {
	s.mu.Lock()
	delete(s.rooms, roomID)
	s.mu.Unlock()
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

func (s *MainStorage) cleanupRooms() {
	for roomID := range cleanupChan {
		if room, exists := s.GetRoom(roomID); exists {
			room.Close(roomID, true)
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
