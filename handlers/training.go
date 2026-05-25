package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

type TrainingsPageData struct {
	Schedules []models.TrainingSchedule
	Clients   []models.Client
}

// ListTrainings renders the list of training schedules
func ListTrainings(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetSessionUser(r)

	schedules, err := models.GetTrainingSchedulesByUserID(userID)
	if err != nil {
		log.Printf("Error fetching training schedules: %v", err)
		RenderTemplate(w, r, "trainings.html", "Jadwal Training", "trainings", nil, "Gagal memuat jadwal training.", "")
		return
	}

	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients for trainings: %v", err)
		clients = []models.Client{}
	}

	data := TrainingsPageData{
		Schedules: schedules,
		Clients:   clients,
	}

	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	RenderTemplate(w, r, "trainings.html", "Jadwal Training", "trainings", data, errorMsg, successMsg)
}

// CreateTraining processes training schedule creation
func CreateTraining(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/trainings", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	trainingDateStr := strings.TrimSpace(r.FormValue("training_date"))
	location := strings.TrimSpace(r.FormValue("location"))
	trainer := strings.TrimSpace(r.FormValue("trainer"))
	trainingType := strings.TrimSpace(r.FormValue("training_type"))
	status := strings.TrimSpace(r.FormValue("status"))

	if clientIDStr == "" || title == "" || trainingDateStr == "" {
		http.Redirect(w, r, "/trainings?error=Klien,+Judul,+dan+Tanggal+wajib+diisi", http.StatusSeeOther)
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Redirect(w, r, "/trainings?error=ID+Klien+tidak+valid", http.StatusSeeOther)
		return
	}

	if trainingType == "" {
		trainingType = "Online"
	}
	if status == "" {
		status = "Scheduled"
	}

	err = models.CreateTrainingSchedule(userID, clientID, title, description, trainingDateStr, location, trainer, trainingType, status)
	if err != nil {
		log.Printf("Error creating training schedule: %v", err)
		http.Redirect(w, r, "/trainings?error=Gagal+membuat+jadwal+training", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/trainings?success=Jadwal+training+berhasil+ditambahkan", http.StatusSeeOther)
}

// UpdateTraining processes training schedule updates
func UpdateTraining(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/trainings", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/trainings?error=ID+jadwal+tidak+valid", http.StatusSeeOther)
		return
	}

	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	trainingDateStr := strings.TrimSpace(r.FormValue("training_date"))
	location := strings.TrimSpace(r.FormValue("location"))
	trainer := strings.TrimSpace(r.FormValue("trainer"))
	trainingType := strings.TrimSpace(r.FormValue("training_type"))
	status := strings.TrimSpace(r.FormValue("status"))

	if clientIDStr == "" || title == "" || trainingDateStr == "" {
		http.Redirect(w, r, "/trainings?error=Klien,+Judul,+dan+Tanggal+wajib+diisi", http.StatusSeeOther)
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Redirect(w, r, "/trainings?error=ID+Klien+tidak+valid", http.StatusSeeOther)
		return
	}

	if trainingType == "" {
		trainingType = "Online"
	}
	if status == "" {
		status = "Scheduled"
	}

	err = models.UpdateTrainingSchedule(id, userID, clientID, title, description, trainingDateStr, location, trainer, trainingType, status)
	if err != nil {
		log.Printf("Error updating training schedule: %v", err)
		http.Redirect(w, r, "/trainings?error=Gagal+memperbarui+jadwal+training", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/trainings?success=Jadwal+training+berhasil+diperbarui", http.StatusSeeOther)
}

// DeleteTraining deletes a training schedule
func DeleteTraining(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/trainings", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/trainings?error=ID+jadwal+tidak+valid", http.StatusSeeOther)
		return
	}

	err = models.DeleteTrainingSchedule(id, userID)
	if err != nil {
		log.Printf("Error deleting training schedule: %v", err)
		http.Redirect(w, r, "/trainings?error=Gagal+menghapus+jadwal+training", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/trainings?success=Jadwal+training+berhasil+dihapus", http.StatusSeeOther)
}


