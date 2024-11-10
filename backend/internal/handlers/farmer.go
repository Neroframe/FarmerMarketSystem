package handlers

import (
	"database/sql"
	"fms/backend/internal/middleware"
	"fms/backend/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

// FarmerHandler handles farmer-related requests.
type FarmerHandler struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

// NewFarmerHandler initializes a new FarmerHandler.
func NewFarmerHandler(db *sql.DB, templates map[string]*template.Template) *FarmerHandler {
	return &FarmerHandler{
		DB:        db,
		Templates: templates,
	}
}

// ListPendingFarmers displays a list of pending farmers (admin-only).
func (h *FarmerHandler) ListPendingFarmers(w http.ResponseWriter, r *http.Request) {
	// Ensure only admins can access this handler.
	_, ok := r.Context().Value(middleware.AdminContextKey).(*models.Admin)
	if !ok {
		http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
		log.Println("Unauthorized access attempt to ListPendingFarmers")
		return
	}

	// Retrieve pending farmers from the model.
	farmers, err := models.GetPendingFarmers(h.DB)
	if err != nil {
		http.Error(w, "Failed to retrieve pending farmers", http.StatusInternalServerError)
		log.Printf("Error retrieving pending farmers: %v", err)
		return
	}

	// Prepare data for the template rendering.
	data := map[string]interface{}{
		"Farmers": farmers,
	}

	// Render the template.
	err = h.Templates["pending_farmers"].Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Template execution error in ListPendingFarmers: %v", err)
		return
	}
}

// ViewFarmerProfile displays the profile of a farmer.
func (h *FarmerHandler) ViewFarmerProfile(w http.ResponseWriter, r *http.Request) {
	// Ensure only admins can access this handler.
	_, ok := r.Context().Value(middleware.AdminContextKey).(*models.Admin)
	if !ok {
		http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
		log.Println("Unauthorized access attempt to ViewFarmerProfile")
		return
	}

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

	// Fetch farmer details from the model.
	farmer, err := models.GetFarmerByID(h.DB, farmerID)
	if err != nil {
		http.Error(w, "Farmer not found", http.StatusNotFound)
		log.Printf("Farmer with ID %d not found: %v", farmerID, err)
		return
	}

	// Prepare data for the template, adding concatenated Name.
	data := map[string]interface{}{
		"Farmer": farmer,
		"Name":   farmer.FirstName + " " + farmer.LastName,
	}

	// Render the farmer profile template.
	err = h.Templates["farmer_profile"].Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering farmer profile", http.StatusInternalServerError)
		log.Printf("Template execution error in ViewFarmerProfile: %v", err)
		return
	}
}

// ApproveFarmer approves a farmer's application (admin-only).
func (h *FarmerHandler) ApproveFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure only admins can access this handler.
	_, ok := r.Context().Value(middleware.AdminContextKey).(*models.Admin)
	if !ok {
		http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
		log.Println("Unauthorized access attempt to ApproveFarmer")
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

	// Redirect back to the dashboard.
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// RejectFarmer rejects a farmer's application with a reason (admin-only).
func (h *FarmerHandler) RejectFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure only admins can access this handler.
	_, ok := r.Context().Value(middleware.AdminContextKey).(*models.Admin)
	if !ok {
		http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
		log.Println("Unauthorized access attempt to RejectFarmer")
		return
	}

	// Get farmer ID and rejection reason from form data.
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

	// Log the rejection reason for reference.
	log.Printf("Farmer ID %d rejected for reason: %s", farmerID, reason)

	// Redirect back to the dashboard.
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ToggleFarmerStatus toggles the active status of a farmer.
func (h *FarmerHandler) ToggleFarmerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get farmer ID from form data.
	farmerIDStr := r.FormValue("id")
	farmerID, err := strconv.Atoi(farmerIDStr)
	if err != nil {
		http.Error(w, "Invalid farmer ID", http.StatusBadRequest)
		return
	}

	// Retrieve the current active status of the farmer.
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

	// Redirect back to the user list.
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// EditFarmer displays and processes the form to edit farmer details.
func (h *FarmerHandler) EditFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Retrieve farmer ID from query parameters.
		farmerID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid farmer ID", http.StatusBadRequest)
			return
		}

		// Fetch the farmer details.
		farmer, err := models.GetFarmerByID(h.DB, farmerID)
		if err != nil {
			http.Error(w, "Farmer not found", http.StatusNotFound)
			return
		}

		// Render the template with farmer data.
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

		// Create a Farmer instance from form data.
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

		// Redirect to the user list after successful update.
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// DeleteFarmer removes a farmer account from the system.
func (h *FarmerHandler) DeleteFarmer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get farmer ID from form data.
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

	// Redirect back to the user list.
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
