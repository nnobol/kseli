package services

import (
	"context"
	"net/http"
	"strings"
	"time"

	"kseli-server/auth"
	"kseli-server/models"
	"kseli-server/models/api"
	"kseli-server/storage"

	"github.com/coder/websocket"
)

type RoomService struct {
	s *storage.MainStorage
}

func NewRoomService(s *storage.MainStorage) *RoomService {
	return &RoomService{
		s: s,
	}
}

func (rs *RoomService) CreateRoom(username, sessionID string, maxParticipants uint8) (*api.CreateRoomResponse, *api.ErrorResponse) {
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
		SessionId: sessionID,
		ID:        1,
		Username:  username,
		Role:      models.Admin,
	}

	roomID := rs.s.CreateRoom(adminUser, maxParticipants)

	claims := models.Claims{
		UserID:   adminUser.ID,
		Username: adminUser.Username,
		Role:     adminUser.Role,
		RoomID:   roomID,
		Exp:      time.Now().Add(time.Hour).Unix(),
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

func (rs *RoomService) JoinRoom(roomID, username, secretKey, sessionID string) (*api.JoinRoomResponse, *api.ErrorResponse) {
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

	room.Mu.RLock()
	if secretKey != *room.SecretKey {
		room.Mu.RUnlock()
		fieldErrors["roomSecretKey"] = "Incorrect Secret Key."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	if _, sessionAlreadyInRoom := room.Participants[sessionID]; sessionAlreadyInRoom {
		room.Mu.RUnlock()
		fieldErrors["roomId"] = "You can not join a room you are already in."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	// no need to lock for MaxParticipants, is an immutable field, won't be concurrently modified
	// need to think of how to optimize
	if uint8(len(room.Participants)) == room.MaxParticipants {
		room.Mu.RUnlock()
		fieldErrors["roomId"] = "Chat Room is full."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}

	if room.IsUsernameTaken(username) {
		room.Mu.RUnlock()
		fieldErrors["username"] = "This username is taken."
		return nil, &api.ErrorResponse{
			StatusCode:  http.StatusBadRequest,
			FieldErrors: fieldErrors,
		}
	}
	room.Mu.RUnlock()

	room.Mu.Lock()
	userId := room.NextUserId
	room.NextUserId++

	user := &models.User{
		SessionId: sessionID,
		ID:        userId,
		Username:  username,
		Role:      models.Member,
	}

	room.Join(user)
	room.Mu.Unlock()

	// edge case: room might get deleted by some go routine and panic on room.RoomID
	claims := models.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RoomID:   room.RoomID,
		Exp:      time.Now().Add(time.Hour).Unix(),
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

	// edge case: room might get deleted by some go routine and panic on room.RoomID
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

	// edge case: room might get deleted by some go routine and panic on room.MaxParticipants
	return &api.GetRoomResponse{
		UserRole:        userClaims.Role,
		MaxParticipants: room.MaxParticipants,
		Participants:    participants,
		SecretKey:       secretKey,
	}, nil
}

func (rs *RoomService) HandleWSConnection(ctx context.Context, roomID, username string, conn *websocket.Conn) {
	room, exists := rs.s.GetRoom(roomID)
	if !exists {
		conn.Write(ctx, websocket.MessageText, []byte(`{"error": "internal error - room not found"}`))
		conn.CloseNow()
		return
	}

	// edge case: room might get deleted by some go routine before AddWSConnection locks
	room.AddWSConnection(ctx, username, conn)
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
