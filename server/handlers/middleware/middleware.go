package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"

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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusForbidden, "Missing Origin header")
				return
			}

			originURL, err := url.Parse(origin)
			if err != nil || originURL.Host == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusBadRequest, "Invalid Origin header")
				return
			}

			if originURL.Host != r.Host {
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

func ValidateSessionID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := r.Header.Get("X-Session-Id")
			if sessionID == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Invalid or missing Session Id.")
				return
			}

			ctx := context.WithValue(r.Context(), models.UserSessionIDKey, sessionID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ValidateFingerprint() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fingerprint := r.Header.Get("X-Fingerprint")
			if fingerprint == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Invalid or missing Fingerprint.")
				return
			}

			ctx := context.WithValue(r.Context(), models.UserFingerprintKey, fingerprint)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ValidateAuthToken() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Missing Authorization token.")
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				utils.WriteSimpleErrorMessage(w, http.StatusUnauthorized, "Invalid Authorization format. Expected 'Bearer <token>'.")
				return
			}
			token := tokenParts[1]

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
