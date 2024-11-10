package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Buyer represents a buyer in the system.
type Buyer struct {
	ID                  int
	Email               string
	PasswordHash        string
	FirstName           string
	LastName            string
	DeliveryAddress     string
	DeliveryPreferences map[string]interface{} // Use a map to represent JSONB fields
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// GetAllBuyers retrieves all buyers from the database.
func GetAllBuyers(db *sql.DB) ([]Buyer, error) {
	rows, err := db.Query(`
		SELECT id, email, first_name, last_name, delivery_address, delivery_preferences, is_active, created_at, updated_at
		FROM buyers
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buyers []Buyer
	for rows.Next() {
		var buyer Buyer
		var deliveryPreferencesJSON []byte // To scan JSONB data

		err := rows.Scan(
			&buyer.ID,
			&buyer.Email,
			&buyer.FirstName,
			&buyer.LastName,
			&buyer.DeliveryAddress,
			&deliveryPreferencesJSON, // Scan JSONB data as bytes
			&buyer.IsActive,
			&buyer.CreatedAt,
			&buyer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSONB data into a map
		if len(deliveryPreferencesJSON) > 0 {
			err = json.Unmarshal(deliveryPreferencesJSON, &buyer.DeliveryPreferences)
			if err != nil {
				return nil, err
			}
		}

		buyers = append(buyers, buyer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buyers, nil
}

// GetBuyerByID retrieves a buyer's details by their ID.
func GetBuyerByID(db *sql.DB, buyerID int) (*Buyer, error) {
	var buyer Buyer
	var deliveryPreferencesJSON []byte

	err := db.QueryRow(`
		SELECT id, email, first_name, last_name, delivery_address, delivery_preferences, is_active, created_at, updated_at
		FROM buyers
		WHERE id = $1`, buyerID).
		Scan(
			&buyer.ID,
			&buyer.Email,
			&buyer.FirstName,
			&buyer.LastName,
			&buyer.DeliveryAddress,
			&deliveryPreferencesJSON,
			&buyer.IsActive,
			&buyer.CreatedAt,
			&buyer.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB data into a map
	if len(deliveryPreferencesJSON) > 0 {
		err = json.Unmarshal(deliveryPreferencesJSON, &buyer.DeliveryPreferences)
		if err != nil {
			return nil, err
		}
	}

	return &buyer, nil
}

// UpdateBuyer updates a buyer's details in the database.
func UpdateBuyer(db *sql.DB, buyer Buyer) error {
	_, err := db.Exec(`
        UPDATE buyers
        SET email = $1, first_name = $2, last_name = $3, delivery_address = $4, is_active = $5, updated_at = $6
        WHERE id = $7`,
		buyer.Email, buyer.FirstName, buyer.LastName, buyer.DeliveryAddress, buyer.IsActive, time.Now(), buyer.ID,
	)
	return err
}