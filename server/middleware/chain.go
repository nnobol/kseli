package middleware

import "net/http"

// WithMiddleware chains multiple middleware functions around a base handler.
func WithMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
