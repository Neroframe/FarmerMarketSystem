package handlers

import (
	"database/sql"
	"fms/backend/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

// BuyerHandler handles buyer-related requests.
type BuyerHandler struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

// NewBuyerHandler initializes a new BuyerHandler.
func NewBuyerHandler(db *sql.DB, templates map[string]*template.Template) *BuyerHandler {
	return &BuyerHandler{
		DB:        db,
		Templates: templates,
	}
}

// ToggleBuyerStatus toggles the active status of a buyer.
func (h *BuyerHandler) ToggleBuyerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get buyer ID from form data.
	buyerIDStr := r.FormValue("id")
	buyerID, err := strconv.Atoi(buyerIDStr)
	if err != nil {
		http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
		return
	}

	// Retrieve the current active status of the buyer.
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

	// Redirect back to the user list.
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// EditBuyer displays and processes the form to edit buyer details.
func (h *BuyerHandler) EditBuyer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Retrieve buyer ID from query parameters.
		buyerID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
			return
		}

		// Fetch the buyer details.
		buyer, err := models.GetBuyerByID(h.DB, buyerID)
		if err != nil {
			http.Error(w, "Buyer not found", http.StatusNotFound)
			return
		}

		// Render the template with buyer data.
		data := map[string]interface{}{"Buyer": buyer}
		err = h.Templates["edit_buyer"].Execute(w, data)
		if err != nil {
			log.Printf("Template rendering error: %v", err)
			http.Error(w, "Error rendering edit page", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Retrieve and parse buyer ID from form.
		buyerID, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
			return
		}

		// Create a Buyer instance from form data.
		updatedBuyer := models.Buyer{
			ID:              buyerID,
			Email:           r.FormValue("email"),
			FirstName:       r.FormValue("first_name"),
			LastName:        r.FormValue("last_name"),
			DeliveryAddress: r.FormValue("delivery_address"),
			IsActive:        r.FormValue("is_active") == "on", // Set to true if checkbox is checked
		}

		// Update the buyer in the database.
		err = models.UpdateBuyer(h.DB, updatedBuyer)
		if err != nil {
			log.Printf("Error updating buyer: %v", err)
			http.Error(w, "Error updating buyer", http.StatusInternalServerError)
			return
		}

		// Redirect to the user list after successful update.
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// DeleteBuyer removes a buyer account from the system.
func (h *BuyerHandler) DeleteBuyer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get buyer ID from form data.
	buyerIDStr := r.FormValue("id")
	buyerID, err := strconv.Atoi(buyerIDStr)
	if err != nil {
		http.Error(w, "Invalid buyer ID", http.StatusBadRequest)
		return
	}

	// Delete buyer from the database.
	_, err = h.DB.Exec("DELETE FROM buyers WHERE id = $1", buyerID)
	if err != nil {
		http.Error(w, "Failed to delete buyer", http.StatusInternalServerError)
		return
	}

	// Redirect back to the user list.
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
