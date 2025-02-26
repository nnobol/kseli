package models

type sessionIDContextKey string
type fingerprintContextKey string

const UserSessionIDKey sessionIDContextKey = "sessionId"
const UserFingerprintKey fingerprintContextKey = "fingerprint"

type Role uint8

const (
	Admin  Role = 1
	Member Role = 2
)

type User struct {
	SessionId   string `json:"-"`
	Fingerprint string `json:"-"`
	ID          uint8  `json:"id"`
	Username    string `json:"username"`
	Role        Role   `json:"role"`
}
