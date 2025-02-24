package services

import (
	"net/http"
	"strings"
	"time"

	"kseli-server/auth"
	"kseli-server/models"
	"kseli-server/models/api"
	"kseli-server/storage"
)

type RoomService struct {
	s *storage.MainStorage
}

func NewRoomService(s *storage.MainStorage) *RoomService {
	return &RoomService{
		s: s,
	}
}

func (rs *RoomService) CreateRoom(username string, maxParticipants uint8, sessionID string, fingerprint string) (*api.CreateRoomResponse, *api.ErrorResponse) {
	fieldErrors := make(map[string]string, 2)

	validateUsername(username, fieldErrors)
	validateMaxParticipants(maxParticipants, fieldErrors)

	if len(fieldErrors) > 0 {
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	adminUser := &models.User{
		SessionId:   sessionID,
		Fingerprint: fingerprint,
		ID:          1,
		Username:    username,
		Role:        models.Admin,
	}

	roomID := rs.s.CreateRoom(adminUser, maxParticipants)

	claims := models.Claims{
		UserID: adminUser.ID,
		RoomID: roomID,
		Role:   adminUser.Role,
		Exp:    time.Now().Add(time.Hour).Unix(),
	}

	token, err := auth.CreateToken(claims)
	if err != nil {
		// TODO delete the room
		return nil, &api.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Failed to create token: " + err.Error(),
		}
	}

	return &api.CreateRoomResponse{
		RoomID: roomID,
		Token:  token,
	}, nil
}

func (rs *RoomService) JoinRoom(roomID string, username string, secretKey string, sessionID string, fingerprint string) (*api.JoinRoomResponse, *api.ErrorResponse) {
	fieldErrors := make(map[string]string, 3)

	validateRoomId(roomID, fieldErrors)
	validateUsername(username, fieldErrors)
	validateSecretKey(secretKey, fieldErrors)

	if len(fieldErrors) > 0 {
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	room, exists := rs.s.GetRoom(roomID)
	if !exists {
		fieldErrors["roomId"] = "Chat Room not found."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	if secretKey != *room.SecretKey {
		fieldErrors["roomSecretKey"] = "Incorrect Secret Key."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	if uint8(len(room.Participants)) == room.MaxParticipants {
		fieldErrors["roomId"] = "Chat Room is full."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	if _, usernameTaken := room.Participants[username]; usernameTaken {
		fieldErrors["username"] = "This username is taken."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	room.Mu.Lock()
	userId := room.NextUserId
	room.NextUserId++
	room.Mu.Unlock()

	user := &models.User{
		SessionId:   sessionID,
		Fingerprint: fingerprint,
		ID:          userId,
		Username:    username,
		Role:        models.Member,
	}

	room.Join(user)

	claims := models.Claims{
		UserID: user.ID,
		RoomID: room.RoomID,
		Role:   user.Role,
		Exp:    time.Now().Add(time.Hour).Unix(),
	}

	token, err := auth.CreateToken(claims)
	if err != nil {
		// TODO remove user from the room
		return nil, &api.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Failed to create token: " + err.Error(),
		}
	}

	return &api.JoinRoomResponse{
		Token: token,
	}, nil
}

func (rs *RoomService) GetRoom(roomID string, userClaims *models.Claims) (*api.GetRoomResponse, *api.ErrorResponse) {
	room, exists := rs.s.GetRoom(roomID)
	if !exists {
		return nil, &api.ErrorResponse{
			StatusCode:   http.StatusNotFound,
			ErrorMessage: "Chat Room not found.",
		}
	}

	if userClaims.RoomID != room.RoomID {
		return nil, &api.ErrorResponse{
			StatusCode:   http.StatusForbidden,
			ErrorMessage: "You do not have access to this room.",
		}
	}

	room.Mu.RLock()
	participants := room.GetParticipantsAsSlice()

	// secretKey will be omitted for non-admin users
	secretKey := ""
	if userClaims.Role == models.Admin {
		secretKey = *room.SecretKey
	}
	room.Mu.RUnlock()

	return &api.GetRoomResponse{
		MaxParticipants: room.MaxParticipants,
		Participants:    participants,
		SecretKey:       secretKey,
	}, nil
}

func validateRoomId(roomID string, fieldErrors map[string]string) {
	if strings.Contains(roomID, " ") {
		fieldErrors["roomId"] = "Chat Room Id cannot contain spaces."
		return
	}
}

func validateUsername(username string, fieldErrors map[string]string) {
	if username == "" {
		fieldErrors["username"] = "Username cannot be empty."
		return
	}

	if strings.Contains(username, " ") {
		fieldErrors["username"] = "Username cannot contain spaces."
		return
	}

	if len(username) < 3 || len(username) > 20 {
		fieldErrors["username"] = "Username must be between 3 and 20 characters."
	}
}

func validateMaxParticipants(maxParticipants uint8, fieldErrors map[string]string) {
	if maxParticipants < 2 || maxParticipants > 5 {
		fieldErrors["maxParticipants"] = "Max participants must be between 2 and 5."
	}
}

func validateSecretKey(secretKey string, fieldErrors map[string]string) {
	if secretKey == "" {
		fieldErrors["roomSecretKey"] = "Chat Room Secret Key is required."
		return
	}

	if strings.Contains(secretKey, " ") {
		fieldErrors["roomSecretKey"] = "Chat Room Secret Key cannot contain spaces."
		return
	}
}
