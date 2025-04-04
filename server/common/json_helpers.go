package common

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message     string            `json:"errorMessage,omitempty"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Message: msg})
}

func WriteFieldErrors(w http.ResponseWriter, status int, fields map[string]string) {
	WriteJSON(w, status, ErrorResponse{
		FieldErrors: fields,
	})
}
