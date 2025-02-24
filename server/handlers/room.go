package handlers

import (
	"encoding/json"
	"net/http"

	"kseli-server/handlers/utils"
	"kseli-server/models"
	"kseli-server/models/api"
	"kseli-server/services"
)

func CreateRoomHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.CreateRoomRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid JSON request body")
			return
		}

		sessionID, ok := r.Context().Value(models.UserSessionIDKey).(string)
		if !ok {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Session ID missing.")
			return
		}

		fingerprint, ok := r.Context().Value(models.UserFingerprintKey).(string)
		if !ok {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Fingerprint missing.")
			return
		}

		resp, errResp := rs.CreateRoom(req.Username, req.MaxParticipants, sessionID, fingerprint)

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
			utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid JSON request body")
			return
		}

		sessionID, ok := r.Context().Value(models.UserSessionIDKey).(string)
		if !ok {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Session ID missing.")
			return
		}

		fingerprint, ok := r.Context().Value(models.UserFingerprintKey).(string)
		if !ok {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Fingerprint missing.")
			return
		}

		resp, errResp := rs.JoinRoom(roomID, req.Username, req.RoomSecretKey, sessionID, fingerprint)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		utils.WriteSuccessResponse(w, http.StatusCreated, resp)
	}
}

func GetRoomHandler(rs *services.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		roomID := r.PathValue("roomID")

		userClaims, ok := r.Context().Value(models.UserClaimsKey).(*models.Claims)
		if !ok || userClaims == nil {
			utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		resp, errResp := rs.GetRoom(roomID, userClaims)

		if errResp != nil {
			utils.WriteErrorResponse(w, errResp)
			return
		}

		utils.WriteSuccessResponse(w, http.StatusCreated, resp)
	}
}
