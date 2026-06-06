package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"task-manager-go/config"
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

	err := r.ParseMultipartForm(5 << 20) // 5MB limit
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
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

	var logoPath string
	file, fileHeader, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		allowedExts := map[string]bool{
			".png": true, ".jpg": true, ".jpeg": true, ".svg": true, ".gif": true, ".webp": true,
		}
		if !allowedExts[ext] {
			http.Redirect(w, r, "/clients?error=Format+logo+tidak+didukung.+Gunakan+PNG,+JPG,+JPEG,+SVG,+GIF,+atau+WEBP", http.StatusSeeOther)
			return
		}

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Error reading logo file: %v", err)
			http.Redirect(w, r, "/clients?error=Gagal+membaca+file+logo", http.StatusSeeOther)
			return
		}

		mimeType := "image/png"
		switch ext {
		case ".png":
			mimeType = "image/png"
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".gif":
			mimeType = "image/gif"
		case ".svg":
			mimeType = "image/svg+xml"
		case ".webp":
			mimeType = "image/webp"
		}

		encoded := base64.StdEncoding.EncodeToString(fileBytes)
		logoPath = fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
	}

	province := strings.TrimSpace(r.FormValue("province"))
	if province == "" {
		province = "DKI Jakarta"
	}

	err = models.CreateClient(name, shortName, email, phone, picName, pricePackage, logoPath, province)
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

	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
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

	// Fetch existing client to preserve old logo if no new file is uploaded
	existingClient, err := models.GetClientByID(id)
	if err != nil || existingClient == nil {
		http.Redirect(w, r, "/clients?error=Klien+tidak+ditemukan", http.StatusSeeOther)
		return
	}
	logoPath := existingClient.Logo

	file, fileHeader, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		allowedExts := map[string]bool{
			".png": true, ".jpg": true, ".jpeg": true, ".svg": true, ".gif": true, ".webp": true,
		}
		if !allowedExts[ext] {
			http.Redirect(w, r, "/clients?error=Format+logo+tidak+didukung.+Gunakan+PNG,+JPG,+JPEG,+SVG,+GIF,+atau+WEBP", http.StatusSeeOther)
			return
		}

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Error reading logo file: %v", err)
			http.Redirect(w, r, "/clients?error=Gagal+membaca+file+logo", http.StatusSeeOther)
			return
		}

		mimeType := "image/png"
		switch ext {
		case ".png":
			mimeType = "image/png"
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".gif":
			mimeType = "image/gif"
		case ".svg":
			mimeType = "image/svg+xml"
		case ".webp":
			mimeType = "image/webp"
		}

		encoded := base64.StdEncoding.EncodeToString(fileBytes)
		logoPath = fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
	}

	province := strings.TrimSpace(r.FormValue("province"))
	if province == "" {
		province = "DKI Jakarta"
	}

	err = models.UpdateClient(id, name, shortName, email, phone, picName, pricePackage, logoPath, province)
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

// SyncClients handles the database synchronization of clients
func SyncClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/clients", http.StatusSeeOther)
		return
	}

	cfg := config.AppConfig
	targetDB := "db_madoo_ms_presention"

	count, err := models.SyncClientsFromLocalDB(cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, targetDB)
	if err != nil {
		log.Printf("Error syncing clients from local DB: %v", err)
		errStr := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, fmt.Sprintf("/clients?error=Gagal+sinkronisasi+klien:+%s", errStr), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/clients?success=Berhasil+sinkronisasi+%d+klien", count), http.StatusSeeOther)
}

