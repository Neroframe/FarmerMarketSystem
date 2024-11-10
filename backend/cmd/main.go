package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"fms/backend/internal/db"
	"fms/backend/internal/handlers"
	"fms/backend/internal/middleware"

	_ "github.com/lib/pq" // Replace with your database driver
)

// go build -o fms-backend ./backend/cmd
//
//	./fms-backend
func main() {
	// Set env variables
	// export DB_HOST=172.17.0.1
	// export DB_PORT=5432
	// export DB_USER=postgres
	// export DB_PASSWORD=123
	// export DB_NAME=fms_db
	// Check db connection
	// psql -h <Windows_Host_IP> -U <your_db_user> -d <your_db_name>

	// Initialize database connection
	dbConn, err := db.NewPostgresDB(
		"172.17.0.1", // Host IP:		WIN_IP=$(cat /etc/resolv.conf | grep nameserver | awk '{print $2}') echo $WIN_IP
		"5432",       // Port
		"postgres",   // User
		"123",        // Password
		"fms",        // Database Name
	)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Println("Successfully connected to the database!")

	// Parse all templates in the 'web/templates' directory
	templates, err := parseTemplates("web/templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	// Initialize handlers with the parsed templates
	adminHandler := handlers.NewAdminHandler(dbConn, templates)
	farmerHandler := handlers.NewFarmerHandler(dbConn, templates)
	buyerHandler := handlers.NewBuyerHandler(dbConn, templates)

	// Public routes
	http.HandleFunc("/register", adminHandler.Register)
	http.HandleFunc("/login", adminHandler.Login)

	// Protected routes (require authentication)
	http.Handle("/dashboard", middleware.Authenticate(dbConn, http.HandlerFunc(adminHandler.Dashboard)))
	http.Handle("/logout", middleware.Authenticate(dbConn, http.HandlerFunc(adminHandler.Logout)))

	// Admin-only farmer routes (require admin authorization)
	http.Handle("/pending-farmers", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ListPendingFarmers))))
	http.Handle("/farmer-profile", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ViewFarmerProfile))))
	http.Handle("/approve-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ApproveFarmer))))
	http.Handle("/reject-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.RejectFarmer))))

	// Admin-only user management routes (require admin authorization)
	// Base route for listing all users
	http.Handle("/admin/users", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(adminHandler.ListUsers))))

	// Routes for farmers
	http.Handle("/admin/users/toggle-farmer-status", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.ToggleFarmerStatus))))
	http.Handle("/admin/users/edit-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.EditFarmer))))
	http.Handle("/admin/users/delete-farmer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(farmerHandler.DeleteFarmer))))

	// Routes for buyers
	http.Handle("/admin/users/toggle-buyer-status", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(buyerHandler.ToggleBuyerStatus))))
	http.Handle("/admin/users/edit-buyer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(buyerHandler.EditBuyer))))
	http.Handle("/admin/users/delete-buyer", middleware.Authenticate(dbConn, middleware.AdminOnly(http.HandlerFunc(buyerHandler.DeleteBuyer))))

	// Handle favicon requests
	http.Handle("/favicon.ico", http.HandlerFunc(http.NotFound))

	// Start the server on port 8080
	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// parseTemplates in the specified directory and returns a map.
func parseTemplates(pattern string) (map[string]*template.Template, error) {
	tmplMap := make(map[string]*template.Template)

	// Parse all templates matching the pattern
	templates, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	// Iterate over each parsed template
	for _, tmpl := range templates.Templates() {
		name := tmpl.Name()                  // e.g., "login.html"
		base := filepath.Base(name)          // e.g., "login.html"
		key := base[:len(base)-len(".html")] // e.g., "login"

		tmplMap[key] = tmpl // Map "login" to the "login.html" template
	}

	return tmplMap, nil
}
