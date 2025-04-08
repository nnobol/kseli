package chat

import (
	"net"
	"sync"
	"time"

	"kseli/common"
)

type Participant struct {
	mu        sync.Mutex
	sessionID string
	id        uint8
	username  string
	role      common.Role
	wsConn    net.Conn
	msgQueue  chan []byte
	// wsTimeout is a timer used to clean up a participant that never establishes a WS connection
	// used in room.go in the "join" method
	// stopped in ws.go in the "addWSConn" method
	wsTimeout *time.Timer
}

type ParticipantView struct {
	ID       uint8       `json:"id"`
	Username string      `json:"username,omitempty"`
	Role     common.Role `json:"role,omitempty"`
}
