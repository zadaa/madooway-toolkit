package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

var (
	// Sessions map: sessionToken -> UserID
	sessions   = make(map[string]int)
	sessionsMu sync.RWMutex
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

	sessionsMu.Lock()
	sessions[token] = userID
	sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
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

	sessionsMu.RLock()
	userID, exists := sessions[cookie.Value]
	sessionsMu.RUnlock()

	return userID, exists
}

// DestroySession removes the session and clears the cookie
func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
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
