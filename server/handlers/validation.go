package handlers

import (
	"strings"

	"kseli-server/models"
)

// validateUsername checks if the provided username meets validation rules.
// Populates `fieldErrors` map with error messages if validation fails.
func validateUsername(username string, fieldErrors map[string]string) {
	if strings.Contains(username, " ") {
		fieldErrors["username"] = "Username cannot contain spaces."
		return
	}

	if len(username) < 3 || len(username) > 20 {
		fieldErrors["username"] = "Username must be between 3 and 20 characters."
		return
	}
}

func validateUsernameNotTaken(username string, participants map[string]*models.User, fieldErrors map[string]string) {
	// If there's already an error on 'username', we skip
	if _, exists := fieldErrors["username"]; exists {
		return
	}

	if _, exists := participants[username]; exists {
		fieldErrors["username"] = "This username is taken."
		return
	}
}

func validateRoomId(roomID string, fieldErrors map[string]string) {
	if strings.Contains(roomID, " ") {
		fieldErrors["roomId"] = "Chat Room Id cannot contain spaces."
		return
	}
}

func validateRoomSecretKey(roomSecretKey string, fieldErrors map[string]string) {
	if roomSecretKey == "" {
		fieldErrors["roomSecretKey"] = "Chat Room Secret Key is required."
		return
	}

	if strings.Contains(roomSecretKey, " ") {
		fieldErrors["roomSecretKey"] = "Chat Room Secret Key cannot contain spaces."
		return
	}
}
