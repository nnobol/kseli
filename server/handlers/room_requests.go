package handlers

type CreateRoomRequest struct {
	Username        string `json:"username" validate:"required"`
	MaxParticipants int    `json:"maxParticipants" validate:"required"`
}

type JoinRoomRequest struct {
	Username      string `json:"username" validate:"required"`
	RoomSecretKey string `json:"roomSecretKey" validate:"required"`
}
