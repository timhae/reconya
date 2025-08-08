// package middleware

// import (
// 	"net/http"
// )

// // CORS middleware to handle cross-origin request issues.
// func CORS(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Set headers
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
// 		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

// 		// Check if the request is for CORS preflight
// 		if r.Method == "OPTIONS" {
// 			// Preflight request; no further processing required
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}

// 		// Call the next handler, which can be another middleware in the chain, or the final handler.
// 		next.ServeHTTP(w, r)
// 	})
// }

package middleware

import (
	"net/http"
)

func SetupCORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow requests from any origin in docker environment
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
