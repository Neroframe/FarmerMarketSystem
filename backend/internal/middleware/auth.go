package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"fms/backend/internal/models"
	"fms/backend/internal/utils"
)

// Define context keys for each user type.
type ContextKey string

const (
	AdminContextKey  ContextKey = "admin"
	BuyerContextKey  ContextKey = "buyer"
	FarmerContextKey ContextKey = "farmer"
)

func Authenticate(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the session ID from the request cookies.
		sessionID, err := utils.GetSessionID(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Get the user ID and user type associated with the session ID.
		userID, userType, err := utils.GetUserIDFromSession(db, sessionID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Fetch user information based on user type and store in context with specific key.
		var ctx context.Context
		switch userType {
		case "admin":
			admin, err := models.GetAdminByID(db, userID)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), AdminContextKey, admin)

		// case "buyer":
		// 	buyer, err := models.GetBuyerByID(db, userID)
		// 	if err != nil {
		// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// 		return
		// 	}
		// 	ctx = context.WithValue(r.Context(), BuyerContextKey, buyer)

		case "farmer":
			farmer, err := models.GetFarmerByID(db, userID)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), FarmerContextKey, farmer)

		default:
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Update the request context and proceed to the next handler.
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// AdminOnly is a middleware that allows only admins to access certain routes.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the admin user from context.
		_, ok := r.Context().Value(AdminContextKey).(*models.Admin)
		if !ok {
			http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}