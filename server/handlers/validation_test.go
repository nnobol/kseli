package handlers

import (
	"testing"

	"kseli-server/models"
)

func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		name          string
		username      string
		expectedError string
	}{
		{
			name:          "Valid username",
			username:      "ValidUser",
			expectedError: "",
		},
		{
			name:          "Username contains spaces",
			username:      "Invalid Username",
			expectedError: "Username cannot contain spaces.",
		},
		{
			name:          "Username is too short",
			username:      "AB",
			expectedError: "Username must be between 3 and 20 characters.",
		},
		{
			name:          "Username is too long",
			username:      "ThisUsernameIsWayTooLongToBeValid",
			expectedError: "Username must be between 3 and 20 characters.",
		},
		{
			name:          "Empty username",
			username:      "",
			expectedError: "Username must be between 3 and 20 characters.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldErrors := make(map[string]string)
			validateUsername(tc.username, fieldErrors)

			actualError, exists := fieldErrors["username"]
			if tc.expectedError == "" && exists {
				t.Errorf("Test [%s] failed: expected no error, but got '%s'", tc.name, actualError)
			} else if tc.expectedError != "" && (!exists || actualError != tc.expectedError) {
				t.Errorf("Test [%s] failed: expected error '%s', but got '%s'", tc.name, tc.expectedError, actualError)
			}
		})
	}
}

func TestValidateUsernameNotTaken(t *testing.T) {
	testCases := []struct {
		name          string
		username      string
		participants  map[string]*models.User
		existingError string
		expectedError string
	}{
		{
			name:          "Username already in room",
			username:      "John",
			participants:  map[string]*models.User{"John": {Username: "John"}},
			existingError: "",
			expectedError: "This username is taken.",
		},
		{
			name:          "Username not taken",
			username:      "John",
			participants:  map[string]*models.User{"Alice": {Username: "Alice"}},
			existingError: "",
			expectedError: "",
		},
		{
			name:          "Another field error already set for username",
			username:      "John",
			participants:  map[string]*models.User{"Alice": {Username: "Alice"}},
			existingError: "Username cannot contain spaces",
			expectedError: "Username cannot contain spaces",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldErrors := make(map[string]string)
			if tc.existingError != "" {
				fieldErrors["username"] = tc.existingError
			}

			validateUsernameNotTaken(tc.username, tc.participants, fieldErrors)

			actualError, exists := fieldErrors["username"]

			if tc.expectedError == "" && exists {
				t.Errorf("Test [%s] failed: expected no error, but got '%s'", tc.name, actualError)
			} else if tc.expectedError != "" {
				// If we expect an error, it must exist and match
				if !exists {
					t.Errorf("Test [%s] failed: expected error '%s', but got no error", tc.name, tc.expectedError)
				} else if actualError != tc.expectedError {
					t.Errorf("Test [%s] failed: expected error '%s', but got '%s'", tc.name, tc.expectedError, actualError)
				}
			}
		})
	}
}

func TestValidateRoomId(t *testing.T) {
	testCases := []struct {
		name          string
		roomID        string
		expectedError string
	}{
		{
			name:          "Room ID has spaces",
			roomID:        "abc 123",
			expectedError: "Chat Room Id cannot contain spaces.",
		},
		{
			name:          "Valid room ID",
			roomID:        "abc123",
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldErrors := make(map[string]string)

			validateRoomId(tc.roomID, fieldErrors)

			actualError, exists := fieldErrors["roomId"]
			if tc.expectedError == "" && exists {
				t.Errorf("Test [%s] failed: expected no error, but got '%s'", tc.name, actualError)
			} else if tc.expectedError != "" {
				if !exists {
					t.Errorf("Test [%s] failed: expected error '%s', but got no error", tc.name, tc.expectedError)
				} else if actualError != tc.expectedError {
					t.Errorf("Test [%s] failed: expected error '%s', but got '%s'", tc.name, tc.expectedError, actualError)
				}
			}
		})
	}
}

func TestValidateRoomSecretKey(t *testing.T) {
	testCases := []struct {
		name          string
		roomSecretKey string
		expectedError string
	}{
		{
			name:          "Empty secret key",
			roomSecretKey: "",
			expectedError: "Chat Room Secret Key is required.",
		},
		{
			name:          "Secret key has spaces",
			roomSecretKey: "secret key",
			expectedError: "Chat Room Secret Key cannot contain spaces.",
		},
		{
			name:          "Valid secret key",
			roomSecretKey: "Secret123",
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldErrors := make(map[string]string)

			validateRoomSecretKey(tc.roomSecretKey, fieldErrors)

			actualError, exists := fieldErrors["roomSecretKey"]
			if tc.expectedError == "" && exists {
				t.Errorf("Test [%s] failed: expected no error, but got '%s'", tc.name, actualError)
			} else if tc.expectedError != "" {
				if !exists {
					t.Errorf("Test [%s] failed: expected error '%s', but got no error", tc.name, tc.expectedError)
				} else if actualError != tc.expectedError {
					t.Errorf("Test [%s] failed: expected error '%s', but got '%s'", tc.name, tc.expectedError, actualError)
				}
			}
		})
	}
}
