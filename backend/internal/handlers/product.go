package handlers

import (
	"database/sql"
	"encoding/json"
	"fms/backend/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ProductHandler struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

func NewProductHandler(db *sql.DB, templates map[string]*template.Template) *ProductHandler {
	return &ProductHandler{
		DB:        db,
		Templates: templates,
	}
}

func (h *ProductHandler) GetProductDetails(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /buyer/product/{id} request")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the product ID from the URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Fetch product details from the database
	product, err := models.GetProductByID(h.DB, id)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Respond with the product details in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(product); err != nil {
		log.Printf("Error encoding product to JSON: %v", err)
		http.Error(w, "Failed to encode product", http.StatusInternalServerError)
	}
}

