package handlers

import (
	"encoding/json"
	"net/http"

	"kseli-server/contextutil"
	"kseli-server/models"
	"kseli-server/storage"
)

func GetRoomHandler(storage *storage.RoomStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		claims, ok := contextutil.GetUserClaimsFromContext(r.Context())
		if !ok {
			WriteJSONError(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}

		fieldErrors := make(map[string]string)

		roomID := r.PathValue("roomID")
		validateRoomId(roomID, fieldErrors)
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

		if claims.RoomID != room.ID {
			WriteJSONError(w, http.StatusForbidden, "You do not have access to this room.", nil)
			return
		}

		room.Mu.RLock()
		participantsArray := make([]models.User, 0, len(room.Participants))
		for _, user := range room.Participants {
			participantsArray = append(participantsArray, *user)
		}
		room.Mu.RUnlock()

		// Only admins get to see the secret key
		response := GetRoomResponse{
			RoomID:          room.ID,
			MaxParticipants: room.MaxParticipants,
			Participants:    participantsArray,
		}

		if claims.Role == models.Admin {
			response.SecretKey = room.SecretKey
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
