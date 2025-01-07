package handlers

import "strings"

// validateUsername checks if the provided username meets validation rules.
// Populates `fieldErrors` map with error messages if validation fails.
func validateUsername(username string, fieldErrors map[string]string) {
	if strings.Contains(username, " ") {
		fieldErrors["username"] = "Username cannot contain spaces"
		return
	}

	if len(username) < 3 || len(username) > 20 {
		fieldErrors["username"] = "Username must be between 3 and 20 characters"
		return
	}
}
