package middleware

import (
	"context"
	"net/http"
	"net/url"

	"kseli-server/auth"
	"kseli-server/common"
	"kseli-server/config"
)

func WithMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func ValidateOrigin() func(http.Handler) http.Handler {
	allowedOrigins := map[string]struct{}{
		"localhost:8080": {},
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
				common.WriteError(w, http.StatusForbidden, "Missing Origin header.")
				return
			}

			originURL, err := url.Parse(origin)
			if err != nil || originURL.Host == "" {
				common.WriteError(w, http.StatusBadRequest, "Invalid Origin header.")
				return
			}

			if _, ok := allowedOrigins[originURL.Host]; !ok {
				common.WriteError(w, http.StatusForbidden, "Origin does not match the requested host.")
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
			if apiKey == "" || apiKey != config.APIKey {
				common.WriteError(w, http.StatusUnauthorized, "Invalid or missing API key.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ValidateParticipantSessionID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := r.Header.Get("X-Participant-Session-Id")
			if sessionID == "" {
				common.WriteError(w, http.StatusUnauthorized, "Invalid or missing Session Id.")
				return
			}

			ctx := context.WithValue(r.Context(), auth.ParticipantSessionIDKey, sessionID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ValidateToken() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				common.WriteError(w, http.StatusUnauthorized, "Missing Authorization token.")
				return
			}

			claims, err := auth.ValidateToken(token)
			if err != nil {
				common.WriteError(w, http.StatusUnauthorized, "Invalid or expired token.")
				return
			}

			ctx := context.WithValue(r.Context(), auth.ParticipantClaimsKey, &claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
