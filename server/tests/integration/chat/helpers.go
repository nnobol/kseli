package chat_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func createRoom(t *testing.T, mustCreate bool, expectedBadStatus int, handler http.Handler, maxParticipants uint8, username, origin, apiKey, sessionID string) (chat.CreateRoomResponse, common.ErrorResponse) {
	headers := map[string]string{
		"Origin":                   origin,
		"X-Api-Key":                apiKey,
		"X-Participant-Session-Id": sessionID,
	}
	body, _ := json.Marshal(chat.CreateRoomRequest{
		Username:        username,
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

func joinRoom(t *testing.T, mustJoin bool, expectedBadStatus int, handler http.Handler, username, origin, token, sessionID string) (chat.JoinRoomResponse, common.ErrorResponse) {
	headers := map[string]string{
		"Origin":                   origin,
		"Authorization":            token,
		"X-Participant-Session-Id": sessionID,
	}
	body, _ := json.Marshal(chat.JoinRoomRequest{
		Username: username,
	})

	status, respBody := sendRequest(handler, http.MethodPost, "/api/rooms/join", bytes.NewReader(body), headers)

	if mustJoin {
		if status != http.StatusCreated {
			t.Fatalf("expected 201, got %d, body: %s", status, string(respBody))
		}
		var resp chat.JoinRoomResponse
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
		return chat.JoinRoomResponse{}, errResp
	}
}

func getRoom(t *testing.T, mustGet, isAdmin bool, expectedBadStatus int, handler http.Handler, roomID, origin, token string) (string, chat.GetRoomResponse, common.ErrorResponse) {
	headers := map[string]string{
		"X-Origin":      origin,
		"Authorization": token,
	}

	status, respBody := sendRequest(handler, http.MethodGet, "/api/rooms/"+url.PathEscape(roomID), nil, headers)

	if mustGet {
		if status != http.StatusOK {
			t.Fatalf("expected 200, got %d, body: %s", status, string(respBody))
		}
		var resp chat.GetRoomResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			t.Fatalf("failed to unmarshal success resp: %v", err)
		}
		var token string
		if isAdmin {
			parts := strings.Split(resp.InviteLink, "?invite=")
			if len(parts) != 2 {
				t.Fatalf("bad invite link %q", resp.InviteLink)
			}
			token = parts[1]
		}
		return token, resp, common.ErrorResponse{}
	} else {
		if status != expectedBadStatus {
			t.Fatalf("expected %d, got %d, body: %s", expectedBadStatus, status, string(respBody))
		}
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			t.Fatalf("failed to unmarshal error resp: %v", err)
		}
		return "", chat.GetRoomResponse{}, errResp
	}
}

func deleteRoom(t *testing.T, mustDel bool, expectedBadStatus int, handler http.Handler, roomID, origin, token string) common.ErrorResponse {
	headers := map[string]string{
		"Origin":        origin,
		"Authorization": token,
	}

	status, respBody := sendRequest(handler, http.MethodDelete, "/api/rooms/"+url.PathEscape(roomID), nil, headers)

	if mustDel {
		if status != http.StatusNoContent {
			t.Fatalf("expected 204, got %d, body: %s", status, string(respBody))
		}

		return common.ErrorResponse{}
	} else {
		if status != expectedBadStatus {
			t.Fatalf("expected %d, got %d, body: %s", expectedBadStatus, status, string(respBody))
		}
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			t.Fatalf("failed to unmarshal error resp: %v", err)
		}
		return errResp
	}
}

func kickOrBanUser(t *testing.T, mustSucceed bool, expectedBadStatus int, handler http.Handler, userID uint8, action, roomID, origin, token string) common.ErrorResponse {
	headers := map[string]string{
		"Origin":        origin,
		"Authorization": token,
	}
	body, _ := json.Marshal(chat.UserRequest{
		TargetUserID: userID,
	})

	status, respBody := sendRequest(handler, http.MethodPost, "/api/rooms/"+url.PathEscape(roomID)+"/"+action, bytes.NewReader(body), headers)

	if mustSucceed {
		if status != http.StatusNoContent {
			t.Fatalf("expected 204, got %d, body: %s", status, string(respBody))
		}
		return common.ErrorResponse{}
	} else {
		if status != expectedBadStatus {
			t.Fatalf("expected %d, got %d, body: %s", expectedBadStatus, status, string(respBody))
		}
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			t.Fatalf("failed to unmarshal error resp: %v", err)
		}
		return errResp
	}
}
