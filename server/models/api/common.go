package api

type ErrorResponse struct {
	StatusCode   int
	ErrorMessage string            `json:"errorMessage,omitempty"`
	FieldErrors  map[string]string `json:"fieldErrors,omitempty"`
}
