package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"task-manager-go/models"
)

// ListClients displays all master clients
func ListClients(w http.ResponseWriter, r *http.Request) {
	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients: %v", err)
		RenderTemplate(w, r, "clients.html", "Master Klien", "clients", nil, "Gagal memuat daftar klien.", "")
		return
	}

	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	RenderTemplate(w, r, "clients.html", "Master Klien", "clients", clients, errorMsg, successMsg)
}

// CreateClient processes new client creation
func CreateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/clients", http.StatusSeeOther)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	shortName := strings.TrimSpace(r.FormValue("short_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	picName := strings.TrimSpace(r.FormValue("pic_name"))
	pricePackage := strings.TrimSpace(r.FormValue("price_package"))

	if name == "" || pricePackage == "" {
		http.Redirect(w, r, "/clients?error=Nama+dan+Paket+Harga+wajib+diisi", http.StatusSeeOther)
		return
	}

	err := models.CreateClient(name, shortName, email, phone, picName, pricePackage)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		http.Redirect(w, r, "/clients?error=Gagal+menambahkan+klien", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/clients?success=Klien+berhasil+ditambahkan", http.StatusSeeOther)
}

// UpdateClient processes client updates
func UpdateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/clients", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/clients?error=ID+klien+tidak+valid", http.StatusSeeOther)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	shortName := strings.TrimSpace(r.FormValue("short_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	picName := strings.TrimSpace(r.FormValue("pic_name"))
	pricePackage := strings.TrimSpace(r.FormValue("price_package"))

	if name == "" || pricePackage == "" {
		http.Redirect(w, r, "/clients?error=Nama+dan+Paket+Harga+wajib+diisi", http.StatusSeeOther)
		return
	}

	err = models.UpdateClient(id, name, shortName, email, phone, picName, pricePackage)
	if err != nil {
		log.Printf("Error updating client: %v", err)
		http.Redirect(w, r, "/clients?error=Gagal+memperbarui+klien", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/clients?success=Klien+berhasil+diperbarui", http.StatusSeeOther)
}

// DeleteClient deletes a client
func DeleteClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/clients", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/clients?error=ID+klien+tidak+valid", http.StatusSeeOther)
		return
	}

	err = models.DeleteClient(id)
	if err != nil {
		log.Printf("Error deleting client: %v", err)
		http.Redirect(w, r, "/clients?error=Gagal+menghapus+klien", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/clients?success=Klien+berhasil+dihapus", http.StatusSeeOther)
}
