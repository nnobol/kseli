package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"kseli-server/auth"
	"kseli-server/models"
	"kseli-server/storage"
	"kseli-server/util"
)

func createTestRoom(store *storage.RoomStorage, secretKey string, maxParticipants int) string {
	// For simplicity, call the store's creation
	admin := &models.User{ID: "AdminID", Username: "AdminUser", Role: models.Admin}
	roomID := store.CreateRoom(admin, maxParticipants)

	// Then override secretKey if needed.
	// Because store.mu and store.rooms are unexported,
	// we can define a small helper in RoomStorage to do it or just do store.GetRoom
	room, _ := store.GetRoom(roomID)
	// Now we have the pointer
	room.SecretKey = secretKey

	// If forcedID is non-empty, you'd do another override, but that requires messing with store.mu again
	return roomID
}

// TestJoinRoomHandler_Success tests the happy path of joining a room.
func TestJoinRoomHandler_Success(t *testing.T) {
	// 1. Create a room in the store
	memStore := storage.NewMemoryStore()
	roomID := createTestRoom(memStore, "secret123", 3)

	mux := http.NewServeMux()
	// if on 1.22, do:
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))

	// 2. Prepare valid request body
	joinRequest := JoinRoomRequest{
		Username:      "NewUser",
		RoomSecretKey: "secret123",
	}
	bodyBytes, _ := json.Marshal(joinRequest)

	// 3. Build HTTP request/recorder
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+roomID+"/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// 4. Serve the request
	mux.ServeHTTP(rr, req)

	// 5. Assertions
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	// Parse the joinRoomResponse
	var resp JoinRoomResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if resp.Token == "" {
		t.Errorf("expected a non-empty token in response")
	}

	// Ensure user actually joined the room in memory
	room, _ := memStore.GetRoom(roomID)
	if _, exists := room.Participants["NewUser"]; !exists {
		t.Errorf("expected 'NewUser' to be in room participants, but not found")
	}
}

func TestJoinRoomHandler_RoomIDInvalid(t *testing.T) {
	memStore := storage.NewMemoryStore()

	// Table of test cases
	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty room ID",
			path:           "/api/rooms//users",
			expectedStatus: http.StatusMovedPermanently, // Expect 301 due to ServeMux sanitization
		},
		{
			name:           "Room ID with spaces (encoded as %20)",
			path:           "/api/rooms/room%20id/users",
			expectedStatus: http.StatusBadRequest, // Expect 400 due to validation
			expectedError:  "Chat Room Id cannot contain spaces.",
		},
	}

	// Set up the mux
	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, nil)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			// Assert the status code
			if rr.Code != tc.expectedStatus {
				t.Fatalf("Test [%s] failed: expected status=%d, got=%d", tc.name, tc.expectedStatus, rr.Code)
			}

			// If it's a bad request, validate the error message
			if tc.expectedStatus == http.StatusBadRequest {
				var resp ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Test [%s] failed: failed to parse error response JSON: %v", tc.name, err)
				}
				if resp.FieldErrors["roomId"] != tc.expectedError {
					t.Errorf("Test [%s] failed: expected error='%s', got='%s'", tc.name, tc.expectedError, resp.FieldErrors["roomId"])
				}
			}

			// If it's a redirect, validate the `Location` header
			if tc.expectedStatus == http.StatusMovedPermanently {
				location := rr.Header().Get("Location")
				if location == "" {
					t.Errorf("Test [%s] failed: expected Location header but it was missing", tc.name)
				}
			}
		})
	}
}

func TestJoinRoomHandler_DecodeFailure(t *testing.T) {
	memStore := storage.NewMemoryStore()

	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))

	invalidJSON := `{"username": "SomeUser", "roomSecretKey": 5,`
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/validRoomID/users", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}

	if resp.ErrorMessage != "failed to parse JSON: unexpected EOF" {
		t.Errorf("expected error message 'failed to parse JSON: unexpected EOF', got '%s'", resp.ErrorMessage)
	}
}

