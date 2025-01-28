package handlers

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
	Token  string `json:"token"`
}

type JoinRoomResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	StatusCode   int               `json:"statusCode"`
	ErrorMessage string            `json:"errorMessage"`
	FieldErrors  map[string]string `json:"fieldErrors,omitempty"`
}
