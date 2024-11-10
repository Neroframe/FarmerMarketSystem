package models

import (
	"database/sql"
	"errors"
	"fms/backend/internal/utils"
	"time"
)

type Admin struct {
	ID           int
	Email        string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func CheckAdminExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM admins WHERE email=$1)`
	err := db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateAdmin(db *sql.DB, admin *Admin) error {
	query := `
        INSERT INTO admins (email, password_hash, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	err := db.QueryRow(query, admin.Email, admin.PasswordHash, admin.IsActive, time.Now(), time.Now()).Scan(&admin.ID)
	if err != nil {
		return err
	}
	return nil
}

func AuthenticateAdmin(db *sql.DB, email, password string) (*Admin, error) {
	admin, err := GetAdminByEmail(db, email)
	if err != nil {
		return nil, errors.New("invalid email")
	}
	if !utils.CheckPasswordHash(password, admin.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}
	return admin, nil
}

func GetAdminByEmail(db *sql.DB, email string) (*Admin, error) {
	admin := &Admin{}
	query := `
        SELECT id, email, password_hash, is_active, created_at, updated_at
        FROM admins
        WHERE email = $1
    `
	err := db.QueryRow(query, email).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.IsActive, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}
	return admin, nil
}

func GetAdminByID(db *sql.DB, id int) (*Admin, error) {
	admin := &Admin{}
	query := `
        SELECT id, email, password_hash, is_active, created_at, updated_at
        FROM admins
        WHERE id = $1
    `
	err := db.QueryRow(query, id).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.IsActive, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}
	return admin, nil
}

func UpdateAdmin(db *sql.DB, admin *Admin) error {
	query := `
        UPDATE admins
        SET email = $1, password_hash = $2, is_active = $3, updated_at = $4
        WHERE id = $5
    `
	_, err := db.Exec(query, admin.Email, admin.PasswordHash, admin.IsActive, time.Now(), admin.ID)
	return err
}

func DeleteAdmin(db *sql.DB, id int) error {
	query := `
        DELETE FROM admins
        WHERE id = $1
    `
	_, err := db.Exec(query, id)
	return err
}
