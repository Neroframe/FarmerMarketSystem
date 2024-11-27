package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Buyer struct {
	ID                  int                    `json:"id"`
	Email               string                 `json:"email"`
	PasswordHash        string                 `json:"-"`
	FirstName           string                 `json:"first_name"`
	LastName            string                 `json:"last_name"`
	DeliveryAddress     string                 `json:"delivery_address"`
	DeliveryPreferences map[string]interface{} `json:"delivery_preferences"`
	IsActive            bool                   `json:"is_active"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

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

func UpdateBuyer(db *sql.DB, buyer Buyer) error {
	_, err := db.Exec(`
        UPDATE buyers
        SET email = $1, first_name = $2, last_name = $3, delivery_address = $4, is_active = $5, updated_at = $6
        WHERE id = $7`,
		buyer.Email, buyer.FirstName, buyer.LastName, buyer.DeliveryAddress, buyer.IsActive, time.Now(), buyer.ID,
	)
	return err
}

func GetBuyerByEmail(db *sql.DB, email string) (*Buyer, error) {
	var buyer Buyer
	var deliveryPreferencesJSON []byte

	err := db.QueryRow(`
		SELECT id, email, password_hash, first_name, last_name, delivery_address, delivery_preferences, is_active, created_at, updated_at
		FROM buyers
		WHERE email = $1`, email).
		Scan(
			&buyer.ID,
			&buyer.Email,
			&buyer.PasswordHash,
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

func CreateBuyer(db *sql.DB, buyer *Buyer) error {
	deliveryPreferencesJSON, err := json.Marshal(buyer.DeliveryPreferences)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO buyers (
			email, 
			password_hash, 
			first_name, 
			last_name, 
			delivery_address, 
			delivery_preferences, 
			is_active, 
			created_at, 
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id
	`

	err = db.QueryRow(query,
		buyer.Email,
		buyer.PasswordHash,
		buyer.FirstName,
		buyer.LastName,
		buyer.DeliveryAddress,
		deliveryPreferencesJSON,
		buyer.IsActive,
		time.Now(),
		time.Now(),
	).Scan(&buyer.ID)

	if err != nil {
		return err
	}

	return nil
}


