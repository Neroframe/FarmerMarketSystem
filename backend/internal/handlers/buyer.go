package handlers

import (
	"database/sql"
	"encoding/json"
	"fms/backend/internal/models"
	"fms/backend/internal/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type BuyerHandler struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

func NewBuyerHandler(db *sql.DB, templates map[string]*template.Template) *BuyerHandler {
	return &BuyerHandler{
		DB:        db,
		Templates: templates,
	}
}

// admin-only funcs
func (h *BuyerHandler) ToggleBuyerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	buyerIDStr := r.FormValue("id")
	buyerID, err := strconv.Atoi(buyerIDStr)
	if err != nil {
		http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
		return
	}

	buyer, err := models.GetBuyerByID(h.DB, buyerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Buyer not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Toggle the `is_active` status.
	newIsActive := !buyer.IsActive

	_, err = h.DB.Exec("UPDATE buyers SET is_active = $1, updated_at = $2 WHERE id = $3", newIsActive, time.Now(), buyer.ID)
	if err != nil {
		http.Error(w, "Failed to update buyer active status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *BuyerHandler) EditBuyer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		buyerID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
			return
		}

		buyer, err := models.GetBuyerByID(h.DB, buyerID)
		if err != nil {
			http.Error(w, "Buyer not found", http.StatusNotFound)
			return
		}

		data := map[string]interface{}{"Buyer": buyer}
		err = h.Templates["edit_buyer"].Execute(w, data)
		if err != nil {
			log.Printf("Template rendering error: %v", err)
			http.Error(w, "Error rendering edit page", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		buyerID, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
			return
		}

		updatedBuyer := models.Buyer{
			ID:              buyerID,
			Email:           r.FormValue("email"),
			FirstName:       r.FormValue("first_name"),
			LastName:        r.FormValue("last_name"),
			DeliveryAddress: r.FormValue("delivery_address"),
			IsActive:        r.FormValue("is_active") == "on", // Set to true if checkbox is checked
		}

		err = models.UpdateBuyer(h.DB, updatedBuyer)
		if err != nil {
			log.Printf("Error updating buyer: %v", err)
			http.Error(w, "Error updating buyer", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

func (h *BuyerHandler) DeleteBuyer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	buyerIDStr := r.FormValue("id")
	buyerID, err := strconv.Atoi(buyerIDStr)
	if err != nil {
		http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec("DELETE FROM buyers WHERE id = $1", buyerID)
	if err != nil {
		http.Error(w, "Failed to delete buyer", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// buyer-specific funcs
func (h *BuyerHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Received /buyer/login request")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed buyer login", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		log.Printf("Error decoding login data: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if loginData.Email == "" || loginData.Password == "" {
		http.Error(w, "Email and Password are required", http.StatusBadRequest)
		return
	}

	// Fetch buyer by email
	buyer, err := models.GetBuyerByEmail(h.DB, loginData.Email)
	if err != nil {
		log.Printf("Error fetching buyer: %v", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Validate password
	if !utils.CheckPasswordHash(loginData.Password, buyer.PasswordHash) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	http.Error(w, "success", http.StatusUnauthorized)

	// Respond with token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
	})
}

func (h *BuyerHandler) Register(w http.ResponseWriter, r *http.Request) {
	
}

func (h *BuyerHandler) Logout(w http.ResponseWriter, r *http.Request) {

}