package handlers

type CreateRoomRequest struct {
	Username        string `json:"username" validate:"required"`
	MaxParticipants int    `json:"maxParticipants" validate:"required"`
}
