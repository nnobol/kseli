package middleware

import (
	"net/http"
	"net/url"

	"kseli-server/handlers"
)

func ValidateOrigin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				handlers.WriteJSONError(w, http.StatusForbidden, "Missing Origin header", nil)
				return
			}

			originURL, err := url.Parse(origin)
			if err != nil || originURL.Host == "" {
				handlers.WriteJSONError(w, http.StatusBadRequest, "Invalid Origin header", nil)
				return
			}

			if originURL.Host != r.Host {
				handlers.WriteJSONError(w, http.StatusForbidden, "Origin does not match the requested host", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
