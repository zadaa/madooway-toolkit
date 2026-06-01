package handlers

import (
	"encoding/json"
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

type TicketMessageJSON struct {
	ID              int    `json:"id"`
	TicketID        int    `json:"ticket_id"`
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	Message         string `json:"message"`
	FilePath        string `json:"file_path"`
	FileName        string `json:"file_name"`
	PrettyCreatedAt string `json:"pretty_created_at"`
}

// GetTicketMessagesJSON returns thread messages for a ticket in JSON format
func GetTicketMessagesJSON(w http.ResponseWriter, r *http.Request) {
	ticketIDStr := r.URL.Query().Get("ticket_id")
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		http.Error(w, "ID tiket tidak valid", http.StatusBadRequest)
		return
	}

	messages, err := models.GetTicketMessagesByTicketID(ticketID)
	if err != nil {
		log.Printf("Error fetching ticket messages: %v", err)
		http.Error(w, "Gagal memuat pesan diskusi", http.StatusInternalServerError)
		return
	}

	var response []TicketMessageJSON
	for _, m := range messages {
		filename := ""
		if m.FilePath != "" {
			filename = m.FilePath[strings.LastIndex(m.FilePath, "/")+1:]
		}

		response = append(response, TicketMessageJSON{
			ID:              m.ID,
			TicketID:        m.TicketID,
			UserID:          m.UserID,
			Username:        m.Username,
			Message:         m.Message,
			FilePath:        m.FilePath,
			FileName:        filename,
			PrettyCreatedAt: m.PrettyCreatedAt(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateTicketMessageAJAX handles adding a message via AJAX
func CreateTicketMessageAJAX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(5 << 20) // 5MB limit
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
	}

	userID, _ := middleware.GetSessionUser(r)
	ticketIDStr := strings.TrimSpace(r.FormValue("ticket_id"))
	message := strings.TrimSpace(r.FormValue("message"))

	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "ID tiket tidak valid"})
		return
	}

	// Verify at least message or file is present
	file, fileHeader, fileErr := r.FormFile("upload_file")
	if message == "" && fileErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Pesan atau file lampiran wajib diisi"})
		return
	}

	var filePath string
	if fileErr == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		
		newFilename := fmt.Sprintf("thread_%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("static", "uploads", "tickets", newFilename)

		dst, err := os.Create(savePath)
		if err != nil {
			log.Printf("Error creating thread upload file: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Gagal menyimpan file lampiran"})
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Printf("Error copying thread upload: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Gagal menyimpan file lampiran"})
			return
		}
		filePath = "/static/uploads/tickets/" + newFilename
	}

	err = models.CreateTicketMessage(ticketID, userID, message, filePath)
	if err != nil {
		log.Printf("Error inserting ticket message: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Gagal mengirimkan pesan"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
