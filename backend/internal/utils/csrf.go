package utils

import (
    "crypto/rand"
    "encoding/base64"
    "errors"
    "net/http"
    "time"
)

// SetCSRFToken generates a CSRF token, sets it as a cookie, and returns the token.
func SetCSRFToken(w http.ResponseWriter) (string, error) {
    tokenBytes := make([]byte, 32)
    _, err := rand.Read(tokenBytes)
    if err != nil {
        return "", err
    }
    token := base64.StdEncoding.EncodeToString(tokenBytes)

    cookie := &http.Cookie{
        Name:     "csrf_token",
        Value:    token,
        Path:     "/",
        Expires:  time.Now().Add(24 * time.Hour),
        HttpOnly: true,                      // Prevents JavaScript access
        Secure:   false,                     // Set to true in production with HTTPS
        SameSite: http.SameSiteStrictMode,   // Prevents CSRF attacks
    }
    http.SetCookie(w, cookie)
    return token, nil
}

// ValidateCSRFToken compares the CSRF token from the form with the token in the cookie.
func ValidateCSRFToken(r *http.Request) error {
    // Get CSRF token from form
    formToken := r.FormValue("csrf_token")
    if formToken == "" {
        return errors.New("csrf token not provided")
    }

    // Get CSRF token from cookie
    cookie, err := r.Cookie("csrf_token")
    if err != nil {
        return errors.New("csrf token cookie not found")
    }

    // Compare tokens
    if formToken != cookie.Value {
        return errors.New("invalid csrf token")
    }

    return nil
}
