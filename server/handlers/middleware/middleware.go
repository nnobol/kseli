package middleware

import (
	"context"
	"net/http"
	"net/url"

	"kseli-server/auth"
	"kseli-server/config"
	"kseli-server/handlers/utils"
	"kseli-server/models"
)

// WithMiddleware chains multiple middleware functions around a base handler.
func WithMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func ValidateOrigin() func(http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"localhost:8080": true,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var origin string

			if r.Method == "GET" {
				origin = r.Header.Get("X-Origin")
			} else {
				origin = r.Header.Get("Origin")
			}

			if origin == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusForbidden, "Missing Origin header")
				return
			}

			originURL, err := url.Parse(origin)
			if err != nil || originURL.Host == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid Origin header")
				return
			}

			if !allowedOrigins[originURL.Host] {
				utils.WriteSimpleErrorMessage(w, http.StatusForbidden, "Origin does not match the requested host")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ValidateAPIKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" || apiKey != config.GlobalConfig.APIKey {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Invalid or missing API key.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ValidateUserSessionID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := r.Header.Get("X-User-Session-Id")
			if sessionID == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Invalid or missing Session Id.")
				return
			}

			ctx := context.WithValue(r.Context(), models.UserSessionIDKey, sessionID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ValidateTokenFromHeader() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Missing Authorization token.")
				return
			}

			claims, err := auth.ValidateToken(token)
			if err != nil {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Invalid or expired token.")
				return
			}

			ctx := context.WithValue(r.Context(), models.UserClaimsKey, &claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
