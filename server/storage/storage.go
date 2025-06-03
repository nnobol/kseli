package storage

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"kseli/features/chat"
)

var (
	DayMinRooms = 300
	DayMaxRooms = 700
)

var cleanupChan = make(chan string, 50)

type MainStorage struct {
	mu    sync.RWMutex
	rooms map[string]*chat.Room // key roomID

	simMu              sync.Mutex
	lastRooms          int
	cachedRooms        int
	cachedParticipants int
}

func InitializeStorage() *MainStorage {
	initRooms := DayMinRooms + rand.Intn(DayMaxRooms-DayMinRooms+1)

	storage := &MainStorage{
		rooms:              make(map[string]*chat.Room),
		lastRooms:          initRooms,
		cachedRooms:        0,
		cachedParticipants: 0,
	}

	storage.updateMetrics()

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			storage.updateMetrics()
		}
	}()

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

func (s *MainStorage) MetricsSnapshot() (roomCount, participantCount int) {
	s.simMu.Lock()
	roomCount = s.cachedRooms
	participantCount = s.cachedParticipants
	s.simMu.Unlock()
	return
}

func (s *MainStorage) updateMetrics() {
	s.mu.RLock()
	realRooms := len(s.rooms)
	realParticipants := 0
	for _, room := range s.rooms {
		realParticipants += room.GetParticipantsLen()
	}
	s.mu.RUnlock()

	minFake := DayMinRooms - realRooms
	if minFake < 0 {
		minFake = 0
	}
	maxFake := DayMaxRooms - realRooms
	if maxFake < 0 {
		maxFake = 0
	}

	const maxJitter = 12

	s.simMu.Lock()
	{
		j := rand.Intn(2*maxJitter+1) - maxJitter
		cand := s.lastRooms + j

		if cand < minFake {
			cand = minFake
		}
		if cand > maxFake {
			cand = maxFake
		}
		s.lastRooms = cand
	}

	fakeRooms := s.lastRooms

	const avgPerRoom = 3
	baseParticipants := fakeRooms * avgPerRoom
	deltaParticipants := maxJitter * avgPerRoom
	fakeParticipants := baseParticipants + deltaParticipants

	hardMin := fakeRooms * 2
	hardMax := fakeRooms * 5
	if fakeParticipants < hardMin {
		fakeParticipants = hardMin
	}
	if fakeParticipants > hardMax {
		fakeParticipants = hardMax
	}

	totalRooms := realRooms + fakeRooms
	totalParticipants := realParticipants + fakeParticipants

	now := time.Now()
	slot := now.Hour()*2 + now.Minute()/30
	multiplier := usageMultiplier(slot)

	totalRooms = int(float64(totalRooms) * multiplier)
	if totalRooms < 1 {
		totalRooms = 1
	}

	totalParticipants = int(float64(totalParticipants) * multiplier)
	if totalParticipants < 1 {
		totalParticipants = 1
	}

	s.cachedRooms = totalRooms
	s.cachedParticipants = totalParticipants
	s.simMu.Unlock()
}

func usageMultiplier(slot int) float64 {
	if slot < 0 || slot >= 48 {
		return 1.0
	}

	shifted := (slot + 12) % 48
	theta := (float64(shifted) / 48.0) * 2.0 * math.Pi

	raw := 0.5 + 0.5*math.Cos(theta)

	min := 0.03
	max := 1.0
	return min + (max-min)*raw
}
