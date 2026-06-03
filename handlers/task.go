package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

type TasksPageData struct {
	Tasks          []models.Task
	Clients        []models.Client
	Users          []models.User
	SearchQuery    string
	ActiveCategory string
	ActiveStatus   string
}

// ListTasks handles showing all tasks with optional filters
func ListTasks(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetSessionUser(r)

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	tasks, err := models.GetTasksByUserID(userID, "", "", "")
	if err != nil {
		log.Printf("Error fetching tasks: %v", err)
		RenderTemplate(w, r, "tasks.html", "Daftar Tugas", "tasks", nil, "Gagal memuat daftar tugas.", "")
		return
	}

	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients for tasks: %v", err)
		clients = []models.Client{}
	}

	users, err := models.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users for tasks: %v", err)
		users = []models.User{}
	}

	data := TasksPageData{
		Tasks:          tasks,
		Clients:        clients,
		Users:          users,
		SearchQuery:    search,
		ActiveCategory: category,
		ActiveStatus:   status,
	}

	// Read temporary success/error messages if any from url query to act as simple flash
	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	RenderTemplate(w, r, "tasks.html", "Daftar Tugas", "tasks", data, errorMsg, successMsg)
}

// CreateTask handles task creation
func CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	category := strings.TrimSpace(r.FormValue("category"))
	source := strings.TrimSpace(r.FormValue("source"))
	status := strings.TrimSpace(r.FormValue("status"))
	dueDateStr := strings.TrimSpace(r.FormValue("due_date"))
	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))

	if title == "" || category == "" {
		http.Redirect(w, r, "/tasks?error=Judul+dan+Kategori+wajib+diisi", http.StatusSeeOther)
		return
	}

	if source == "" {
		source = "WA Supp" // default value
	}
	if status == "" {
		status = "Pending"
	}

	var clientID sql.NullInt64
	if clientIDStr != "" {
		cID, err := strconv.Atoi(clientIDStr)
		if err == nil {
			clientID = sql.NullInt64{Int64: int64(cID), Valid: true}
		}
	}

	err := models.CreateTask(userID, title, description, category, source, status, dueDateStr, clientID)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		http.Redirect(w, r, "/tasks?error=Gagal+membuat+tugas", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/tasks?success=Tugas+berhasil+ditambahkan", http.StatusSeeOther)
}

// UpdateTask handles task updating
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/tasks?error=ID+tugas+tidak+valid", http.StatusSeeOther)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	category := strings.TrimSpace(r.FormValue("category"))
	source := strings.TrimSpace(r.FormValue("source"))
	status := strings.TrimSpace(r.FormValue("status"))
	dueDateStr := strings.TrimSpace(r.FormValue("due_date"))
	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))

	if title == "" || category == "" {
		http.Redirect(w, r, "/tasks?error=Judul+dan+Kategori+wajib+diisi", http.StatusSeeOther)
		return
	}

	if source == "" {
		source = "WA Supp"
	}
	if status == "" {
		status = "Pending"
	}

	var clientID sql.NullInt64
	if clientIDStr != "" {
		cID, err := strconv.Atoi(clientIDStr)
		if err == nil {
			clientID = sql.NullInt64{Int64: int64(cID), Valid: true}
		}
	}

	err = models.UpdateTask(id, userID, title, description, category, source, status, dueDateStr, clientID)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		http.Redirect(w, r, "/tasks?error=Gagal+memperbarui+tugas", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/tasks?success=Tugas+berhasil+diperbarui", http.StatusSeeOther)
}

// DeleteTask handles task deletion
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/tasks?error=ID+tugas+tidak+valid", http.StatusSeeOther)
		return
	}

	err = models.DeleteTask(id, userID)
	if err != nil {
		log.Printf("Error deleting task: %v", err)
		http.Redirect(w, r, "/tasks?error=Gagal+menghapus+tugas", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/tasks?success=Tugas+berhasil+dihapus", http.StatusSeeOther)
}
