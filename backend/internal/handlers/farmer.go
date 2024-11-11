package handlers

import (
	"database/sql"
	"fms/backend/internal/models"
	"fms/backend/internal/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type FarmerHandler struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

func NewFarmerHandler(db *sql.DB, templates map[string]*template.Template) *FarmerHandler {
	return &FarmerHandler{
		DB:        db,
		Templates: templates,
	}
}

// admin-only funcs
func (h *FarmerHandler) ListPendingFarmers(w http.ResponseWriter, r *http.Request) {
	farmers, err := models.GetPendingFarmers(h.DB)
	if err != nil {
		http.Error(w, "Failed to retrieve pending farmers", http.StatusInternalServerError)
		log.Printf("Error retrieving pending farmers: %v", err)
		return
	}

	data := map[string]interface{}{
		"Farmers": farmers,
	}

	err = h.Templates["pending_farmers"].Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Template execution error in ListPendingFarmers: %v", err)
		return
	}
}

func (h *FarmerHandler) ViewFarmerProfile(w http.ResponseWriter, r *http.Request) {
	// Get farmer ID from URL query.
	farmerIDStr := r.URL.Query().Get("id")
	if farmerIDStr == "" {
		http.Error(w, "Bad Request: Missing farmer ID", http.StatusBadRequest)
		return
	}

	farmerID, err := strconv.Atoi(farmerIDStr)
	if err != nil {
		http.Error(w, "Bad Request: Invalid farmer ID", http.StatusBadRequest)
		return
	}

	farmer, err := models.GetFarmerByID(h.DB, farmerID)
	if err != nil {
		http.Error(w, "Farmer not found", http.StatusNotFound)
		log.Printf("Farmer with ID %d not found: %v", farmerID, err)
		return
	}

	data := map[string]interface{}{
		"Farmer": farmer,
		"Name":   farmer.FirstName + " " + farmer.LastName,
	}

	err = h.Templates["farmer_profile"].Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering farmer profile", http.StatusInternalServerError)
		log.Printf("Template execution error in ViewFarmerProfile: %v", err)
		return
	}
}

func (h *FarmerHandler) ApproveFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get farmer ID from form data.
	farmerIDStr := r.FormValue("id")
	if farmerIDStr == "" {
		http.Error(w, "Bad Request: Missing farmer ID", http.StatusBadRequest)
		return
	}

	farmerID, err := strconv.Atoi(farmerIDStr)
	if err != nil {
		http.Error(w, "Bad Request: Invalid farmer ID", http.StatusBadRequest)
		return
	}

	// Update farmer status to 'approved' and set approved_at timestamp.
	_, err = h.DB.Exec("UPDATE farmers SET status = $1, approved_at = $2, updated_at = $3 WHERE id = $4", "approved", time.Now(), time.Now(), farmerID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error updating farmer status: %v", err)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *FarmerHandler) RejectFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	farmerIDStr := r.FormValue("id")
	reason := r.FormValue("reason")
	if farmerIDStr == "" || reason == "" {
		http.Error(w, "Bad Request: Missing farmer ID or reason", http.StatusBadRequest)
		return
	}

	farmerID, err := strconv.Atoi(farmerIDStr)
	if err != nil {
		http.Error(w, "Bad Request: Invalid farmer ID", http.StatusBadRequest)
		return
	}

	// Update farmer status to 'rejected' and set the rejection_reason.
	_, err = h.DB.Exec("UPDATE farmers SET status = $1, rejection_reason = $2, updated_at = $3 WHERE id = $4", "rejected", reason, time.Now(), farmerID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error updating farmer rejection status: %v", err)
		return
	}

	log.Printf("Farmer ID %d rejected for reason: %s", farmerID, reason)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *FarmerHandler) ToggleFarmerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	farmerIDStr := r.FormValue("id")
	farmerID, err := strconv.Atoi(farmerIDStr)
	if err != nil {
		http.Error(w, "Invalid farmer ID", http.StatusBadRequest)
		return
	}

	farmer, err := models.GetFarmerByID(h.DB, farmerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Farmer not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Toggle the `is_active` status.
	newIsActive := !farmer.IsActive

	_, err = h.DB.Exec("UPDATE farmers SET is_active = $1, updated_at = $2 WHERE id = $3", newIsActive, time.Now(), farmer.ID)
	if err != nil {
		http.Error(w, "Failed to update farmer active status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *FarmerHandler) EditFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Retrieve farmer ID from query parameters.
		farmerID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid farmer ID", http.StatusBadRequest)
			return
		}

		farmer, err := models.GetFarmerByID(h.DB, farmerID)
		if err != nil {
			http.Error(w, "Farmer not found", http.StatusNotFound)
			return
		}

		data := map[string]interface{}{"Farmer": farmer}

		err = h.Templates["edit_farmer"].Execute(w, data)
		if err != nil {
			log.Printf("Template rendering error: %v", err)
			http.Error(w, "Error rendering edit page", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Retrieve and parse farmer ID from form.
		farmerID, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, "Invalid farmer ID", http.StatusBadRequest)
			return
		}

		updatedFarmer := models.Farmer{
			ID:        farmerID,
			Email:     r.FormValue("email"),
			FirstName: r.FormValue("first_name"),
			LastName:  r.FormValue("last_name"),
			FarmName:  r.FormValue("farm_name"),
			FarmSize:  r.FormValue("farm_size"),
			Location:  r.FormValue("location"),
			Status:    r.FormValue("status"),
			IsActive:  r.FormValue("is_active") == "on", // Set to true if checkbox is checked
		}

		// Update the farmer in the database.
		err = models.UpdateFarmer(h.DB, updatedFarmer)
		if err != nil {
			log.Printf("Error updating farmer: %v", err)
			http.Error(w, "Error updating farmer", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

func (h *FarmerHandler) DeleteFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	farmerIDStr := r.FormValue("id")
	farmerID, err := strconv.Atoi(farmerIDStr)
	if err != nil {
		http.Error(w, "Invalid farmer ID", http.StatusBadRequest)
		return
	}

	// Delete farmer from the database.
	_, err = h.DB.Exec("DELETE FROM farmers WHERE id = $1", farmerID)
	if err != nil {
		http.Error(w, "Failed to delete farmer", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// farmer-specific funcs
// router.HandleFunc("/farmer/register", farmerHandler.Register).Methods("GET", "POST")
func (h *FarmerHandler) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		csrfToken, err := utils.SetCSRFToken(w)
		if err != nil {
			log.Printf("Error setting CSRF token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// no templ for now
		err = h.Templates["farmer_register"].Execute(w, map[string]string{"CSRFToken": csrfToken})
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
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
		confirmPassword := r.FormValue("confirm_password")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		farmName := r.FormValue("farm_name")
		farmSize := r.FormValue("farm_size")
		location := r.FormValue("location")

		if email == "" || password == "" || confirmPassword == "" || firstName == "" || lastName == "" || farmName == "" || farmSize == "" || location == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		if password != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		exists, err := models.CheckFarmerExists(h.DB, email)
		if err != nil {
			log.Printf("Error checking farmer existence: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "A farmer with this email already exists", http.StatusConflict)
			return
		}

		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		farmer := &models.Farmer{
			Email:        email,
			PasswordHash: hashedPassword,
			FirstName:    firstName,
			LastName:     lastName,
			FarmName:     farmName,
			FarmSize:     farmSize,
			Location:     location,
			Status:       "pending",
			IsActive:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = models.CreateFarmer(h.DB, farmer)
		if err != nil {
			log.Printf("Error creating farmer: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/farmer/login", http.StatusSeeOther)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
