package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Neroframe/FarmerMarketSystem/backend/internal/middleware"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/models"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/utils"
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

	// Toggle the `is_active` status
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
func (h *BuyerHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email               string                 `json:"email"`
		Password            string                 `json:"password"`
		FirstName           string                 `json:"first_name"`
		LastName            string                 `json:"last_name"`
		DeliveryAddress     string                 `json:"delivery_address"`
		DeliveryPreferences map[string]interface{} `json:"delivery_preferences"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	existingBuyer, err := models.GetBuyerByEmail(h.DB, req.Email)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking existing buyer: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingBuyer != nil {
		http.Error(w, "Buyer with this email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buyer := &models.Buyer{
		Email:               req.Email,
		PasswordHash:        string(hashedPassword),
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		DeliveryAddress:     req.DeliveryAddress,
		DeliveryPreferences: req.DeliveryPreferences,
		IsActive:            true,
	}

	err = models.CreateBuyer(h.DB, buyer)
	if err != nil {
		http.Error(w, "Failed to register buyer", http.StatusInternalServerError)
		return
	}

	response := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{
		ID:    buyer.ID,
		Email: buyer.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response JSON: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully registered buyer with ID: %d", buyer.ID)
}

func (h *BuyerHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed buyer login", http.StatusMethodNotAllowed)
		return
	}

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

	if loginData.Email == "" || loginData.Password == "" {
		http.Error(w, "Email and Password are required", http.StatusBadRequest)
		return
	}

	// log.Printf("Attempting to fetch buyer with email: %s", loginData.Email)
	buyer, err := models.GetBuyerByEmail(h.DB, loginData.Email)
	if err != nil {
		log.Printf("Error fetching buyer: %v", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// log.Printf("Checking password: %s against hash: %s", loginData.Password, buyer.PasswordHash)
	if !utils.CheckPasswordHash(loginData.Password, buyer.PasswordHash) {
		log.Println("Password validation failed")
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	_, err = utils.CreateSession(w, h.DB, buyer.ID, "buyer")
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
	})
}

func (h *BuyerHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Invalidate the user's session or token (implementation depends on your session/token management strategy)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Successfully logged out"}`))
}

func (h *BuyerHandler) Home(w http.ResponseWriter, r *http.Request) {
	buyer, ok := r.Context().Value(middleware.BuyerContextKey).(*models.Buyer)
	if !ok || buyer == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized: Buyer not found in context",
		})
		return
	}

	// Parse query parameters
	queryValues := r.URL.Query()
	filters := make(map[string]string)

	// Category filter
	if category := queryValues.Get("category"); category != "" {
		filters["category"] = category
	}

	// Search term
	if search := queryValues.Get("search"); search != "" {
		filters["search"] = search
	}

	// Sorting option
	if sort := queryValues.Get("sort"); sort != "" {
		filters["sort"] = sort
	}

	// Pagination parameters
	limit := 20 // default limit
	if l := queryValues.Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	page := 1 // default page
	if p := queryValues.Get("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	offset := (page - 1) * limit

	products, err := models.GetProductsWithFilters(h.DB, filters, limit, offset)
	if err != nil {
		http.Error(w, "Internal Server Error: Unable to retrieve products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Internal Server Error: Unable to encode products", http.StatusInternalServerError)
		return
	}
}
