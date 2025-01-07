package middleware

import (
	"net/http"

	"kseli-server/handlers"
)

func ValidateHTTPMethod(allowedMethod string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != allowedMethod {
				handlers.WriteJSONError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
