package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kseli-server/auth"
	"kseli-server/config"
	"kseli-server/models"
	"kseli-server/storage"
	"kseli-server/util"
)

func JoinRoomHandler(storage *storage.RoomStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fieldErrors := make(map[string]string)

		roomID := r.PathValue("roomID")

		validateRoomId(roomID, fieldErrors)
		// If the roomID is invalid, no point in reading body or looking up the room in memory.
		if len(fieldErrors) > 0 {
			WriteJSONError(w, http.StatusBadRequest, "", fieldErrors)
			return
		}

		var req JoinRoomRequest
		if err := StrictDecodeJSON(r, &req); err != nil {
			WriteJSONError(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		validateUsername(req.Username, fieldErrors)
		validateRoomSecretKey(req.RoomSecretKey, fieldErrors)
		// If the username and secretKey fields are invalid, no point looking up the room in memory
		if len(fieldErrors) > 0 {
			WriteJSONError(w, http.StatusBadRequest, "", fieldErrors)
			return
		}

		room, err := storage.GetRoom(roomID)
		if err != nil {
			fieldErrors["roomId"] = "Chat Room not found."
			WriteJSONError(w, http.StatusNotFound, "", fieldErrors)
			return
		}

		if req.RoomSecretKey != room.SecretKey {
			fieldErrors["roomSecretKey"] = "Incorrect Secret Key."
			WriteJSONError(w, http.StatusBadRequest, "", fieldErrors)
			return
		}

		if len(room.Participants) == room.MaxParticipants {
			fieldErrors["roomId"] = "Chat Room is full."
			WriteJSONError(w, http.StatusBadRequest, "", fieldErrors)
			return
		}

		validateUsernameNotTaken(req.Username, room.Participants, fieldErrors)
		if len(fieldErrors) > 0 {
			WriteJSONError(w, http.StatusBadRequest, "", fieldErrors)
			return
		}

		userID, err := util.GenerateRandomIDFunc()
		if err != nil {
			WriteJSONError(w, http.StatusInternalServerError, "Failed to generate user ID", nil)
			return
		}

		// Create the user
		user := models.User{
			ID:       userID,
			Username: req.Username,
			Role:     models.Member,
		}

		// Join the room
		room.JoinRoom(&user)

		// Build JWT claims
		claims := auth.Claims{
			UserID: userID,
			RoomID: roomID,
			Role:   models.Member,
			Exp:    time.Now().Add(time.Hour).Unix(),
		}

		// Get the SECRET_KEY from the global config
		secretKey := config.GlobalConfig.SecretKey

		// Generate the token
		token, err := auth.CreateTokenFunc(claims, secretKey)
		if err != nil {
			WriteJSONError(w, http.StatusInternalServerError, "Failed to create token", nil)
			return
		}

		joinRoomResponse := JoinRoomResponse{
			RoomID: room.ID,
			Token:  token,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(joinRoomResponse)
	}
}
