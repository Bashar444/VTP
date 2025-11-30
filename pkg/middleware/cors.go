package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment variable
		allowedOrigins := os.Getenv("CORS_ORIGINS")
		if allowedOrigins == "" {
			// Default to common origins in development
			allowedOrigins = "http://localhost:3000,https://vtp-mu.vercel.app,https://vtp-git-main-bashar-al-aghas-projects.vercel.app"
		}

		// Get the request origin
		origin := r.Header.Get("Origin")

		// Check if origin is in allowed list
		if origin != "" {
			origins := strings.Split(allowedOrigins, ",")
			for _, allowedOrigin := range origins {
				allowedOrigin = strings.TrimSpace(allowedOrigin)
				if origin == allowedOrigin || allowedOrigin == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// Set other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
