package middleware

import (
	"net/http"

	"kseli-server/config"
	"kseli-server/handlers"
)

func ValidateAPIKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" || apiKey != config.GlobalConfig.APIKey {
				handlers.WriteJSONError(w, http.StatusUnauthorized, "Invalid or missing API key.", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