func TestJoinRoomHandler_ValidationErrors(t *testing.T) {
	memStore := storage.NewMemoryStore()

	// Table of test cases
	testCases := []struct {
		name           string            // Test case name
		username       string            // Input username
		roomSecretKey  string            // Input room secret key
		expectedStatus int               // Expected HTTP status
		expectedFields map[string]string // Expected field-specific errors
	}{
		{
			name:           "Username contains spaces",
			username:       "User Name",
			roomSecretKey:  "ValidSecretKey",
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{
				"username": "Username cannot contain spaces.",
			},
		},
		{
			name:           "Username too short",
			username:       "Us",
			roomSecretKey:  "ValidSecretKey",
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{
				"username": "Username must be between 3 and 20 characters.",
			},
		},
		{
			name:           "Room secret key contains spaces",
			username:       "ValidUser",
			roomSecretKey:  "Invalid Secret Key",
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{
				"roomSecretKey": "Chat Room Secret Key cannot contain spaces.",
			},
		},
		{
			name:           "Both username and room secret key invalid",
			username:       "Us",
			roomSecretKey:  "Invalid Secret Key",
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{
				"username":      "Username must be between 3 and 20 characters.",
				"roomSecretKey": "Chat Room Secret Key cannot contain spaces.",
			},
		},
		{
			name:           "Both username and room secret key have spaces",
			username:       "User Name",
			roomSecretKey:  "Invalid Secret Key",
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{
				"username":      "Username cannot contain spaces.",
				"roomSecretKey": "Chat Room Secret Key cannot contain spaces.",
			},
		},
		{
			name:           "Valid username and secret key",
			username:       "ValidUser",
			roomSecretKey:  "ValidSecretKey",
			expectedStatus: http.StatusNotFound, // Will proceed to look for the room, which doesn't exist
		},
	}

	// Set up the mux
	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request body
			reqBody := JoinRoomRequest{
				Username:      tc.username,
				RoomSecretKey: tc.roomSecretKey,
			}
			bodyBytes, _ := json.Marshal(reqBody)

			// Build the request
			req := httptest.NewRequest(http.MethodPost, "/api/rooms/testRoomID/users", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			// Assert the status code
			if rr.Code != tc.expectedStatus {
				t.Fatalf("Test [%s] failed: expected status=%d, got=%d", tc.name, tc.expectedStatus, rr.Code)
			}

			// Validate field errors for bad request
			if tc.expectedStatus == http.StatusBadRequest {
				var resp ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Test [%s] failed: failed to parse error response JSON: %v", tc.name, err)
				}
				// Compare expected field errors
				for field, expectedMsg := range tc.expectedFields {
					actualMsg, exists := resp.FieldErrors[field]
					if !exists {
						t.Errorf("Test [%s] failed: expected field error for '%s', but it was missing", tc.name, field)
					} else if actualMsg != expectedMsg {
						t.Errorf("Test [%s] failed: expected error for field '%s' to be '%s', got '%s'", tc.name, field, expectedMsg, actualMsg)
					}
				}
				// Ensure no unexpected field errors exist
				if len(resp.FieldErrors) != len(tc.expectedFields) {
					t.Errorf("Test [%s] failed: expected %d field errors, got %d", tc.name, len(tc.expectedFields), len(resp.FieldErrors))
				}
			}
		})
	}
}

func TestJoinRoomHandler_RoomNotFound(t *testing.T) {
	memStore := storage.NewMemoryStore() // no rooms created

	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))

	// A valid JSON but no such room in memory
	validBody := JoinRoomRequest{Username: "User1", RoomSecretKey: "secret123"}
	bodyBytes, _ := json.Marshal(validBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms/nonExistent/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}
	if resp.FieldErrors["roomId"] != "Chat Room not found." {
		t.Errorf("expected 'Chat Room not found.', got '%s'", resp.FieldErrors["roomId"])
	}
}

func TestJoinRoomHandler_SecretKeyMismatch(t *testing.T) {
	memStore := storage.NewMemoryStore()
	// create a room with the correct key
	roomID := createTestRoom(memStore, "correctSecret", 3)

	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))

	joinReq := JoinRoomRequest{
		Username:      "NewUser",
		RoomSecretKey: "wrongSecret",
	}
	bodyBytes, _ := json.Marshal(joinReq)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+roomID+"/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status=%d, got %d", http.StatusBadRequest, rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}
	if resp.FieldErrors["roomSecretKey"] != "Incorrect Secret Key." {
		t.Errorf("expected 'Incorrect Secret Key.', got '%s'", resp.FieldErrors["roomSecretKey"])
	}
}

