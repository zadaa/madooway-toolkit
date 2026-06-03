package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

type KPIDashboardData struct {
	Users          []models.User
	Tasks          []models.Task
	SelectedUserID int
	SelectedUser   *models.User
	StartDate      string
	EndDate        string
	
	// Metrics
	TotalTasks     int
	PendingTasks   int
	ProgressTasks  int
	CompletedTasks int
	CompletionRate float64
	TrainingCount  int
	LeadsCount     int

	// JSON format for Chart.js
	StatusJSON template.JS
}

// ShowUserKPI handles requests to the user KPI dashboard
func ShowUserKPI(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := middleware.GetSessionUser(r)

	// Fetch all users for filter dropdown
	users, err := models.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users for KPI: %v", err)
		users = []models.User{}
	}

	// Determine selected user
	userIDStr := r.URL.Query().Get("user_id")
	selectedUserID := currentUserID
	if userIDStr != "" {
		if id, err := strconv.Atoi(userIDStr); err == nil {
			selectedUserID = id
		}
	}

	// Fetch selected user details
	selectedUser, err := models.GetUserByID(selectedUserID)
	if err != nil || selectedUser == nil {
		log.Printf("Error fetching selected user details: %v", err)
		// Fallback to current user if not found
		selectedUser, _ = models.GetUserByID(currentUserID)
	}

	// Get dates
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	isFiltered := r.URL.Query().Get("filter") == "1"

	// If no filter request has been made yet, default to the current month
	if !isFiltered && startDate == "" && endDate == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	}

	// Fetch tasks matching filters
	tasks, err := models.GetKPITasks(selectedUserID, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching KPI tasks: %v", err)
		tasks = []models.Task{}
	}

	// Calculate metrics
	var pending, progress, completed int
	for _, t := range tasks {
		switch t.Status {
		case "Pending":
			pending++
		case "In Progress":
			progress++
		case "Completed":
			completed++
		}
	}

	total := len(tasks)
	var completionRate float64
	if total > 0 {
		completionRate = (float64(completed) / float64(total)) * 100
	}

	// Fetch training count
	trainingCount, err := models.GetTrainingCountByUserAndPeriod(selectedUser.Username, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching training count for KPI: %v", err)
		trainingCount = 0
	}

	// Fetch leads count
	leadsCount, err := models.GetLeadCountByUserAndPeriod(selectedUserID, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching leads count for KPI: %v", err)
		leadsCount = 0
	}

	// Prepare status stats for Chart.js
	statusStats := map[string]int{
		"Pending":     pending,
		"In Progress": progress,
		"Completed":   completed,
	}

	statusJSON, err := json.Marshal(statusStats)
	if err != nil {
		statusJSON = []byte(`{"Pending":0,"In Progress":0,"Completed":0}`)
	}

	data := KPIDashboardData{
		Users:          users,
		Tasks:          tasks,
		SelectedUserID: selectedUserID,
		SelectedUser:   selectedUser,
		StartDate:      startDate,
		EndDate:        endDate,
		TotalTasks:     total,
		PendingTasks:   pending,
		ProgressTasks:  progress,
		CompletedTasks: completed,
		CompletionRate: completionRate,
		TrainingCount:  trainingCount,
		LeadsCount:     leadsCount,
		StatusJSON:     template.JS(statusJSON),
	}

	RenderTemplate(w, r, "kpi.html", "Dashboard User", "kpi", data, "", "")
}
