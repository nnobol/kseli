package storage

import (
	"crypto/rand"
	"encoding/base64"
	"sync"

	"kseli-server/models"
)

type Room struct {
	ID              string         `json:"id"`
	Admin           *models.User   `json:"admin"`
	Participants    []*models.User `json:"participants"`
	MaxParticipants int            `json:"maxParticipants"`
}

type RoomStorage struct {
	rwMutex sync.RWMutex
	rooms   map[string]*Room
}

func NewMemoryStore() *RoomStorage {
	return &RoomStorage{
		rooms: make(map[string]*Room),
	}
}

func (storage *RoomStorage) CreateRoom(adminUser *models.User, maxParticipants int) string {
	storage.rwMutex.Lock()
	defer storage.rwMutex.Unlock()

	var roomID string
	for {
		roomID = generateRoomID()
		if _, exists := storage.rooms[roomID]; !exists {
			break
		}
	}

	room := &Room{
		ID:              roomID,
		Admin:           adminUser,
		Participants:    []*models.User{adminUser},
		MaxParticipants: maxParticipants,
	}

	storage.rooms[roomID] = room
	return roomID
}

func generateRoomID() string {
	randomBytes := make([]byte, 6)
	rand.Read(randomBytes)
	return base64.RawURLEncoding.EncodeToString(randomBytes)
}
