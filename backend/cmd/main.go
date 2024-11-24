package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"fms/backend/internal/db"
	"fms/backend/internal/handlers"
	"fms/backend/internal/middleware"

	_ "github.com/lib/pq"
)

func main() {
	// Fetch DATABASE_URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set. Ensure it's available in your environment variables.")
	}

	dbConn, err := db.NewPostgresDB(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer dbConn.Close()

	log.Println("Successfully connected to the database!")

	templates, err := parseTemplates("web/templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	adminHandler := handlers.NewAdminHandler(dbConn, templates)
	farmerHandler := handlers.NewFarmerHandler(dbConn, templates)
	buyerHandler := handlers.NewBuyerHandler(dbConn, templates)
	productHandler := handlers.NewProductHandler(dbConn, templates)

	http.Handle("/favicon.ico", http.HandlerFunc(http.NotFound))

	http.HandleFunc("/register", adminHandler.Register)
	http.HandleFunc("/login", adminHandler.Login)
	http.Handle("/logout", middleware.Authenticate(dbConn, http.HandlerFunc(adminHandler.Logout)))

	// Dashboard routes
	http.Handle("/dashboard", middleware.Authenticate(dbConn, http.HandlerFunc(adminHandler.Dashboard)))
	http.Handle("/dashboard/pending-farmers", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ListPendingFarmers))))
	http.Handle("/dashboard/farmer-profile", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ViewFarmerProfile))))
	http.Handle("/dashboard/approve-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ApproveFarmer))))
	http.Handle("/dashboard/reject-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.RejectFarmer))))

	// User management routes
	http.Handle("/admin/users", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(adminHandler.ListUsers))))

	http.Handle("/admin/users/toggle-farmer-status", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ToggleFarmerStatus))))
	http.Handle("/admin/users/edit-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.EditFarmer))))
	http.Handle("/admin/users/delete-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.DeleteFarmer))))

	http.Handle("/admin/users/toggle-buyer-status", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(buyerHandler.ToggleBuyerStatus))))
	http.Handle("/admin/users/edit-buyer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(buyerHandler.EditBuyer))))
	http.Handle("/admin/users/delete-buyer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(buyerHandler.DeleteBuyer))))

	// Buyer Routes
	http.Handle("/buyer/register", middleware.CORS(http.HandlerFunc(buyerHandler.Register)))
	http.Handle("/buyer/login", middleware.CORS(http.HandlerFunc(buyerHandler.Login)))
	http.Handle("/buyer/logout", middleware.CORS(middleware.Authenticate(dbConn, http.HandlerFunc(buyerHandler.Logout))))
	// home - search, categories
	http.Handle("/buyer/home", middleware.CORS(http.HandlerFunc(buyerHandler.Home)))
	http.Handle("/buyer/product/", middleware.CORS(http.HandlerFunc(productHandler.GetProductDetails)))

	// Farmer Routes
	http.Handle("/farmer/register", middleware.CORS(http.HandlerFunc(farmerHandler.Register)))
	http.Handle("/farmer/login", middleware.CORS(http.HandlerFunc(farmerHandler.Login)))
	http.Handle("/farmer/logout", middleware.CORS(middleware.Authenticate(dbConn, http.HandlerFunc(farmerHandler.Logout))))

	// dashboard - list products, manage inventory, track sales
	http.Handle("/farmer/dashboard", middleware.CORS(middleware.Authenticate(dbConn, http.HandlerFunc(farmerHandler.Dashboard))))
	http.Handle("/farmer/product/add-product", middleware.CORS(middleware.Authenticate(dbConn, http.HandlerFunc(farmerHandler.AddProduct))))
	http.Handle("/farmer/product/list-products", middleware.CORS(middleware.Authenticate(dbConn, http.HandlerFunc(farmerHandler.ListProducts))))
	http.Handle("/farmer/product/edit-product", middleware.CORS(middleware.Authenticate(dbConn, http.HandlerFunc(farmerHandler.EditProduct))))
	http.Handle("/farmer/product/delete-product", middleware.CORS(middleware.Authenticate(dbConn, http.HandlerFunc(farmerHandler.DeleteProduct))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func parseTemplates(pattern string) (map[string]*template.Template, error) {
	tmplMap := make(map[string]*template.Template)

	// Parse all templates matching the pattern
	templates, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	// Iterate over each parsed template
	for _, tmpl := range templates.Templates() {
		name := tmpl.Name()
		base := filepath.Base(name)
		key := base[:len(base)-len(".html")]

		tmplMap[key] = tmpl // Map "login" to the "login.html" template
	}

	return tmplMap, nil
}
