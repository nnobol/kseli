package handlers

import (
	"kseli-server/models"
)

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
	Token  string `json:"token"`
}

type JoinRoomResponse struct {
	RoomID string `json:"roomId"`
	Token  string `json:"token"`
}

type GetRoomResponse struct {
	RoomID          string        `json:"roomId"`
	MaxParticipants int           `json:"maxParticipants"`
	Participants    []models.User `json:"participants"`
	SecretKey       string        `json:"secretKey,omitempty"`
}

type ErrorResponse struct {
	StatusCode   int               `json:"statusCode"`
	ErrorMessage string            `json:"errorMessage"`
	FieldErrors  map[string]string `json:"fieldErrors,omitempty"`
}
