package db

import (
	"database/sql"
	"fmt"
	"log"
)

func NewPostgresDB(connStr string) (*sql.DB, error) {
	log.Println("DEBUG: Opening database connection...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	log.Println("DEBUG: Pinging database...")

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("DEBUG: Database connection successful.")
	return db, nil
}
