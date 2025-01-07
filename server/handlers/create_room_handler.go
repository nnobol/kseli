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

func CreateRoomHandler(storage *storage.RoomStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req CreateRoomRequest
		if err := StrictDecodeJSON(r, &req); err != nil {
			WriteJSONError(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		// Validate inputs
		fieldErrors := make(map[string]string)
		validateUsername(req.Username, fieldErrors)
		if req.MaxParticipants < 2 || req.MaxParticipants > 5 {
			fieldErrors["maxParticipants"] = "Max participants must be between 2 and 5"
		}
		if len(fieldErrors) > 0 {
			WriteJSONError(w, http.StatusBadRequest, "Validation error", fieldErrors)
			return
		}

		// Generate a random admin user ID
		adminUserID, err := util.GenerateRandomIDFunc()
		if err != nil {
			WriteJSONError(w, http.StatusInternalServerError, "Failed to generate user ID", nil)
			return
		}

		// Create the admin user
		adminUser := models.User{
			ID:       adminUserID,
			Username: req.Username,
			Role:     models.Admin,
		}

		// Create the room and store it in memory
		roomID := storage.CreateRoom(&adminUser, req.MaxParticipants)

		// Build JWT claims
		claims := auth.Claims{
			UserID: adminUserID,
			RoomID: roomID,
			Role:   models.Admin,
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

		createRoomResponse := CreateRoomResponse{
			RoomID: roomID,
			Token:  token,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(createRoomResponse)
	}
}
