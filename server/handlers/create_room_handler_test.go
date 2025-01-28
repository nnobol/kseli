package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"kseli-server/auth"
	"kseli-server/config"
	"kseli-server/storage"
	"kseli-server/util"
)

// TestMain sets up shared configuration for all tests in this package.
func TestMain(m *testing.M) {
	config.GlobalConfig = &config.Config{
		SecretKey: "test-secret-key",
	}

	code := m.Run()

	os.Exit(code)
}

// TestCreateRoomHandler_Success validates the happy path of CreateRoomHandler.
func TestCreateRoomHandler_Success(t *testing.T) {
	requestBody := CreateRoomRequest{
		Username:        "JohnDoe",
		MaxParticipants: 3,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	memStore := storage.NewMemoryStore()

	handler := CreateRoomHandler(memStore)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	var resp CreateRoomResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if resp.RoomID == "" {
		t.Error("expected a non-empty room ID in response")
	}
	if resp.Token == "" {
		t.Error("expected a non-empty token in response")
	}
}

func TestCreateRoomHandler_DecodeFailure(t *testing.T) {
	// Invalid JSON payload
	invalidJSON := `{"username": "JohnDoe", "maxParticipants": 3,`

	req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	memStore := storage.NewMemoryStore()
	handler := CreateRoomHandler(memStore)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if resp.ErrorMessage != "failed to parse JSON: unexpected EOF" {
		t.Errorf("expected error message 'failed to parse JSON: unexpected EOF', got '%s'", resp.ErrorMessage)
	}
}

func TestCreateRoomHandler_GenerateRandomIDFailure(t *testing.T) {
	// Mock GenerateRandomIDFunc to return an error
	originalFunc := util.GenerateRandomIDFunc
	defer func() { util.GenerateRandomIDFunc = originalFunc }()

	util.GenerateRandomIDFunc = func() (string, error) {
		return "", errors.New("mock ID generation error")
	}

	requestBody := CreateRoomRequest{
		Username:        "JohnDoe",
		MaxParticipants: 3,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	memStore := storage.NewMemoryStore()
	handler := CreateRoomHandler(memStore)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if resp.ErrorMessage != "Failed to generate user ID" {
		t.Errorf("expected error message about ID generation failure, got '%s'", resp.ErrorMessage)
	}
}

func TestCreateRoomHandler_CreateTokenFailure(t *testing.T) {
	// Mock CreateTokenFunc to return an error
	originalFunc := auth.CreateTokenFunc
	defer func() { auth.CreateTokenFunc = originalFunc }()

	auth.CreateTokenFunc = func(claims auth.Claims, secretKey string) (string, error) {
		return "", errors.New("mock token generation error")
	}

	requestBody := CreateRoomRequest{
		Username:        "JohnDoe",
		MaxParticipants: 3,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	memStore := storage.NewMemoryStore()
	handler := CreateRoomHandler(memStore)
	handler.ServeHTTP(rr, req)

	// Assert HTTP status code
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Parse the error response
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert the error message
	if resp.ErrorMessage != "Failed to create token" {
		t.Errorf("expected error message about token creation failure, got '%s'", resp.ErrorMessage)
	}
}

// TestCreateRoomHandler_InvalidUsername validates Username validation errors
func TestCreateRoomHandler_InvalidUsername(t *testing.T) {
	testCases := []struct {
		name           string            // Test case name
		username       string            // Input username
		expectedStatus int               // Expected HTTP status
		expectedError  string            // Expected error message
		expectedFields map[string]string // Expected field-specific errors
	}{
		{
			name:           "Username contains spaces",
			username:       "John Doe",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation error",
			expectedFields: map[string]string{
				"username": "Username cannot contain spaces.",
			},
		},
		{
			name:           "Username is too short",
			username:       "JD",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation error",
			expectedFields: map[string]string{
				"username": "Username must be between 3 and 20 characters.",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request body
			requestBody := CreateRoomRequest{
				Username:        tc.username,
				MaxParticipants: 3,
			}
			bodyBytes, _ := json.Marshal(requestBody)

			// Prepare HTTP request and recorder
			req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			memStore := storage.NewMemoryStore()
			handler := CreateRoomHandler(memStore)
			handler.ServeHTTP(rr, req)

			// Assert HTTP status
			if rr.Code != tc.expectedStatus {
				t.Fatalf("Test [%s] failed: expected status %d, got %d", tc.name, tc.expectedStatus, rr.Code)
			}

			// Parse the error response
			var resp ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Test [%s] failed: failed to parse response JSON: %v", tc.name, err)
			}

			// Assert general error message
			if resp.ErrorMessage != tc.expectedError {
				t.Errorf("Test [%s] failed: expected error message '%s', got '%s'", tc.name, tc.expectedError, resp.ErrorMessage)
			}

			// Assert field-specific errors
			for field, expectedMsg := range tc.expectedFields {
				actualMsg, exists := resp.FieldErrors[field]
				if !exists {
					t.Errorf("Test [%s] failed: expected field error for '%s' but it was missing", tc.name, field)
					continue
				}
				if actualMsg != expectedMsg {
					t.Errorf("Test [%s] failed: expected error message for field '%s' to be '%s', got '%s'", tc.name, field, expectedMsg, actualMsg)
				}
			}

			// Ensure no unexpected field errors exist
			if len(resp.FieldErrors) != len(tc.expectedFields) {
				t.Errorf("Test [%s] failed: expected %d field errors, got %d", tc.name, len(tc.expectedFields), len(resp.FieldErrors))
			}
		})
	}
}

// TestCreateRoomHandler_InvalidMaxParticipants validates MaxParticipants validation errors.
func TestCreateRoomHandler_InvalidMaxParticipants(t *testing.T) {
	testCases := []struct {
		name            string
		maxParticipants int
		expectedStatus  int
		expectedError   string
		expectedFields  map[string]string
	}{
		{
			name:            "MaxParticipants is less than minimum (1)",
			maxParticipants: 1,
			expectedStatus:  http.StatusBadRequest,
			expectedError:   "Validation error",
			expectedFields: map[string]string{
				"maxParticipants": "Max participants must be between 2 and 5.",
			},
		},
		{
			name:            "MaxParticipants is greater than maximum (6)",
			maxParticipants: 6,
			expectedStatus:  http.StatusBadRequest,
			expectedError:   "Validation error",
			expectedFields: map[string]string{
				"maxParticipants": "Max participants must be between 2 and 5.",
			},
		},
		{
			name:            "MaxParticipants is negative",
			maxParticipants: -1,
			expectedStatus:  http.StatusBadRequest,
			expectedError:   "Validation error",
			expectedFields: map[string]string{
				"maxParticipants": "Max participants must be between 2 and 5.",
			},
		},
	}

	// Iterate over each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request body
			requestBody := CreateRoomRequest{
				Username:        "ValidUser",
				MaxParticipants: tc.maxParticipants,
			}
			bodyBytes, _ := json.Marshal(requestBody)

			// Prepare HTTP request and recorder
			req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			memStore := storage.NewMemoryStore()
			handler := CreateRoomHandler(memStore)
			handler.ServeHTTP(rr, req)

			// Assert HTTP status
			if rr.Code != tc.expectedStatus {
				t.Fatalf("Test [%s] failed: expected status %d, got %d", tc.name, tc.expectedStatus, rr.Code)
			}

			// Parse the error response
			var resp ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Test [%s] failed: failed to parse response JSON: %v", tc.name, err)
			}

			// Assert general error message
			if resp.ErrorMessage != tc.expectedError {
				t.Errorf("Test [%s] failed: expected error message '%s', got '%s'", tc.name, tc.expectedError, resp.ErrorMessage)
			}

			// Assert field-specific errors
			for field, expectedMsg := range tc.expectedFields {
				actualMsg, exists := resp.FieldErrors[field]
				if !exists {
					t.Errorf("Test [%s] failed: expected field error for '%s' but it was missing", tc.name, field)
					continue
				}
				if actualMsg != expectedMsg {
					t.Errorf("Test [%s] failed: expected error message for field '%s' to be '%s', got '%s'", tc.name, field, expectedMsg, actualMsg)
				}
			}

			// Ensure no unexpected field errors exist
			if len(resp.FieldErrors) != len(tc.expectedFields) {
				t.Errorf("Test [%s] failed: expected %d field errors, got %d", tc.name, len(tc.expectedFields), len(resp.FieldErrors))
			}
		})
	}
}
