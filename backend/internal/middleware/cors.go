package middleware

import (
	"log"
	"net/http"
)

// TO DO: deploy front and add to origin
// var allowedOrigins = []string{
// 	"https://your-frontend-domain.com",
// 	"https://farmermarketsystem-production.up.railway.app",
// }

// func isOriginAllowed(origin string) bool {
// 	if strings.HasPrefix(origin, "http://localhost:") {
// 		return true
// 	}

// 	if strings.HasSuffix(origin, ".exp.direct") {
// 		return true
// 	}

// 	for _, o := range allowedOrigins {
// 		if o == origin {
// 			return true
// 		}
// 	}
// 	return false
// }

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		log.Printf("CORS Middleware: %s %s Origin: %s", r.Method, r.URL.Path, origin)

		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
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
