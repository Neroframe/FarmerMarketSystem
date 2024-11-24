package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
)

func CreateSession(w http.ResponseWriter, db *sql.DB, userID int, userType string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = db.Exec(`
		INSERT INTO sessions (session_id, user_id, user_type, expires_at)
		VALUES ($1, $2, $3, $4)`,
		sessionID, userID, userType, expiresAt)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: expiresAt,
		Path:    "/",
		// Production values
		HttpOnly: true,
		Secure:   true,
		// SameSite: http.SameSiteStrictMode, // Prevents CSRF attacks
		SameSite: http.SameSiteNoneMode, // Allows cross-site cookie
	})

	return sessionID, nil
}

func GetSessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", errors.New("session not found")
	}
	return cookie.Value, nil
}

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

	// Check and delete the expired session from the database
	if time.Now().After(expiresAt) {
		DestroySession(db, sessionID)
		return 0, "", errors.New("session expired")
	}

	return userID, userType, nil
}

func DestroySession(db *sql.DB, sessionID string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE session_id = $1`, sessionID)
	return err
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
