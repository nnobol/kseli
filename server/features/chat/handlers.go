package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kseli-server/auth"
	"kseli-server/common"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func validateUsername(username string, fieldErrors map[string]string) {
	if username == "" {
		fieldErrors["username"] = "Username cannot be empty."
		return
	}

	if strings.Contains(username, " ") {
		fieldErrors["username"] = "Username cannot contain spaces."
		return
	}

	if len(username) < 3 || len(username) > 15 {
		fieldErrors["username"] = "Username must be between 3 and 15 characters."
	}
}

func validateMaxParticipants(maxParticipants uint8, fieldErrors map[string]string) {
	if maxParticipants < 2 || maxParticipants > 5 {
		fieldErrors["maxParticipants"] = "Max participants must be between 2 and 5."
	}
}

func validateRoomId(roomID string, fieldErrors map[string]string) {
	if roomID == "" {
		fieldErrors["roomId"] = "Chat Room Id is required."
		return
	}

	if strings.Contains(roomID, " ") {
		fieldErrors["roomId"] = "Chat Room Id cannot contain spaces."
		return
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

type CreateRoomRequest struct {
	Username        string `json:"username"`
	MaxParticipants uint8  `json:"maxParticipants"`
}

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
	Token  string `json:"token"`
}

func CreateRoomHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateRoomRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		sessionID, ok := r.Context().Value(auth.ParticipantSessionIDKey).(string)
		if !ok {
			common.WriteError(w, http.StatusUnauthorized, "Session ID missing.")
			return
		}

		fieldErrors := make(map[string]string, 2) // field name -> error message

		validateUsername(req.Username, fieldErrors)
		validateMaxParticipants(req.MaxParticipants, fieldErrors)

		if len(fieldErrors) > 0 {
			common.WriteFieldErrors(w, http.StatusBadRequest, fieldErrors)
			return
		}

		admin := &Participant{
			sessionID: sessionID,
			id:        1,
			username:  req.Username,
			role:      common.Admin,
		}

		roomID := generateUniqueRoomID(s)
		roomSecretKey := generateSecretKey()

		room := &Room{
			nextParticipantID:  2,
			roomID:             roomID,
			maxParticipants:    req.MaxParticipants,
			secretKey:          roomSecretKey,
			participants:       make(map[string]*Participant, req.MaxParticipants),
			bannedParticipants: make(map[string]struct{}),
			onClose:            func(roomID string) { s.DeleteRoom(roomID) },
			onExpire:           time.AfterFunc(30*time.Minute, func() { s.RoomCleanupFunc()(roomID) }),
			expiresAt:          time.Now().UTC().Add(30 * time.Minute).Unix(),
		}

		// no need to lock here since room is not in storage yet
		room.join(admin)

		s.AddRoom(roomID, room)

		claims := auth.Claims{
			UserID:   admin.id,
			Username: admin.username,
			Role:     admin.role,
			RoomID:   roomID,
			Exp:      time.Now().Add(time.Hour).Unix(),
		}

		token, err := auth.CreateToken(claims)
		if err != nil {
			room.participants = nil
			room.bannedParticipants = nil
			room.onClose = nil
			room.onExpire.Stop()
			room.onExpire = nil
			s.DeleteRoom(roomID)
			common.WriteError(w, http.StatusInternalServerError, "Failed to create token: "+err.Error())
			return
		}

		common.WriteJSON(w, http.StatusCreated, &CreateRoomResponse{
			RoomID: roomID,
			Token:  token,
		})
	}
}

type JoinRoomRequest struct {
	Username      string `json:"username"`
	RoomSecretKey string `json:"roomSecretKey"`
}

type JoinRoomResponse struct {
	Token string `json:"token"`
}

func JoinRoomHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req JoinRoomRequest

		roomID := r.PathValue("roomID")

		fieldErrors := make(map[string]string, 3) // field name -> error message

		validateRoomId(roomID, fieldErrors)

		if len(fieldErrors) > 0 {
			common.WriteFieldErrors(w, http.StatusBadRequest, fieldErrors)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		sessionID, ok := r.Context().Value(auth.ParticipantSessionIDKey).(string)
		if !ok {
			common.WriteError(w, http.StatusUnauthorized, "Session ID missing.")
			return
		}

		validateUsername(req.Username, fieldErrors)
		validateSecretKey(req.RoomSecretKey, fieldErrors)

		if len(fieldErrors) > 0 {
			common.WriteFieldErrors(w, http.StatusBadRequest, fieldErrors)
			return
		}

		room, exists := s.GetRoom(roomID)
		if !exists {
			fieldErrors["roomId"] = "Chat Room not found."
			common.WriteFieldErrors(w, http.StatusNotFound, fieldErrors)
			return
		}

		room.mu.RLock()
		if req.RoomSecretKey != room.secretKey {
			room.mu.RUnlock()
			fieldErrors["roomSecretKey"] = "Incorrect Secret Key."
			common.WriteFieldErrors(w, http.StatusForbidden, fieldErrors)
			return
		}

		if _, alreadyInRoom := room.participants[sessionID]; alreadyInRoom {
			room.mu.RUnlock()
			fieldErrors["roomId"] = "You can not join a room you are already in."
			common.WriteFieldErrors(w, http.StatusBadRequest, fieldErrors)
			return
		}

		if _, banned := room.bannedParticipants[sessionID]; banned {
			room.mu.RUnlock()
			fieldErrors["roomId"] = "You are banned from this room."
			common.WriteFieldErrors(w, http.StatusForbidden, fieldErrors)
			return
		}

		if uint8(len(room.participants)) == room.maxParticipants {
			room.mu.RUnlock()
			fieldErrors["roomId"] = "Chat Room is full."
			common.WriteFieldErrors(w, http.StatusConflict, fieldErrors)
			return
		}

		if room.isUsernameTaken(req.Username) {
			room.mu.RUnlock()
			fieldErrors["username"] = "This username is taken."
			common.WriteFieldErrors(w, http.StatusBadRequest, fieldErrors)
			return
		}
		room.mu.RUnlock()

		room.mu.Lock()
		pID := room.nextParticipantID
		room.nextParticipantID++

		p := &Participant{
			sessionID: sessionID,
			id:        pID,
			username:  req.Username,
			role:      common.Member,
		}

		room.join(p)

		roomIDCopy := room.roomID

		room.mu.Unlock()

		claims := auth.Claims{
			UserID:   p.id,
			Username: p.username,
			Role:     p.role,
			RoomID:   roomIDCopy,
			Exp:      time.Now().Add(time.Hour).Unix(),
		}

		token, err := auth.CreateToken(claims)
		if err != nil {
			room.mu.Lock()
			if p.wsTimeout != nil {
				p.wsTimeout.Stop()
				p.wsTimeout = nil
			}
			delete(room.participants, sessionID)
			room.mu.Unlock()
			common.WriteError(w, http.StatusInternalServerError, "Failed to create token: "+err.Error())
			return
		}

		common.WriteJSON(w, http.StatusCreated, &JoinRoomResponse{
			Token: token,
		})
	}
}

type GetRoomResponse struct {
	UserRole        common.Role       `json:"userRole"`
	MaxParticipants uint8             `json:"maxParticipants"`
	Participants    []ParticipantView `json:"participants"`
	ExpiresAt       int64             `json:"expiresAt"`
	RoomID          string            `json:"roomId"`
	SecretKey       string            `json:"secretKey,omitempty"`
}

func GetRoomHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomID := r.PathValue("roomID")

		if roomID == "" {
			common.WriteError(w, http.StatusBadRequest, "Chat Room Id is required.")
			return
		}

		if strings.Contains(roomID, " ") {
			common.WriteError(w, http.StatusBadRequest, "Chat Room Id cannot contain spaces.")
			return
		}

		claims, ok := r.Context().Value(auth.ParticipantClaimsKey).(*auth.Claims)
		if !ok || claims == nil {
			common.WriteError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		room, exists := s.GetRoom(roomID)
		if !exists {
			common.WriteError(w, http.StatusNotFound, "Chat Room not found.")
			return
		}

		room.mu.RLock()
		if claims.RoomID != room.roomID {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusForbidden, "You do not have access to this room.")
			return
		}

		secretKey := ""
		if claims.Role == common.Admin {
			secretKey = room.secretKey
		}

		participants := room.getParticipantsAsSlice()

		resp := &GetRoomResponse{
			UserRole:        claims.Role,
			MaxParticipants: room.maxParticipants,
			Participants:    participants,
			ExpiresAt:       room.expiresAt,
			RoomID:          room.roomID,
			SecretKey:       secretKey,
		}
		room.mu.RUnlock()

		common.WriteJSON(w, http.StatusCreated, resp)
	}
}

func DeleteRoomHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomID := r.PathValue("roomID")

		if roomID == "" {
			common.WriteError(w, http.StatusBadRequest, "Chat Room Id is required.")
			return
		}

		if strings.Contains(roomID, " ") {
			common.WriteError(w, http.StatusBadRequest, "Chat Room Id cannot contain spaces.")
			return
		}

		claims, ok := r.Context().Value(auth.ParticipantClaimsKey).(*auth.Claims)
		if !ok || claims == nil {
			common.WriteError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		room, exists := s.GetRoom(roomID)
		if !exists {
			common.WriteError(w, http.StatusNotFound, "Chat Room not found.")
			return
		}

		room.mu.RLock()
		if claims.RoomID != room.roomID {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusForbidden, "You do not have access to this room.")
			return
		}
		room.mu.RUnlock()

		if claims.Role != common.Admin {
			common.WriteError(w, http.StatusForbidden, "You are not an admin and can't close this room.")
			return
		}

		room.Close(false)

		w.WriteHeader(http.StatusNoContent)
	}
}

type UserRequest struct {
	TargetUserID uint8 `json:"userId"`
}

func performRoomAction(s Storage, action string, actionFunc func(r *Room, targetID uint8) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UserRequest

		roomID := r.PathValue("roomID")

		if roomID == "" {
			common.WriteError(w, http.StatusBadRequest, "Chat Room Id is required.")
			return
		}

		if strings.Contains(roomID, " ") {
			common.WriteError(w, http.StatusBadRequest, "Chat Room Id cannot contain spaces.")
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		claims, ok := r.Context().Value(auth.ParticipantClaimsKey).(*auth.Claims)
		if !ok || claims == nil {
			common.WriteError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		room, exists := s.GetRoom(roomID)
		if !exists {
			common.WriteError(w, http.StatusNotFound, "Chat Room not found.")
			return
		}

		room.mu.RLock()
		if claims.RoomID != room.roomID {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusForbidden, "You do not have access to this room.")
			return
		}
		room.mu.RUnlock()

		if claims.Role != common.Admin {
			common.WriteError(w, http.StatusForbidden, fmt.Sprintf("You are not an admin and can't %s anyone from this room.", action))
			return
		}

		if req.TargetUserID == 0 {
			common.WriteError(w, http.StatusBadRequest, "User Id not sent in the request.")
			return
		}

		if err := actionFunc(room, req.TargetUserID); err != nil {
			common.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func KickParticipantHandler(s Storage) http.HandlerFunc {
	return performRoomAction(s, "kick", func(r *Room, id uint8) error {
		return r.kick(id)
	})
}

func BanParticipantHandler(s Storage) http.HandlerFunc {
	return performRoomAction(s, "ban", func(r *Room, id uint8) error {
		return r.ban(id)
	})
}

func RoomWSHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.HTTPUpgrader{
			Timeout: 2 * time.Second,
		}.Upgrade(r, w)
		if err != nil {
			return
		}

		token := r.URL.Query().Get("token")
		if token == "" {
			wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "token-missing"))
			conn.Close()
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "token-invalid"))
			conn.Close()
			return
		}

		room, exists := s.GetRoom(claims.RoomID)
		if !exists {
			wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "room-not-exists"))
			conn.Close()
			return
		}

		room.addWSConn(conn, claims.Username)
	}
}
