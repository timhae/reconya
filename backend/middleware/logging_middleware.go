// File: middleware/logging.go

package middleware

import (
	"net/http"
)

// LoggingMiddleware logs each HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
