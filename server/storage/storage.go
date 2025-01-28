package storage

import (
	"fmt"
	"sync"

	"kseli-server/models"
	"kseli-server/util"
)

type Room struct {
	mu              sync.RWMutex
	ID              string                  `json:"id"`
	MaxParticipants int                     `json:"maxParticipants"`
	SecretKey       string                  `json:"secretKey"`
	Participants    map[string]*models.User `json:"participants"`
}

type RoomStorage struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewMemoryStore() *RoomStorage {
	return &RoomStorage{
		rooms: make(map[string]*Room),
	}
}

func (storage *RoomStorage) CreateRoom(adminUser *models.User, maxParticipants int) string {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	var roomID, secretKey string
	for {
		roomID = util.GenerateRoomIDFunc()
		secretKey = util.GenerateRoomIDFunc()
		if _, exists := storage.rooms[roomID]; !exists {
			break
		}
	}

	fmt.Print(secretKey)

	room := &Room{
		ID:              roomID,
		MaxParticipants: maxParticipants,
		SecretKey:       secretKey,
		Participants:    make(map[string]*models.User),
	}

	room.Participants[adminUser.Username] = adminUser

	storage.rooms[roomID] = room
	return roomID
}

func (room *Room) JoinRoom(user *models.User) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.Participants[user.Username] = user
}

func (storage *RoomStorage) GetRoom(roomID string) (*Room, error) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	room, exists := storage.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	return room, nil
}
