package chat_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"kseli/common"
	"kseli/features/chat"
)

func sendRequest(handler http.Handler, method, url string, body io.Reader, headers map[string]string) (status int, respBody []byte) {
	req := httptest.NewRequest(method, url, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w.Result().StatusCode, w.Body.Bytes()
}

func createRoom(t *testing.T, mustCreate bool, expectedBadStatus int, handler http.Handler, maxParticipants uint8, origin, apiKey, sessionID string) (chat.CreateRoomResponse, common.ErrorResponse) {
	headers := map[string]string{
		"Origin":                   origin,
		"X-Api-Key":                apiKey,
		"X-Participant-Session-Id": sessionID,
	}
	body, _ := json.Marshal(chat.CreateRoomRequest{
		Username:        "admin",
		MaxParticipants: maxParticipants,
	})

	status, respBody := sendRequest(handler, http.MethodPost, "/api/rooms", bytes.NewReader(body), headers)

	if mustCreate {
		if status != http.StatusCreated {
			t.Fatalf("expected 201, got %d, body: %s", status, string(respBody))
		}
		var resp chat.CreateRoomResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			t.Fatalf("failed to unmarshal success resp: %v", err)
		}
		return resp, common.ErrorResponse{}
	} else {
		if status != expectedBadStatus {
			t.Fatalf("expected %d, got %d, body: %s", expectedBadStatus, status, string(respBody))
		}
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			t.Fatalf("failed to unmarshal error resp: %v", err)
		}
		return chat.CreateRoomResponse{}, errResp
	}
}
