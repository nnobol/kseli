package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kseli/auth"
	"kseli/common"
	"kseli/middleware"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

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
		r.Body = http.MaxBytesReader(w, r.Body, 128)

		var req CreateRoomRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		fieldErrors := make(map[string]string, 2) // field name -> error message

		validateUsername(req.Username, fieldErrors)
		if req.MaxParticipants < 2 || req.MaxParticipants > 5 {
			fieldErrors["maxParticipants"] = "Max participants must be between 2 and 5."
		}

		if len(fieldErrors) > 0 {
			common.WriteFieldErrors(w, http.StatusBadRequest, fieldErrors)
			return
		}

		roomID := generateUniqueRoomID(s)
		roomSecretKey := generateRandomString(10)
		roomExpiration := time.Now().Add(30 * time.Minute).Unix()

		inviteClaims := auth.InviteClaims{
			RoomID:    roomID,
			SecretKey: roomSecretKey,
			Exp:       roomExpiration,
		}

		inviteToken, err := auth.CreateToken(inviteClaims)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, "Failed to create invite token: "+err.Error())
			return
		}
		origin := r.Header.Get("Origin")
		inviteLink := fmt.Sprintf("%s/join#invite=%s", origin, inviteToken)

		room := &Room{
			nextParticipantID:  2,
			maxParticipants:    req.MaxParticipants,
			roomID:             roomID,
			secretKey:          roomSecretKey,
			inviteLink:         inviteLink,
			participants:       make(map[string]*Participant, req.MaxParticipants),
			bannedParticipants: make(map[string]struct{}),
			onClose:            func(roomID string) { s.DeleteRoom(roomID) },
			onExpire:           time.AfterFunc(30*time.Minute, func() { s.RoomCleanupFunc()(roomID) }),
			expiresAt:          roomExpiration,
		}

		sessionID, ok := r.Context().Value(auth.ParticipantSessionIDKey).(string)
		if !ok {
			common.WriteError(w, http.StatusInternalServerError, "Invalid Session ID.")
			return
		}

		admin := &Participant{
			sessionID: sessionID,
			id:        1,
			username:  req.Username,
			role:      common.Admin,
		}

		// no need to lock here since room is not in storage yet
		room.join(admin)

		s.AddRoom(roomID, room)

		claims := auth.Claims{
			UserID:   admin.id,
			Username: admin.username,
			Role:     admin.role,
			RoomID:   roomID,
			Exp:      roomExpiration,
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
	Username string `json:"username"`
}

type JoinRoomResponse struct {
	RoomID string `json:"roomId"`
	Token  string `json:"token"`
}

func JoinRoomHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 128)

		var req JoinRoomRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		fieldErrors := make(map[string]string, 1) // field name -> error message

		validateUsername(req.Username, fieldErrors)

		if len(fieldErrors) > 0 {
			common.WriteFieldErrors(w, http.StatusBadRequest, fieldErrors)
			return
		}

		sessionID, ok := r.Context().Value(auth.ParticipantSessionIDKey).(string)
		if !ok {
			common.WriteError(w, http.StatusInternalServerError, "Invalid Session ID.")
			return
		}

		inviteClaims, ok := r.Context().Value(auth.InviteClaimsKey).(*auth.InviteClaims)
		if !ok {
			common.WriteError(w, http.StatusInternalServerError, "Invalid invite token.")
			return
		}

		room, exists := s.GetRoom(inviteClaims.RoomID)
		if !exists {
			common.WriteError(w, http.StatusNotFound, "Chat Room not found.")
			return
		}

		room.mu.RLock()
		if inviteClaims.SecretKey != room.secretKey {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusForbidden, "Invalid invite link.")
			return
		}

		if _, alreadyInRoom := room.participants[sessionID]; alreadyInRoom {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusBadRequest, "You can not join a room you are already in.")
			return
		}

		if _, banned := room.bannedParticipants[sessionID]; banned {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusForbidden, "You are banned from this room.")
			return
		}

		if uint8(len(room.participants)) == room.maxParticipants {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusConflict, "Chat Room is full.")
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
		room.mu.Unlock()

		claims := auth.Claims{
			UserID:   p.id,
			Username: p.username,
			Role:     p.role,
			RoomID:   inviteClaims.RoomID,
			Exp:      time.Now().Add(30 * time.Minute).Unix(),
		}

		token, err := auth.CreateToken(claims)
		if err != nil {
			room.mu.Lock()
			// if p.wsTimeout != nil {
			// 	p.wsTimeout.Stop()
			// 	p.wsTimeout = nil
			// }
			delete(room.participants, sessionID)
			room.mu.Unlock()
			common.WriteError(w, http.StatusInternalServerError, "Failed to create token: "+err.Error())
			return
		}

		common.WriteJSON(w, http.StatusCreated, &JoinRoomResponse{
			RoomID: inviteClaims.RoomID,
			Token:  token,
		})
	}
}

type GetRoomResponse struct {
	UserRole        common.Role       `json:"userRole"`
	MaxParticipants uint8             `json:"maxParticipants"`
	Participants    []ParticipantView `json:"participants"`
	ExpiresAt       int64             `json:"expiresAt"`
	InviteLink      string            `json:"inviteLink,omitempty"`
}

func GetRoomHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomID := r.PathValue("roomID")

		if err := validateRoomId(roomID); err != "" {
			common.WriteError(w, http.StatusBadRequest, err)
			return
		}

		room, exists := s.GetRoom(roomID)
		if !exists {
			common.WriteError(w, http.StatusNotFound, "Chat Room not found.")
			return
		}

		claims, ok := r.Context().Value(auth.ParticipantClaimsKey).(*auth.Claims)
		if !ok || claims == nil {
			common.WriteError(w, http.StatusInternalServerError, "Invalid authorizaton token.")
			return
		}

		room.mu.RLock()
		if claims.RoomID != room.roomID {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusForbidden, "You do not have access to this room.")
			return
		}

		_, exists = room.getParticipantByID(claims.UserID)
		if !exists {
			room.mu.RUnlock()
			common.WriteError(w, http.StatusForbidden, "You are not in this room and can't retrieve the details. Try joining again.")
			return
		}

		inviteLink := ""
		if claims.Role == common.Admin {
			inviteLink = room.inviteLink
		}

		participants := room.getParticipantsAsSlice()

		resp := &GetRoomResponse{
			UserRole:        claims.Role,
			MaxParticipants: room.maxParticipants,
			Participants:    participants,
			ExpiresAt:       room.expiresAt,
			InviteLink:      inviteLink,
		}
		room.mu.RUnlock()

		common.WriteJSON(w, http.StatusOK, resp)
	}
}

func DeleteRoomHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomID := r.PathValue("roomID")

		if err := validateRoomId(roomID); err != "" {
			common.WriteError(w, http.StatusBadRequest, err)
			return
		}

		room, exists := s.GetRoom(roomID)
		if !exists {
			common.WriteError(w, http.StatusNotFound, "Chat Room not found.")
			return
		}

		claims, ok := r.Context().Value(auth.ParticipantClaimsKey).(*auth.Claims)
		if !ok || claims == nil {
			common.WriteError(w, http.StatusInternalServerError, "Invalid authorizaton token.")
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
		r.Body = http.MaxBytesReader(w, r.Body, 128)

		var req UserRequest

		roomID := r.PathValue("roomID")

		if err := validateRoomId(roomID); err != "" {
			common.WriteError(w, http.StatusBadRequest, err)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		if req.TargetUserID == 0 {
			common.WriteError(w, http.StatusBadRequest, "User Id is required in the request.")
			return
		}

		room, exists := s.GetRoom(roomID)
		if !exists {
			common.WriteError(w, http.StatusNotFound, "Chat Room not found.")
			return
		}

		claims, ok := r.Context().Value(auth.ParticipantClaimsKey).(*auth.Claims)
		if !ok || claims == nil {
			common.WriteError(w, http.StatusInternalServerError, "Invalid authorizaton token.")
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

		if claims.UserID == req.TargetUserID {
			common.WriteError(w, http.StatusBadRequest, fmt.Sprintf("You can't %s yourself from the room.", action))
			return
		}

		if err := actionFunc(room, req.TargetUserID); err != nil {
			common.WriteError(w, http.StatusNotFound, err.Error())
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

var WaitForTestClientWSSetup chan struct{}

func waitForTestClientWS() {
	if WaitForTestClientWSSetup != nil {
		<-WaitForTestClientWSSetup
		time.Sleep(1 * time.Millisecond)
	}
}

func RoomWSHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		_, errMsg := middleware.ValidateOriginHost(origin)

		conn, _, _, err := ws.HTTPUpgrader{
			Timeout: 2 * time.Second,
		}.Upgrade(r, w)
		if err != nil {
			return
		}

		waitForTestClientWS()

		if errMsg != "" {
			wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "invalid-origin"))
			conn.Close()
			return
		}

		token := r.URL.Query().Get("token")
		if token == "" {
			wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "token-missing"))
			conn.Close()
			return
		}

		claims, err := auth.ValidateToken[auth.Claims](token)
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

type MetricsMessage struct {
	RoomCount        int `json:"roomCount"`
	ParticipantCount int `json:"participantCount"`
}

func RoomMetricsWSHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		_, errMsg := middleware.ValidateOriginHost(origin)

		conn, _, _, err := ws.HTTPUpgrader{
			Timeout: 2 * time.Second,
		}.Upgrade(r, w)
		if err != nil {
			return
		}

		if errMsg != "" {
			wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "invalid-origin"))
			conn.Close()
			return
		}

		go func() {
			defer conn.Close()

			roomCount, participantCount := s.MetricsSnapshot()
			initialMsg := MetricsMessage{
				RoomCount:        roomCount,
				ParticipantCount: participantCount,
			}
			bytes, err := json.Marshal(initialMsg)
			if err == nil {
				wsutil.WriteServerMessage(conn, ws.OpText, bytes)
			}

			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				roomCount, participantCount := s.MetricsSnapshot()

				msg := MetricsMessage{
					RoomCount:        roomCount,
					ParticipantCount: participantCount,
				}

				data, err := json.Marshal(msg)
				if err != nil {
					return
				}

				if err := wsutil.WriteServerMessage(conn, ws.OpText, data); err != nil {
					return
				}
			}
		}()
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

	runeCount := len([]rune(username))
	if runeCount < 3 || runeCount > 15 {
		fieldErrors["username"] = "Username must be between 3 and 15 characters."
	}
}

func validateRoomId(roomID string) string {
	if strings.Contains(roomID, " ") {
		return "Chat Room Id cannot contain spaces."
	}

	if len(roomID) > 16 {
		return "Incorrect Chat Room Id, it is too long."
	}

	return ""
}
