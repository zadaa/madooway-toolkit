package main

import (
	"fmt"
	"log"
	"net/http"

	"task-manager-go/config"
	"task-manager-go/db"
	"task-manager-go/handlers"
	"task-manager-go/middleware"
)

func main() {
	// 1. Load configuration
	config.LoadConfig()

	// 2. Initialize Database and Auto-migrations
	db.InitDB()
	defer db.DB.Close()

	// 3. Set up Router
	mux := http.NewServeMux()

	// Serve Static Files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Guest Routes (Login & Register)
	mux.HandleFunc("GET /login", middleware.GuestOnly(handlers.ShowLogin))
	mux.HandleFunc("POST /login", middleware.GuestOnly(handlers.HandleLogin))
	mux.HandleFunc("GET /register", middleware.GuestOnly(handlers.ShowRegister))
	mux.HandleFunc("POST /register", middleware.GuestOnly(handlers.HandleRegister))

	// Authenticated Routes
	mux.HandleFunc("GET /logout", handlers.HandleLogout)
	mux.HandleFunc("GET /dashboard", middleware.AdminOnly(handlers.ShowDashboard))
	mux.HandleFunc("GET /kpi", middleware.AdminOnly(handlers.ShowUserKPI))
	mux.HandleFunc("GET /kpi/", middleware.AdminOnly(handlers.ShowUserKPI))
	
	mux.HandleFunc("GET /tasks", middleware.AuthRequired(handlers.ListTasks))
	mux.HandleFunc("POST /tasks/create", middleware.AuthRequired(handlers.CreateTask))
	mux.HandleFunc("POST /tasks/update", middleware.AuthRequired(handlers.UpdateTask))
	mux.HandleFunc("POST /tasks/delete", middleware.AuthRequired(handlers.DeleteTask))

	mux.HandleFunc("GET /tickets", middleware.AuthRequired(handlers.ListTickets))
	mux.HandleFunc("POST /tickets/create", middleware.AuthRequired(handlers.CreateTicket))
	mux.HandleFunc("POST /tickets/update", middleware.AuthRequired(handlers.UpdateTicket))
	mux.HandleFunc("POST /tickets/delete", middleware.AuthRequired(handlers.DeleteTicket))
	mux.HandleFunc("POST /tickets/assign", middleware.AuthRequired(handlers.AssignTicket))
	mux.HandleFunc("POST /tickets/clickup", middleware.AuthRequired(handlers.CreateClickUpTaskForTicket))

	mux.HandleFunc("GET /tickets/messages", middleware.AuthRequired(handlers.GetTicketMessagesJSON))
	mux.HandleFunc("POST /tickets/messages/create", middleware.AuthRequired(handlers.CreateTicketMessageAJAX))

	mux.HandleFunc("GET /clients", middleware.AdminOnly(handlers.ListClients))
	mux.HandleFunc("POST /clients/create", middleware.AdminOnly(handlers.CreateClient))
	mux.HandleFunc("POST /clients/update", middleware.AdminOnly(handlers.UpdateClient))
	mux.HandleFunc("POST /clients/delete", middleware.AdminOnly(handlers.DeleteClient))
	mux.HandleFunc("POST /clients/sync", middleware.AdminOnly(handlers.SyncClients))


	mux.HandleFunc("GET /trainings", middleware.AuthRequired(handlers.ListTrainings))
	mux.HandleFunc("POST /trainings/create", middleware.AuthRequired(handlers.CreateTraining))
	mux.HandleFunc("POST /trainings/update", middleware.AuthRequired(handlers.UpdateTraining))
	mux.HandleFunc("POST /trainings/delete", middleware.AuthRequired(handlers.DeleteTraining))

	mux.HandleFunc("GET /performance", middleware.AuthRequired(handlers.ListPerformances))
	mux.HandleFunc("POST /performance/create", middleware.AuthRequired(handlers.CreatePerformance))
	mux.HandleFunc("POST /performance/update", middleware.AuthRequired(handlers.UpdatePerformance))
	mux.HandleFunc("POST /performance/delete", middleware.AuthRequired(handlers.DeletePerformance))

	mux.HandleFunc("GET /users", middleware.AdminOnly(handlers.ListUsers))
	mux.HandleFunc("POST /users/create", middleware.AdminOnly(handlers.CreateUser))
	mux.HandleFunc("POST /users/update", middleware.AdminOnly(handlers.UpdateUser))
	mux.HandleFunc("POST /users/delete", middleware.AdminOnly(handlers.DeleteUser))

	// Leads Tracker Routes
	mux.HandleFunc("GET /leads", middleware.AuthRequired(handlers.ListLeads))
	mux.HandleFunc("POST /leads/create", middleware.AuthRequired(handlers.CreateLead))
	mux.HandleFunc("POST /leads/update", middleware.AuthRequired(handlers.UpdateLead))
	mux.HandleFunc("POST /leads/delete", middleware.AuthRequired(handlers.DeleteLead))
	mux.HandleFunc("POST /leads/history", middleware.AuthRequired(handlers.UpdateLeadHistory))

	// Root Redirect Handler
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		userID, loggedIn := middleware.GetSessionUser(r)
		if loggedIn {
			var role string
			err := db.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
			if err == nil && role == "admin" {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/tasks", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// 4. Start HTTP Server
	addr := fmt.Sprintf(":%s", config.AppConfig.Port)
	log.Printf("Server starting on http://localhost%s\n", addr)
	
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
