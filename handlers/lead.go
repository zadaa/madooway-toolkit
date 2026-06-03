package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"task-manager-go/models"
)

type LeadsPageData struct {
	Leads []models.Lead
	Users []models.User
}

// ListLeads lists all leads
func ListLeads(w http.ResponseWriter, r *http.Request) {
	leads, err := models.GetAllLeads()
	if err != nil {
		log.Printf("Error fetching leads: %v", err)
		leads = []models.Lead{}
	}

	users, err := models.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users for leads dropdown: %v", err)
		users = []models.User{}
	}

	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	data := LeadsPageData{
		Leads: leads,
		Users: users,
	}

	RenderTemplate(w, r, "leads.html", "Leads Tracker", "leads", data, errorMsg, successMsg)
}

// CreateLead creates a new lead
func CreateLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/leads", http.StatusSeeOther)
		return
	}

	source := strings.TrimSpace(r.FormValue("source"))
	companyName := strings.TrimSpace(r.FormValue("company_name"))
	contactName := strings.TrimSpace(r.FormValue("contact_name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.TrimSpace(r.FormValue("email"))
	employeeCountStr := strings.TrimSpace(r.FormValue("employee_count"))
	status := strings.TrimSpace(r.FormValue("status"))
	salesIDStr := strings.TrimSpace(r.FormValue("sales_id"))

	if source == "" || companyName == "" || contactName == "" || salesIDStr == "" {
		http.Redirect(w, r, "/leads?error=Sumber,+Nama+Perusahaan,+Nama+Kontak,+dan+Sales+wajib+diisi", http.StatusSeeOther)
		return
	}

	employeeCount := 0
	if employeeCountStr != "" {
		if val, err := strconv.Atoi(employeeCountStr); err == nil {
			employeeCount = val
		}
	}

	salesID, err := strconv.Atoi(salesIDStr)
	if err != nil {
		http.Redirect(w, r, "/leads?error=Sales+tidak+valid", http.StatusSeeOther)
		return
	}

	if status == "" {
		status = "Reachout"
	}

	err = models.CreateLead(source, companyName, contactName, phone, email, employeeCount, status, salesID)
	if err != nil {
		log.Printf("Error creating lead: %v", err)
		http.Redirect(w, r, "/leads?error=Gagal+menambahkan+lead", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/leads?success=Lead+berhasil+ditambahkan", http.StatusSeeOther)
}

// UpdateLead updates a lead
func UpdateLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/leads", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/leads?error=ID+lead+tidak+valid", http.StatusSeeOther)
		return
	}

	source := strings.TrimSpace(r.FormValue("source"))
	companyName := strings.TrimSpace(r.FormValue("company_name"))
	contactName := strings.TrimSpace(r.FormValue("contact_name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.TrimSpace(r.FormValue("email"))
	employeeCountStr := strings.TrimSpace(r.FormValue("employee_count"))
	status := strings.TrimSpace(r.FormValue("status"))
	salesIDStr := strings.TrimSpace(r.FormValue("sales_id"))

	if source == "" || companyName == "" || contactName == "" || salesIDStr == "" {
		http.Redirect(w, r, "/leads?error=Sumber,+Nama+Perusahaan,+Nama+Kontak,+dan+Sales+wajib+diisi", http.StatusSeeOther)
		return
	}

	employeeCount := 0
	if employeeCountStr != "" {
		if val, err := strconv.Atoi(employeeCountStr); err == nil {
			employeeCount = val
		}
	}

	salesID, err := strconv.Atoi(salesIDStr)
	if err != nil {
		http.Redirect(w, r, "/leads?error=Sales+tidak+valid", http.StatusSeeOther)
		return
	}

	err = models.UpdateLead(id, source, companyName, contactName, phone, email, employeeCount, status, salesID)
	if err != nil {
		log.Printf("Error updating lead: %v", err)
		http.Redirect(w, r, "/leads?error=Gagal+memperbarui+lead", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/leads?success=Lead+berhasil+diperbarui", http.StatusSeeOther)
}

// DeleteLead deletes a lead
func DeleteLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/leads", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/leads?error=ID+lead+tidak+valid", http.StatusSeeOther)
		return
	}

	err = models.DeleteLead(id)
	if err != nil {
		log.Printf("Error deleting lead: %v", err)
		http.Redirect(w, r, "/leads?error=Gagal+menghapus+lead", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/leads?success=Lead+berhasil+dihapus", http.StatusSeeOther)
}

// UpdateLeadHistory updates the follow-up history of a lead
func UpdateLeadHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/leads", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/leads?error=ID+lead+tidak+valid", http.StatusSeeOther)
		return
	}

	history := strings.TrimSpace(r.FormValue("follow_up_history"))

	err = models.UpdateFollowUpHistory(id, history)
	if err != nil {
		log.Printf("Error updating lead follow up history: %v", err)
		http.Redirect(w, r, "/leads?error=Gagal+memperbarui+history+follow+up", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/leads?success=History+follow-up+berhasil+diperbarui", http.StatusSeeOther)
}
