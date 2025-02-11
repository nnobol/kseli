package middleware

import (
	"context"
	"net/http"
	"strings"

	"kseli-server/auth"
	"kseli-server/config"
	"kseli-server/contextutil"
	"kseli-server/handlers"
)

func ValidateAuthToken() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handlers.WriteJSONError(w, http.StatusUnauthorized, "Missing Authorization token.", nil)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				handlers.WriteJSONError(w, http.StatusUnauthorized, "Invalid Authorization format. Expected 'Bearer <token>'.", nil)
				return
			}
			token := tokenParts[1]

			claims, err := auth.ValidateToken(token, config.GlobalConfig.SecretKey)
			if err != nil {
				handlers.WriteJSONError(w, http.StatusUnauthorized, "Invalid or expired token.", nil)
				return
			}

			ctx := context.WithValue(r.Context(), contextutil.UserClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
