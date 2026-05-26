package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

type DashboardData struct {
	KPI           *models.KPIStats
	StatusJSON    template.JS
	CategoryJSON  template.JS
	DateJSON      template.JS
	Clients       []models.Client
	TasksJSON     template.JS
	SchedulesJSON template.JS
	ProvinceJSON  template.JS
	ClientsJSON   template.JS
}

// ShowDashboard gathers stats and renders the main dashboard page
func ShowDashboard(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetSessionUser(r)

	kpi, err := models.GetKPIStats(userID)
	if err != nil {
		log.Printf("Error fetching KPI stats: %v", err)
		kpi = &models.KPIStats{}
	}

	statusStats, err := models.GetTaskStatsByStatus(userID)
	if err != nil {
		log.Printf("Error fetching status stats: %v", err)
		statusStats = map[string]int{"Pending": 0, "In Progress": 0, "Completed": 0}
	}

	categoryStats, err := models.GetTaskStatsByCategory(userID)
	if err != nil {
		log.Printf("Error fetching category stats: %v", err)
		categoryStats = make(map[string]int)
	}

	dateStats, err := models.GetTaskStatsByDate(userID)
	if err != nil {
		log.Printf("Error fetching date stats: %v", err)
		dateStats = []models.DateStat{}
	}

	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients for dashboard: %v", err)
		clients = []models.Client{}
	}

	tasks, err := models.GetTasksByUserID(userID, "", "", "")
	if err != nil {
		log.Printf("Error fetching tasks for dashboard: %v", err)
		tasks = []models.Task{}
	}

	schedules, err := models.GetTrainingSchedulesByUserID(userID)
	if err != nil {
		log.Printf("Error fetching schedules for dashboard: %v", err)
		schedules = []models.TrainingSchedule{}
	}

	// Marshal into JSON strings to inject in templates safely as JS variables
	statusJSON, err := json.Marshal(statusStats)
	if err != nil {
		statusJSON = []byte("{}")
	}

	categoryJSON, err := json.Marshal(categoryStats)
	if err != nil {
		categoryJSON = []byte("{}")
	}

	dateJSON, err := json.Marshal(dateStats)
	if err != nil {
		dateJSON = []byte("[]")
	}

	tasksJSON, err := json.Marshal(tasks)
	if err != nil {
		tasksJSON = []byte("[]")
	}

	schedulesJSON, err := json.Marshal(schedules)
	if err != nil {
		schedulesJSON = []byte("[]")
	}

	provinceStats, err := models.GetClientStatsByProvince()
	if err != nil {
		log.Printf("Error fetching client province stats: %v", err)
		provinceStats = []models.ProvinceStat{}
	}
	provinceJSON, err := json.Marshal(provinceStats)
	if err != nil {
		provinceJSON = []byte("[]")
	}

	clientsJSON, err := json.Marshal(clients)
	if err != nil {
		clientsJSON = []byte("[]")
	}

	data := DashboardData{
		KPI:           kpi,
		StatusJSON:    template.JS(statusJSON),
		CategoryJSON:  template.JS(categoryJSON),
		DateJSON:      template.JS(dateJSON),
		Clients:       clients,
		TasksJSON:     template.JS(tasksJSON),
		SchedulesJSON: template.JS(schedulesJSON),
		ProvinceJSON:  template.JS(provinceJSON),
		ClientsJSON:   template.JS(clientsJSON),
	}

	RenderTemplate(w, r, "dashboard.html", "Dashboard Analytics", "dashboard", data, "", "")
}
