package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/middleware"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/models"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/utils"
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
func (h *FarmerHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define a struct to parse the incoming JSON payload.
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FarmName  string `json:"farm_name"`
		FarmSize  string `json:"farm_size"`
		Location  string `json:"location"`
	}

	// Decode the JSON payload into the struct.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate that all required fields are present.
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" || req.FarmName == "" || req.FarmSize == "" || req.Location == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Check if a farmer with the provided email already exists.
	existingFarmer, err := models.GetFarmerByEmail(h.DB, req.Email)
	if err == nil && existingFarmer != nil {
		http.Error(w, "Farmer with this email already exists", http.StatusConflict)
		return
	}

	// Hash the provided password using bcrypt.
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create a new Farmer object with the provided details.
	newFarmer := &models.Farmer{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		FarmName:     req.FarmName,
		FarmSize:     req.FarmSize,
		Location:     req.Location,
		Status:       "pending", // Initial status
		IsActive:     false,     // Initial active status
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert the new farmer into the database.
	if err := models.CreateFarmer(h.DB, newFarmer); err != nil {
		http.Error(w, "Failed to create farmer", http.StatusInternalServerError)
		return
	}

	// Respond with a success message.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Farmer registered successfully. Awaiting approval.",
	})
}

func (h *FarmerHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define a struct to parse the incoming JSON payload.
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the JSON payload into the struct.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate that both email and password are provided.
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and Password are required", http.StatusBadRequest)
		return
	}

	// Retrieve the farmer by email from the database.
	farmer, err := models.GetFarmerByEmail(h.DB, req.Email)
	if err != nil || farmer == nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Check if the farmer's status is approved and active.
	if farmer.Status != "approved" || !farmer.IsActive {
		http.Error(w, "Account not active or pending approval", http.StatusUnauthorized)
		return
	}

	// Compare the provided password with the stored hashed password.
	if !utils.CheckPasswordHash(req.Password, farmer.PasswordHash) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Create a new session for the authenticated farmer.
	sessionID, err := utils.CreateSession(w, h.DB, farmer.ID, "farmer")
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Respond with a success message and session details.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Login successful",
		"session_id": sessionID, // Optional: Can be omitted since it's set as a cookie
	})
}

func (h *FarmerHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the session ID from the request cookies.
	sessionID, err := utils.GetSessionID(r)
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	// Destroy the session in the database and clear the session cookie.
	if err := utils.DestroySession(h.DB, sessionID); err != nil {
		http.Error(w, "Failed to destroy session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

func (h *FarmerHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the authenticated farmer from the request context.
	// The Authenticate middleware should have attached the farmer data.
	farmer, ok := r.Context().Value(middleware.FarmerContextKey).(*models.Farmer)
	if !ok || farmer == nil {
		http.Error(w, "Unauthorized: Farmer not found in context", http.StatusUnauthorized)
		return
	}

	// Define a response struct to exclude sensitive information.
	type FarmerResponse struct {
		ID        int       `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		FarmName  string    `json:"farm_name"`
		FarmSize  string    `json:"farm_size"`
		Location  string    `json:"location"`
		Status    string    `json:"status"`
		IsActive  bool      `json:"is_active"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	// Populate the response struct with farmer data.
	response := FarmerResponse{
		ID:        farmer.ID,
		Email:     farmer.Email,
		FirstName: farmer.FirstName,
		LastName:  farmer.LastName,
		FarmName:  farmer.FarmName,
		FarmSize:  farmer.FarmSize,
		Location:  farmer.Location,
		Status:    farmer.Status,
		IsActive:  farmer.IsActive,
		CreatedAt: farmer.CreatedAt,
		UpdatedAt: farmer.UpdatedAt,
	}

	// Set response headers.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send the response.
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *FarmerHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	farmer, ok := r.Context().Value(middleware.FarmerContextKey).(*models.Farmer)
	if !ok || farmer == nil {
		http.Error(w, "Unauthorized: Farmer not found in context", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string   `json:"name"`
		CategoryID  int      `json:"category_id"`
		Price       float64  `json:"price"`
		Quantity    int      `json:"quantity"`
		Description string   `json:"description"`
		Images      []string `json:"images"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error decoding AddProduct request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.CategoryID == 0 || req.Price <= 0 || req.Quantity < 0 {
		http.Error(w, "Missing or invalid required fields", http.StatusBadRequest)
		return
	}

	newProduct := models.Product{
		FarmerID:    farmer.ID,
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		Quantity:    req.Quantity,
		Description: req.Description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Images:      req.Images,
	}

	err := models.CreateProduct(h.DB, &newProduct)
	if err != nil {
		log.Printf("Error creating product: %v", err)
		http.Error(w, "Failed to add product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"product": newProduct,
	})
}

func (h *FarmerHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	farmer, ok := r.Context().Value(middleware.FarmerContextKey).(*models.Farmer)
	if !ok || farmer == nil {
		http.Error(w, "Unauthorized: Farmer not found in context", http.StatusUnauthorized)
		return
	}

	products, err := models.GetActiveProducts(h.DB, farmer.ID)
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"products": products,
	})
}

func (h *FarmerHandler) EditProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	farmer, ok := r.Context().Value(middleware.FarmerContextKey).(*models.Farmer)
	if !ok || farmer == nil {
		http.Error(w, "Unauthorized: Farmer not found in context", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID          int      `json:"id"`
		Name        string   `json:"name"`
		CategoryID  int      `json:"category_id"`
		Price       float64  `json:"price"`
		Quantity    int      `json:"quantity"`
		Description string   `json:"description"`
		IsActive    bool     `json:"is_active"`
		Images      []string `json:"images"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error decoding EditProduct request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.ID == 0 || req.Name == "" || req.CategoryID == 0 || req.Price <= 0 || req.Quantity < 0 {
		http.Error(w, "Missing or invalid required fields", http.StatusBadRequest)
		return
	}

	updatedProduct := models.Product{
		ID:          req.ID,
		FarmerID:    farmer.ID,
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		Quantity:    req.Quantity,
		Description: req.Description,
		IsActive:    req.IsActive,
		UpdatedAt:   time.Now(),
		Images:      req.Images,
	}

	err := models.UpdateProduct(h.DB, &updatedProduct)
	if err != nil {
		log.Printf("Error updating product: %v", err)
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"product": updatedProduct,
	})
}

func (h *FarmerHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	farmer, ok := r.Context().Value(middleware.FarmerContextKey).(*models.Farmer)
	if !ok || farmer == nil {
		http.Error(w, "Unauthorized: Farmer not found in context", http.StatusUnauthorized)
		return
	}

	var payload struct {
		ID int `json:"id"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("Error decoding DeleteProduct request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.ID == 0 {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	err := models.DeleteProduct(h.DB, payload.ID, farmer.ID)
	if err != nil {
		log.Printf("Error deleting product: %v", err)
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Product deleted successfully",
	})
}