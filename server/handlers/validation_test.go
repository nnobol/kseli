package handlers

import "testing"

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
			expectedError: "Username cannot contain spaces",
		},
		{
			name:          "Username is too short",
			username:      "AB",
			expectedError: "Username must be between 3 and 20 characters",
		},
		{
			name:          "Username is too long",
			username:      "ThisUsernameIsWayTooLongToBeValid",
			expectedError: "Username must be between 3 and 20 characters",
		},
		{
			name:          "Empty username",
			username:      "",
			expectedError: "Username must be between 3 and 20 characters",
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
