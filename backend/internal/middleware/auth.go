package middleware

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/Neroframe/FarmerMarketSystem/backend/internal/models"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/utils"
)

type ContextKey string

const (
	AdminContextKey  ContextKey = "admin"
	BuyerContextKey  ContextKey = "buyer"
	FarmerContextKey ContextKey = "farmer"
)

func Authenticate(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := utils.GetSessionID(r)
		if err != nil {
			log.Println("Auth Middleware: couldn't retrieve sessionID")
			// http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		userID, userType, err := utils.GetUserIDFromSession(db, sessionID)
		if err != nil {
			log.Println("Auth Middleware: couldn't retrieve userID")
			// http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Fetch user information store in context with specific key
		var ctx context.Context
		switch userType {
		case "admin":
			admin, err := models.GetAdminByID(db, userID)
			if err != nil {
				log.Println("Auth Middleware: couldn't retrieve AdminID")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), AdminContextKey, admin)

		case "buyer":
			buyer, err := models.GetBuyerByID(db, userID)
			if err != nil {
				log.Println("Auth Middleware: couldn't retrieve BuyerByID")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), BuyerContextKey, buyer)

		case "farmer":
			farmer, err := models.GetFarmerByID(db, userID)
			if err != nil {
				log.Println("Auth Middleware: couldn't retrieve FarmerByID")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), FarmerContextKey, farmer)

		default:
			log.Println("Auth Middleware: couldn't handle userType")
			// http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		log.Println("Auth Middleware: success!")
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(AdminContextKey).(*models.Admin)
		if !ok {
			http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
