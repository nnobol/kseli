package utils

import (
	"encoding/json"
	"net/http"

	"kseli-server/models/api"
)

func WriteSuccessResponse[T any](w http.ResponseWriter, statusCode int, data *T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteErrorResponse(w http.ResponseWriter, err *api.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(err)
}

func WriteSimpleErrorMessage(w http.ResponseWriter, statusCode int, message string) {
	WriteErrorResponse(w, &api.ErrorResponse{
		StatusCode:   statusCode,
		ErrorMessage: message,
	})
}
