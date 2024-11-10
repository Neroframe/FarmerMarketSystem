package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
)

// CreateSession creates a new session for the user with a specified user type.
func CreateSession(w http.ResponseWriter, db *sql.DB, userID int, userType string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	// Store session in the database with user_type
	_, err = db.Exec(`
		INSERT INTO sessions (session_id, user_id, user_type, expires_at)
		VALUES ($1, $2, $3, $4)`,
		sessionID, userID, userType, expiresAt)
	if err != nil {
		return "", err
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
	})

	return sessionID, nil
}

// GetSessionID retrieves the session ID from the request cookies.
func GetSessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", errors.New("session not found")
	}
	return cookie.Value, nil
}

// GetUserIDFromSession retrieves the user ID and type associated with the session ID.
func GetUserIDFromSession(db *sql.DB, sessionID string) (int, string, error) {
	var userID int
	var userType string
	var expiresAt time.Time

	err := db.QueryRow(`
		SELECT user_id, user_type, expires_at FROM sessions WHERE session_id = $1`, sessionID).
		Scan(&userID, &userType, &expiresAt)
	if err != nil {
		return 0, "", errors.New("invalid session")
	}

	// Check if the session has expired
	if time.Now().After(expiresAt) {
		// Optionally, delete the expired session from the database
		_, _ = db.Exec(`DELETE FROM sessions WHERE session_id = $1`, sessionID)
		return 0, "", errors.New("session expired")
	}

	return userID, userType, nil
}

// DestroySession removes the session from the database.
func DestroySession(db *sql.DB, sessionID string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE session_id = $1`, sessionID)
	return err
}

// generateSessionID generates a secure random session ID.
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
