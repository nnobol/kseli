package api

import "kseli-server/models"

type CreateRoomRequest struct {
	Username        string `json:"username"`
	MaxParticipants uint8  `json:"maxParticipants"`
}

type JoinRoomRequest struct {
	Username      string `json:"username"`
	RoomSecretKey string `json:"roomSecretKey"`
}

type UserRequest struct {
	TargetUserID uint8 `json:"userId"`
}

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
	Token  string `json:"token"`
}

type JoinRoomResponse struct {
	Token string `json:"token"`
}

type GetRoomResponse struct {
	UserRole        models.Role   `json:"userRole"`
	MaxParticipants uint8         `json:"maxParticipants"`
	Participants    []models.User `json:"participants"`
	SecretKey       string        `json:"secretKey,omitempty"`
}