func TestJoinRoomHandler_RoomFull(t *testing.T) {
	memStore := storage.NewMemoryStore()

	// Create a room and fill it to capacity (Admin is created inside the func)
	roomID := createTestRoom(memStore, "secret123", 2)
	room, _ := memStore.GetRoom(roomID)
	room.Participants["User1"] = &models.User{ID: "1", Username: "User1", Role: models.Member}

	// Prepare request body
	reqBody := JoinRoomRequest{
		Username:      "NewUser",
		RoomSecretKey: "secret123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Build HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+roomID+"/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))
	mux.ServeHTTP(rr, req)

	// Assert HTTP status
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status=%d, got=%d", http.StatusBadRequest, rr.Code)
	}

	// Parse the response
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert field error
	if resp.FieldErrors["roomId"] != "Chat Room is full." {
		t.Errorf("expected 'Chat Room is full.', got '%s'", resp.FieldErrors["roomId"])
	}
}

func TestJoinRoomHandler_UsernameTaken(t *testing.T) {
	memStore := storage.NewMemoryStore()

	// Create a room with one participant
	roomID := createTestRoom(memStore, "secret123", 3)
	room, _ := memStore.GetRoom(roomID)
	room.Participants["ExistingUser"] = &models.User{ID: "1", Username: "ExistingUser", Role: models.Member}

	// Prepare request body
	reqBody := JoinRoomRequest{
		Username:      "ExistingUser", // Duplicate username
		RoomSecretKey: "secret123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Build HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+roomID+"/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))
	mux.ServeHTTP(rr, req)

	// Assert HTTP status
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status=%d, got=%d", http.StatusBadRequest, rr.Code)
	}

	// Parse the response
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert field error
	if resp.FieldErrors["username"] != "This username is taken." {
		t.Errorf("expected 'This username is taken.', got '%s'", resp.FieldErrors["username"])
	}
}

func TestJoinRoomHandler_GenerateRandomIDFailure(t *testing.T) {
	// Mock GenerateRandomIDFunc to return an error
	originalFunc := util.GenerateRandomIDFunc
	defer func() { util.GenerateRandomIDFunc = originalFunc }()

	util.GenerateRandomIDFunc = func() (string, error) {
		return "", errors.New("mock ID generation error")
	}

	memStore := storage.NewMemoryStore()
	roomID := createTestRoom(memStore, "secret123", 3)

	// Prepare request body
	reqBody := JoinRoomRequest{
		Username:      "NewUser",
		RoomSecretKey: "secret123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Build HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+roomID+"/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))
	mux.ServeHTTP(rr, req)

	// Assert HTTP status
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status=%d, got=%d", http.StatusInternalServerError, rr.Code)
	}

	// Parse the response
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert error message
	if resp.ErrorMessage != "Failed to generate user ID" {
		t.Errorf("expected error message 'Failed to generate user ID', got '%s'", resp.ErrorMessage)
	}
}

func TestJoinRoomHandler_CreateTokenFailure(t *testing.T) {
	// Mock CreateTokenFunc to return an error
	originalFunc := auth.CreateTokenFunc
	defer func() { auth.CreateTokenFunc = originalFunc }()

	auth.CreateTokenFunc = func(claims auth.Claims, secretKey string) (string, error) {
		return "", errors.New("mock token generation error")
	}

	memStore := storage.NewMemoryStore()

	// Create a test room
	roomID := createTestRoom(memStore, "secret123", 3)

	// Prepare request body
	reqBody := JoinRoomRequest{
		Username:      "NewUser",
		RoomSecretKey: "secret123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Build HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+roomID+"/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.Handle("POST /api/rooms/{roomID}/users", JoinRoomHandler(memStore))
	mux.ServeHTTP(rr, req)

	// Assert HTTP status code
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Parse the error response
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert error message
	if resp.ErrorMessage != "Failed to create token" {
		t.Errorf("expected error message 'Failed to create token', got '%s'", resp.ErrorMessage)
	}
}
