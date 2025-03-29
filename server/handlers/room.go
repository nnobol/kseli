package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kseli-server/auth"
	"kseli-server/handlers/utils"
	"kseli-server/models"
	"kseli-server/models/api"
	"kseli-server/services"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func CreateRoomHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.CreateRoomRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		sessionID, ok := r.Context().Value(models.UserSessionIDKey).(string)
		if !ok {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Session ID missing.")
			return
		}

		resp, errResp := rs.CreateRoom(req.Username, sessionID, req.MaxParticipants)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		utils.WriteSuccessResponse(w, http.StatusCreated, resp)
	}
}

func JoinRoomHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.JoinRoomRequest

		roomID := r.PathValue("roomID")

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		sessionID, ok := r.Context().Value(models.UserSessionIDKey).(string)
		if !ok {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Session ID missing.")
			return
		}

		resp, errResp := rs.JoinRoom(roomID, req.Username, req.RoomSecretKey, sessionID)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		utils.WriteSuccessResponse(w, http.StatusCreated, resp)
	}
}

func KickUserHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.UserRequest

		roomID := r.PathValue("roomID")

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		userClaims, ok := r.Context().Value(models.UserClaimsKey).(*models.Claims)
		if !ok || userClaims == nil {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		errResp := rs.KickUser(roomID, req.TargetUserID, userClaims)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func BanUserHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.UserRequest

		roomID := r.PathValue("roomID")

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid JSON request body.")
			return
		}

		userClaims, ok := r.Context().Value(models.UserClaimsKey).(*models.Claims)
		if !ok || userClaims == nil {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		errResp := rs.BanUser(roomID, req.TargetUserID, userClaims)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func GetRoomHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomID := r.PathValue("roomID")

		userClaims, ok := r.Context().Value(models.UserClaimsKey).(*models.Claims)
		if !ok || userClaims == nil {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		resp, errResp := rs.GetRoom(roomID, userClaims)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		utils.WriteSuccessResponse(w, http.StatusOK, resp)
	}
}

func DeleteRoomHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomID := r.PathValue("roomID")

		userClaims, ok := r.Context().Value(models.UserClaimsKey).(*models.Claims)
		if !ok || userClaims == nil {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		errResp := rs.DeleteRoom(roomID, userClaims)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func RoomWebSocketHandler(rs *services.RoomService) http.HandlerFunc {
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

		rs.HandleRoomWSConnection(conn, claims.RoomID, claims.Username)
	}
}
