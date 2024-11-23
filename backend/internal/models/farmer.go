package models

import (
	"database/sql"
	"time"
)

type Farmer struct {
	ID           int
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	FarmName     string
	FarmSize     string
	Location     string
	Status       string // "pending", "approved", or "rejected"
	IsActive     bool   // Active or inactive status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func GetPendingFarmers(db *sql.DB) ([]Farmer, error) {
	rows, err := db.Query(`
		SELECT id, email, first_name, last_name, farm_name, farm_size, location, status, created_at
		FROM farmers
		WHERE status = 'pending'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var farmers []Farmer
	for rows.Next() {
		var farmer Farmer
		err := rows.Scan(
			&farmer.ID,
			&farmer.Email,
			&farmer.FirstName,
			&farmer.LastName,
			&farmer.FarmName,
			&farmer.FarmSize,
			&farmer.Location,
			&farmer.Status,
			&farmer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		farmers = append(farmers, farmer)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return farmers, nil
}

func GetFarmerByID(db *sql.DB, farmerID int) (*Farmer, error) {
	var farmer Farmer
	err := db.QueryRow(`
        SELECT id, email, first_name, last_name, farm_name, farm_size, location, status, is_active, created_at, updated_at
        FROM farmers
        WHERE id = $1`, farmerID).
		Scan(&farmer.ID, &farmer.Email, &farmer.FirstName, &farmer.LastName, &farmer.FarmName, &farmer.FarmSize, &farmer.Location, &farmer.Status, &farmer.IsActive, &farmer.CreatedAt, &farmer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &farmer, nil
}

func GetAllFarmers(db *sql.DB) ([]Farmer, error) {
	rows, err := db.Query(`
		SELECT id, email, first_name, last_name, farm_name, farm_size, location, status, is_active, created_at
		FROM farmers
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var farmers []Farmer
	for rows.Next() {
		var farmer Farmer
		err := rows.Scan(
			&farmer.ID,
			&farmer.Email,
			&farmer.FirstName,
			&farmer.LastName,
			&farmer.FarmName,
			&farmer.FarmSize,
			&farmer.Location,
			&farmer.Status,
			&farmer.IsActive,
			&farmer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		farmers = append(farmers, farmer)
	}
	return farmers, nil
}

func UpdateFarmer(db *sql.DB, farmer Farmer) error {
	_, err := db.Exec(`
        UPDATE farmers
        SET email = $1, first_name = $2, last_name = $3, farm_name = $4, farm_size = $5, location = $6, status = $7, is_active = $8, updated_at = $9
        WHERE id = $10`,
		farmer.Email, farmer.FirstName, farmer.LastName, farmer.FarmName, farmer.FarmSize, farmer.Location, farmer.Status, farmer.IsActive, time.Now(), farmer.ID,
	)
	return err
}

func CheckFarmerExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM farmers WHERE email=$1)`
	err := db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateFarmer(db *sql.DB, farmer *Farmer) error {
	query := `
		INSERT INTO farmers (email, password_hash, first_name, last_name, farm_name, farm_size, location, status, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	err := db.QueryRow(query, farmer.Email, farmer.PasswordHash, farmer.FirstName, farmer.LastName, farmer.FarmName, farmer.FarmSize, farmer.Location, "pending", false, time.Now(), time.Now()).Scan(&farmer.ID)
	if err != nil {
		return err
	}
	return nil
}

func GetFarmerByEmail(db *sql.DB, email string) (*Farmer, error) {
	var farmer Farmer
	err := db.QueryRow(`
		SELECT id, email, password_hash, first_name, last_name, farm_name, farm_size, location, status, is_active, created_at, updated_at
		FROM farmers
		WHERE email = $1
	`, email).Scan(
		&farmer.ID, &farmer.Email, &farmer.PasswordHash, &farmer.FirstName, &farmer.LastName,
		&farmer.FarmName, &farmer.FarmSize, &farmer.Location, &farmer.Status,
		&farmer.IsActive, &farmer.CreatedAt, &farmer.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &farmer, nil
}
