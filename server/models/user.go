package models

type sessionIDContextKey string

const UserSessionIDKey sessionIDContextKey = "sessionId"

type Role uint8

const (
	Admin  Role = 1
	Member Role = 2
)

type User struct {
	SessionId string
	ID        uint8  `json:"id"`
	Username  string `json:"username"`
	Role      Role   `json:"role"`
}
