package handlers

import (
	"net/http"
	"strings"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

// ShowLogin renders the login page
func ShowLogin(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "login.html", "Login", "", nil, "", "")
}

// HandleLogin processes login credentials
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		RenderTemplate(w, r, "login.html", "Login", "", nil, "Harap isi semua kolom.", "")
		return
	}

	user, err := models.GetUserByIdentifier(email)
	if err != nil {
		RenderTemplate(w, r, "login.html", "Login", "", nil, "Terjadi kesalahan pada server.", "")
		return
	}

	if user == nil || !models.CheckPasswordHash(password, user.PasswordHash) {
		RenderTemplate(w, r, "login.html", "Login", "", nil, "Email atau password salah.", "")
		return
	}

	// Create session
	middleware.CreateSession(w, user.ID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ShowRegister renders the registration page
func ShowRegister(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "register.html", "Daftar Akun", "", nil, "", "")
}

// HandleRegister processes user registration
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		RenderTemplate(w, r, "register.html", "Daftar Akun", "", nil, "Harap isi semua kolom.", "")
		return
	}

	if len(password) < 6 {
		RenderTemplate(w, r, "register.html", "Daftar Akun", "", nil, "Password minimal harus 6 karakter.", "")
		return
	}

	// Check if email already registered
	existingUser, err := models.GetUserByEmail(email)
	if err != nil {
		RenderTemplate(w, r, "register.html", "Daftar Akun", "", nil, "Terjadi kesalahan pada server.", "")
		return
	}
	if existingUser != nil {
		RenderTemplate(w, r, "register.html", "Daftar Akun", "", nil, "Email sudah terdaftar.", "")
		return
	}

	// Create user
	err = models.CreateUser(username, email, password, "", "user", "")
	if err != nil {
		RenderTemplate(w, r, "register.html", "Daftar Akun", "", nil, "Gagal mendaftarkan akun. Silakan coba lagi.", "")
		return
	}

	// Redirect to login page with success message
	RenderTemplate(w, r, "login.html", "Login", "", nil, "", "Pendaftaran berhasil! Silakan masuk.")
}

// HandleLogout terminates the user session
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	middleware.DestroySession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
