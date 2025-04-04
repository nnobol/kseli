package storage

import (
	"sync"

	"kseli-server/features/chat"
)

var cleanupChan = make(chan string, 50)

type MainStorage struct {
	mu    sync.RWMutex
	rooms map[string]*chat.Room // key roomID
}

func InitializeStorage() *MainStorage {
	storage := &MainStorage{
		rooms: make(map[string]*chat.Room),
	}

	go storage.cleanupRooms()

	return storage
}

func (s *MainStorage) AddRoom(roomID string, room *chat.Room) {
	s.mu.Lock()
	s.rooms[roomID] = room
	s.mu.Unlock()
}

func (s *MainStorage) GetRoom(roomID string) (*chat.Room, bool) {
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

func (s *MainStorage) RoomCleanupFunc() func(roomID string) {
	return func(roomID string) {
		cleanupChan <- roomID
	}
}

func (s *MainStorage) cleanupRooms() {
	for roomID := range cleanupChan {
		if room, exists := s.GetRoom(roomID); exists {
			room.Close(true)
		}
	}
}
