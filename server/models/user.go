package models

import "net"

type sessionIDContextKey string

const UserSessionIDKey sessionIDContextKey = "sessionId"

type Role uint8

const (
	Admin  Role = 1
	Member Role = 2
)

type User struct {
	SessionId    string      `json:"-"`
	ID           uint8       `json:"id"`
	Username     string      `json:"username,omitempty"`
	Role         Role        `json:"role,omitempty"`
	WSConnection net.Conn    `json:"-"`
	MessageQueue chan []byte `json:"-"`
}
