package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

type TicketsPageData struct {
	Tickets []models.Ticket
	Clients []models.Client
}

// ListTickets displays the list of tickets
func ListTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := models.GetAllTickets("", "", "", "", "", "")
	if err != nil {
		log.Printf("Error fetching tickets: %v", err)
		RenderTemplate(w, r, "tickets.html", "Daftar Tiket", "tickets", nil, "Gagal memuat daftar tiket.", "")
		return
	}

	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients for tickets: %v", err)
		clients = []models.Client{}
	}

	data := TicketsPageData{
		Tickets: tickets,
		Clients: clients,
	}

	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	RenderTemplate(w, r, "tickets.html", "Daftar Tiket", "tickets", data, errorMsg, successMsg)
}

// CreateTicket processes ticket creation with optional file upload
func CreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tickets", http.StatusSeeOther)
		return
	}

	err := r.ParseMultipartForm(5 << 20) // 5MB limit
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
	}

	userID, _ := middleware.GetSessionUser(r)
	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	issueDateStr := strings.TrimSpace(r.FormValue("issue_date"))
	category := strings.TrimSpace(r.FormValue("category"))
	ticketLink := strings.TrimSpace(r.FormValue("ticket_link"))
	status := strings.TrimSpace(r.FormValue("status"))
	finishedDateStr := strings.TrimSpace(r.FormValue("finished_date"))

	if title == "" || clientIDStr == "" || category == "" {
		http.Redirect(w, r, "/tickets?error=Klien,+Judul,+dan+Kategori+wajib+diisi", http.StatusSeeOther)
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Redirect(w, r, "/tickets?error=Klien+tidak+valid", http.StatusSeeOther)
		return
	}

	if status == "" {
		status = "Pending"
	}

	var filePath string
	file, fileHeader, err := r.FormFile("upload_file")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		
		// Set safe filename using timestamp
		newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("static", "uploads", "tickets", newFilename)

		dst, err := os.Create(savePath)
		if err != nil {
			log.Printf("Error creating upload file: %v", err)
			http.Redirect(w, r, "/tickets?error=Gagal+menyimpan+file+lampiran", http.StatusSeeOther)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Printf("Error copying upload content: %v", err)
			http.Redirect(w, r, "/tickets?error=Gagal+menyimpan+file+lampiran", http.StatusSeeOther)
			return
		}
		filePath = "/static/uploads/tickets/" + newFilename
	}

	err = models.CreateTicket(clientID, title, description, userID, filePath, issueDateStr, category, ticketLink, status, finishedDateStr)
	if err != nil {
		log.Printf("Error creating ticket: %v", err)
		http.Redirect(w, r, "/tickets?error=Gagal+menambahkan+tiket", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/tickets?success=Tiket+berhasil+ditambahkan", http.StatusSeeOther)
}

// UpdateTicket processes ticket edits
func UpdateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tickets", http.StatusSeeOther)
		return
	}

	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/tickets?error=ID+tiket+tidak+valid", http.StatusSeeOther)
		return
	}

	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	issueDateStr := strings.TrimSpace(r.FormValue("issue_date"))
	category := strings.TrimSpace(r.FormValue("category"))
	ticketLink := strings.TrimSpace(r.FormValue("ticket_link"))
	status := strings.TrimSpace(r.FormValue("status"))
	finishedDateStr := strings.TrimSpace(r.FormValue("finished_date"))

	if title == "" || clientIDStr == "" || category == "" {
		http.Redirect(w, r, "/tickets?error=Klien,+Judul,+dan+Kategori+wajib+diisi", http.StatusSeeOther)
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Redirect(w, r, "/tickets?error=Klien+tidak+valid", http.StatusSeeOther)
		return
	}

	// Fetch existing ticket to keep old file path if no new file is uploaded
	existingTicket, err := models.GetTicketByID(id)
	if err != nil || existingTicket == nil {
		http.Redirect(w, r, "/tickets?error=Tiket+tidak+ditemukan", http.StatusSeeOther)
		return
	}
	filePath := existingTicket.FilePath

	file, fileHeader, err := r.FormFile("upload_file")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("static", "uploads", "tickets", newFilename)

		dst, err := os.Create(savePath)
		if err != nil {
			log.Printf("Error creating upload file: %v", err)
			http.Redirect(w, r, "/tickets?error=Gagal+menyimpan+file+lampiran", http.StatusSeeOther)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Printf("Error copying upload content: %v", err)
			http.Redirect(w, r, "/tickets?error=Gagal+menyimpan+file+lampiran", http.StatusSeeOther)
			return
		}
		
		// Delete old file if it exists
		if existingTicket.FilePath != "" {
			oldPath := strings.TrimPrefix(existingTicket.FilePath, "/")
			_ = os.Remove(oldPath)
		}

		filePath = "/static/uploads/tickets/" + newFilename
	}

	err = models.UpdateTicket(id, clientID, title, description, filePath, issueDateStr, category, ticketLink, status, finishedDateStr)
	if err != nil {
		log.Printf("Error updating ticket: %v", err)
		http.Redirect(w, r, "/tickets?error=Gagal+memperbarui+tiket", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/tickets?success=Tiket+berhasil+diperbarui", http.StatusSeeOther)
}

// DeleteTicket deletes a ticket and its uploaded file
func DeleteTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tickets", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/tickets?error=ID+tiket+tidak+valid", http.StatusSeeOther)
		return
	}

	ticket, err := models.GetTicketByID(id)
	if err != nil || ticket == nil {
		http.Redirect(w, r, "/tickets?error=Tiket+tidak+ditemukan", http.StatusSeeOther)
		return
	}

	err = models.DeleteTicket(id)
	if err != nil {
		log.Printf("Error deleting ticket: %v", err)
		http.Redirect(w, r, "/tickets?error=Gagal+menghapus+tiket", http.StatusSeeOther)
		return
	}

	// Delete file from disk if it exists
	if ticket.FilePath != "" {
		filePathOnDisk := strings.TrimPrefix(ticket.FilePath, "/")
		err = os.Remove(filePathOnDisk)
		if err != nil {
			log.Printf("Warning: failed to delete ticket file from disk: %v", err)
		}
	}

	http.Redirect(w, r, "/tickets?success=Tiket+berhasil+dihapus", http.StatusSeeOther)
}
