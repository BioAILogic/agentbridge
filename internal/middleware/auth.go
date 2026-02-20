package middleware

import (
	"net/http"
)

// SessionMiddleware is a placeholder for session authentication.
// Currently passes through all requests. Will be implemented in M2.
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement session validation in M2
		next.ServeHTTP(w, r)
	})
}
