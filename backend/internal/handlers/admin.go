package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/Neroframe/FarmerMarketSystem/backend/internal/middleware"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/models"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/utils"
)

type AdminHandler struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

func NewAdminHandler(db *sql.DB, templates map[string]*template.Template) *AdminHandler {
	return &AdminHandler{
		DB:        db,
		Templates: templates,
	}
}

func (h *AdminHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		csrfToken, err := utils.SetCSRFToken(w)
		if err != nil {
			log.Printf("Error setting CSRF token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = h.Templates["register"].Execute(w, map[string]string{"CSRFToken": csrfToken})
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		// Validate CSRF token
		err := utils.ValidateCSRFToken(r)
		if err != nil {
			log.Printf("Invalid CSRF token: %v", err)
			http.Error(w, "Invalid CSRF Token", http.StatusForbidden)
			return
		}

		// Parse form data
		err = r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Input validation
		if email == "" || password == "" || confirmPassword == "" {
			http.Error(w, "Email and Password are required", http.StatusBadRequest)
			return
		}

		if password != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		exists, err := models.CheckAdminExists(h.DB, email)
		if err != nil {
			log.Printf("Error checking admin existence: %v", err)
			http.Error(w, "Internal Server Error: CheckAdminExists", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "An admin with this email already exists", http.StatusConflict)
			return
		}

		// Hash the password
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create the new admin
		admin := &models.Admin{
			Email:        email,
			PasswordHash: hashedPassword,
			IsActive:     true,
		}

		err = models.CreateAdmin(h.DB, admin)
		if err != nil {
			log.Printf("Error creating admin: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect to login page after successful registration
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		csrfToken, err := utils.SetCSRFToken(w)
		if err != nil {
			log.Printf("Error setting CSRF token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = h.Templates["login"].Execute(w, map[string]string{"CSRFToken": csrfToken})
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		err := utils.ValidateCSRFToken(r)
		if err != nil {
			log.Printf("Invalid CSRF token: %v", err)
			http.Error(w, "Invalid CSRF Token", http.StatusForbidden)
			return
		}

		err = r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			http.Error(w, "Email and Password are required", http.StatusBadRequest)
			return
		}

		admin, err := models.AuthenticateAdmin(h.DB, email, password)
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Create a new session for the authenticated admin
		_, err = utils.CreateSession(w, h.DB, admin.ID, "admin")
		if err != nil {
			log.Printf("Error creating session: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := utils.GetSessionID(r)
	if err == nil {
		err = utils.DestroySession(h.DB, sessionID)
		if err != nil {
			log.Printf("Error destroying session: %v", err)
		}
	}

	// Delete the session cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Retrieve the admin from context
	admin, ok := r.Context().Value(middleware.AdminContextKey).(*models.Admin)
	if !ok {
		log.Println("Admin not found in context. Redirecting to login.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Prepare data to pass to the template
	data := map[string]interface{}{
		"Email": admin.Email,
	}

	// Fetch pending farmers 
	pendingFarmers, err := models.GetPendingFarmers(h.DB)
	if err != nil {
		log.Printf("Error fetching pending farmers: %v", err)
		http.Error(w, "Could not retrieve pending farmers", http.StatusInternalServerError)
		return
	}

	var displayFarmers []map[string]interface{}
	for _, farmer := range pendingFarmers {
		displayFarmer := map[string]interface{}{
			"ID":       farmer.ID,
			"Name":     farmer.FirstName + " " + farmer.LastName,
			"Email":    farmer.Email,
			"FarmSize": farmer.FarmSize,
			"Location": farmer.Location,
		}
		displayFarmers = append(displayFarmers, displayFarmer)
	}
	data["PendingFarmers"] = displayFarmers

	// Render the dashboard template.
	err = h.Templates["dashboard"].Execute(w, data)
	if err != nil {
		log.Printf("Error rendering dashboard template: %v", err)
		http.Error(w, "Could not render dashboard", http.StatusInternalServerError)
		return
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	farmers, err := models.GetAllFarmers(h.DB)
	if err != nil {
		http.Error(w, "Could not retrieve farmers", http.StatusInternalServerError)
		log.Printf("Error retrieving farmers: %v", err)
		return
	}

	buyers, err := models.GetAllBuyers(h.DB)
	if err != nil {
		http.Error(w, "Could not retrieve buyers", http.StatusInternalServerError)
		log.Printf("Error retrieving buyers: %v", err)
		return
	}

	data := map[string]interface{}{
		"Farmers": farmers,
		"Buyers":  buyers,
	}

	err = h.Templates["user_list"].Execute(w, data)
	if err != nil {
		http.Error(w, "Could not render users list", http.StatusInternalServerError)
		log.Printf("Template execution error in ListUsers: %v", err)
		return
	}
}
