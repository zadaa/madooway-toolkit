package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"task-manager-go/db"
)

// GenerateSessionToken creates a cryptographically secure random token
func GenerateSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CreateSession registers a new session and sets the HTTP-only cookie
func CreateSession(w http.ResponseWriter, userID int) {
	token := GenerateSessionToken()
	expiresAt := time.Now().Add(365 * 24 * time.Hour) // 1 year expiry

	_, err := db.DB.Exec("INSERT INTO user_sessions (token, user_id, expires_at) VALUES (?, ?, ?)", token, userID, expiresAt)
	if err != nil {
		log.Printf("Error creating session in database: %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		// SameSite: http.SameSiteLaxMode, // Good practice
	})
}

// GetSessionUser extracts the user ID from the session cookie
func GetSessionUser(r *http.Request) (int, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, false
	}

	var userID int
	err = db.DB.QueryRow("SELECT user_id FROM user_sessions WHERE token = ? AND expires_at > NOW()", cookie.Value).Scan(&userID)
	if err != nil {
		return 0, false
	}

	return userID, true
}

// DestroySession removes the session and clears the cookie
func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		_, err = db.DB.Exec("DELETE FROM user_sessions WHERE token = ?", cookie.Value)
		if err != nil {
			log.Printf("Error destroying session in database: %v", err)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
}

// AuthRequired ensures the user is logged in, redirecting to /login otherwise
func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, loggedIn := GetSessionUser(r)
		if !loggedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// GuestOnly ensures the user is NOT logged in, redirecting to /dashboard otherwise
func GuestOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, loggedIn := GetSessionUser(r)
		if loggedIn {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// AdminOnly ensures the user is logged in as an admin
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, loggedIn := GetSessionUser(r)
		if !loggedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		var role string
		err := db.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
		if err != nil || role != "admin" {
			// Redirect non-admins to /tasks (the main toolkit page)
			http.Redirect(w, r, "/tasks", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
