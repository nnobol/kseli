package middleware

import (
	"context"
	"net/http"
	"net/url"

	"kseli/auth"
	"kseli/common"
	"kseli/config"
)

func WithMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func ValidateOriginHost(origin string) (int, string) {
	allowedHosts := map[string]struct{}{
		"localhost:3000": {},
		"kseli.app":      {},
		"www.kseli.app":  {},
	}

	if origin == "" {
		return http.StatusForbidden, "Missing Origin header."
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return http.StatusBadRequest, "Invalid Origin header."
	}

	if _, ok := allowedHosts[originURL.Host]; !ok {
		return http.StatusForbidden, "Origin not allowed. Access from this origin is restricted."
	}

	return 0, ""
}

func ValidateOrigin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var origin string

			if r.Method == "GET" {
				origin = r.Header.Get("X-Origin")
			} else {
				origin = r.Header.Get("Origin")
			}

			status, err := ValidateOriginHost(origin)
			if err != "" {
				common.WriteError(w, status, err)
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

func ValidateParticipantToken() func(http.Handler) http.Handler {
	return validateToken[auth.Claims](auth.ParticipantClaimsKey)
}

func ValidateInviteToken() func(http.Handler) http.Handler {
	return validateToken[auth.InviteClaims](auth.InviteClaimsKey)
}

func validateToken[T auth.TokenClaims](contextKey auth.ContextKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				common.WriteError(w, http.StatusUnauthorized, "Missing Authorization token.")
				return
			}

			claims, err := auth.ValidateToken[T](token)
			if err != nil {
				common.WriteError(w, http.StatusUnauthorized, "Invalid or expired token.")
				return
			}

			ctx := context.WithValue(r.Context(), contextKey, &claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
