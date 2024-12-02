// handlers/cart_handler.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Neroframe/FarmerMarketSystem/backend/internal/middleware"
	"github.com/Neroframe/FarmerMarketSystem/backend/internal/models"
)

type CartHandler struct {
	DB *sql.DB
}
func NewCartHandler(db *sql.DB) *CartHandler {
	return &CartHandler{DB: db}
}
// GetCart handles GET /cart
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	// Retrieve buyer from context
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

	// Fetch cart items from the database
	cartItems, err := models.GetCartByBuyerID(h.DB, buyer.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve cart", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"success": true,
		"cart":    cartItems,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddToCart handles POST /cart/add
func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	// Retrieve buyer from context
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

	// Parse the request body
	var request struct {
		ProductID int `json:"productId"`
		Quantity  int `json:"quantity"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.ProductID == 0 || request.Quantity < 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request payload",
		})
		return
	}

	// Add the product to the cart
	err = models.AddProductToCart(h.DB, buyer.ID, request.ProductID, request.Quantity)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Failed to add product to cart", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"success": true,
		"message": "Product added to cart successfully",
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveFromCart handles DELETE /cart/remove/{productId}
func (h *CartHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	// Retrieve buyer from context
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

	// Extract productId from the URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	productIDStr := pathParts[3]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	// Remove the product from the cart
	err = models.RemoveProductFromCart(h.DB, buyer.ID, productID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Failed to remove product from cart", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"success": true,
		"message": "Product removed from cart successfully",
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateCart handles POST /cart/update
func (h *CartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	// Retrieve buyer from context
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

	// Parse the request body
	var request struct {
		ProductID int `json:"productId"`
		Quantity  int `json:"quantity"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.ProductID == 0 || request.Quantity < 0 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update the cart item
	err = models.UpdateCartItem(h.DB, buyer.ID, request.ProductID, request.Quantity)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Failed to update cart item", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"success": true,
		"message": "Cart updated successfully",
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
