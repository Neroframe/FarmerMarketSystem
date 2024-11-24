package middleware

import (
	"log"
	"net/http"
)

// List of allowed origins. Replace these with your actual frontend URLs.
var allowedOrigins = []string{
	"http://localhost:19006",             // Expo Go on localhost
	"http://172.22.16.140:19006",         // Expo Go on your local network (adjust the port if different)
	"https://your-production-domain.com", // Production domain
	"http://localhost:8081",
	"http://0.0.0.0:8080",
	"http://127.0.1.1:8080",
}

// isOriginAllowed checks if the request's origin is in the allowedOrigins list.
func isOriginAllowed(origin string) bool {
	for _, o := range allowedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		log.Printf("CORS Middleware: %s %s Origin: %s", r.Method, r.URL.Path, origin)

		if isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// Optionally, you can handle disallowed origins here
			log.Printf("CORS Middleware: Origin %s not allowed", origin)
		}

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			log.Println("CORS Middleware: Preflight request handled")
			return
		}

		next.ServeHTTP(w, r)
	})
}
