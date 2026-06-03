package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

// ListUsers displays all users
func ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := models.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		RenderTemplate(w, r, "users.html", "Manajemen User", "users", nil, "Gagal memuat daftar user.", "")
		return
	}

	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	RenderTemplate(w, r, "users.html", "Manajemen User", "users", users, errorMsg, successMsg)
}

// CreateUser handles user addition from user management screen
func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	color := strings.TrimSpace(r.FormValue("color"))
	role := strings.TrimSpace(r.FormValue("role"))
	nip := strings.TrimSpace(r.FormValue("nip"))

	if username == "" || email == "" || password == "" {
		http.Redirect(w, r, "/users?error=Username, email, dan password wajib diisi", http.StatusSeeOther)
		return
	}

	if len(password) < 6 {
		http.Redirect(w, r, "/users?error=Password+minimal+harus+6+karakter", http.StatusSeeOther)
		return
	}

	// Check if username already exists
	usernameExists, err := models.CheckUsernameExists(username, 0)
	if err != nil {
		log.Printf("Error checking username: %v", err)
		http.Redirect(w, r, "/users?error=Terjadi+kesalahan+pada+server", http.StatusSeeOther)
		return
	}
	if usernameExists {
		http.Redirect(w, r, "/users?error=Username+sudah+digunakan", http.StatusSeeOther)
		return
	}

	// Check if email already exists
	emailExists, err := models.CheckEmailExists(email, 0)
	if err != nil {
		log.Printf("Error checking email: %v", err)
		http.Redirect(w, r, "/users?error=Terjadi+kesalahan+pada+server", http.StatusSeeOther)
		return
	}
	if emailExists {
		http.Redirect(w, r, "/users?error=Email+sudah+terdaftar", http.StatusSeeOther)
		return
	}
	// Check if NIP already exists
	nipExists, err := models.CheckNIPExists(nip, 0)
	if err != nil {
		log.Printf("Error checking NIP: %v", err)
		http.Redirect(w, r, "/users?error=Terjadi+kesalahan+pada+server", http.StatusSeeOther)
		return
	}
	if nipExists {
		http.Redirect(w, r, "/users?error=NIP+sudah+digunakan", http.StatusSeeOther)
		return
	}

	err = models.CreateUser(username, email, password, color, role, nip)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Redirect(w, r, "/users?error=Gagal+menambahkan+user", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/users?success=User+berhasil+ditambahkan", http.StatusSeeOther)
}

// UpdateUser handles user profile updates
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/users?error=ID+user+tidak+valid", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password") // optional
	color := strings.TrimSpace(r.FormValue("color"))
	role := strings.TrimSpace(r.FormValue("role"))
	nip := strings.TrimSpace(r.FormValue("nip"))

	if username == "" || email == "" {
		http.Redirect(w, r, "/users?error=Username+dan+email+wajib+diisi", http.StatusSeeOther)
		return
	}

	if password != "" && len(password) < 6 {
		http.Redirect(w, r, "/users?error=Password+minimal+harus+6+karakter", http.StatusSeeOther)
		return
	}

	// Check if username already exists for other users
	usernameExists, err := models.CheckUsernameExists(username, id)
	if err != nil {
		log.Printf("Error checking username: %v", err)
		http.Redirect(w, r, "/users?error=Terjadi+kesalahan+pada+server", http.StatusSeeOther)
		return
	}
	if usernameExists {
		http.Redirect(w, r, "/users?error=Username+sudah+digunakan+oleh+user+lain", http.StatusSeeOther)
		return
	}

	// Check if email already exists for other users
	emailExists, err := models.CheckEmailExists(email, id)
	if err != nil {
		log.Printf("Error checking email: %v", err)
		http.Redirect(w, r, "/users?error=Terjadi+kesalahan+pada+server", http.StatusSeeOther)
		return
	}
	if emailExists {
		http.Redirect(w, r, "/users?error=Email+sudah+terdaftar+oleh+user+lain", http.StatusSeeOther)
		return
	}
	// Check if NIP already exists for other users
	nipExists, err := models.CheckNIPExists(nip, id)
	if err != nil {
		log.Printf("Error checking NIP: %v", err)
		http.Redirect(w, r, "/users?error=Terjadi+kesalahan+pada+server", http.StatusSeeOther)
		return
	}
	if nipExists {
		http.Redirect(w, r, "/users?error=NIP+sudah+digunakan+oleh+user+lain", http.StatusSeeOther)
		return
	}

	err = models.UpdateUser(id, username, email, password, color, role, nip)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Redirect(w, r, "/users?error=Gagal+memperbarui+user", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/users?success=User+berhasil+diperbarui", http.StatusSeeOther)
}

// DeleteUser handles user deletion, blocking active user self-deletion
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/users?error=ID+user+tidak+valid", http.StatusSeeOther)
		return
	}

	currentUserID, _ := middleware.GetSessionUser(r)
	if id == currentUserID {
		http.Redirect(w, r, "/users?error=Anda+tidak+dapat+menghapus+akun+aktif+Anda+sendiri.", http.StatusSeeOther)
		return
	}

	err = models.DeleteUser(id)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Redirect(w, r, "/users?error=Gagal+menghapus+user", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/users?success=User+berhasil+dihapus", http.StatusSeeOther)
}
