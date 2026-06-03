package models

import (
	"database/sql"
	"errors"
	"time"

	"task-manager-go/db"
)

type Lead struct {
	ID              int
	Source          string // Referal, Linkedin, Instagram, BP
	CompanyName     string
	ContactName     string
	Phone           string
	Email           string
	EmployeeCount   int
	Status          string // Reachout, Demo, Convert
	SalesID         int
	SalesName       string // To display username
	FollowUpHistory string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PrettyCreatedAt formats created_at to a readable string
func (l *Lead) PrettyCreatedAt() string {
	return l.CreatedAt.Format("02 Jan 2006, 15:04")
}

// CreateLead inserts a new lead
func CreateLead(source, companyName, contactName, phone, email string, employeeCount int, status string, salesID int) error {
	query := `INSERT INTO leads (source, company_name, contact_name, phone, email, employee_count, status, sales_id, follow_up_history)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, '')`
	_, err := db.DB.Exec(query, source, companyName, contactName, phone, email, employeeCount, status, salesID)
	return err
}

// UpdateLead updates details of a lead
func UpdateLead(leadID int, source, companyName, contactName, phone, email string, employeeCount int, status string, salesID int) error {
	query := `UPDATE leads SET source = ?, company_name = ?, contact_name = ?, phone = ?, email = ?, employee_count = ?, status = ?, sales_id = ?
	          WHERE id = ?`
	_, err := db.DB.Exec(query, source, companyName, contactName, phone, email, employeeCount, status, salesID, leadID)
	return err
}

// UpdateFollowUpHistory updates the follow up history of a lead
func UpdateFollowUpHistory(leadID int, history string) error {
	query := `UPDATE leads SET follow_up_history = ? WHERE id = ?`
	_, err := db.DB.Exec(query, history, leadID)
	return err
}

// DeleteLead deletes a lead from database
func DeleteLead(leadID int) error {
	query := `DELETE FROM leads WHERE id = ?`
	_, err := db.DB.Exec(query, leadID)
	return err
}

// GetLeadByID fetches a single lead by id
func GetLeadByID(leadID int) (*Lead, error) {
	query := `SELECT l.id, l.source, l.company_name, l.contact_name, COALESCE(l.phone, ''), COALESCE(l.email, ''), COALESCE(l.employee_count, 0), l.status, l.sales_id, u.username as sales_name, COALESCE(l.follow_up_history, ''), l.created_at, l.updated_at
	          FROM leads l
	          INNER JOIN users u ON l.sales_id = u.id
	          WHERE l.id = ?`
	row := db.DB.QueryRow(query, leadID)

	var l Lead
	err := row.Scan(&l.ID, &l.Source, &l.CompanyName, &l.ContactName, &l.Phone, &l.Email, &l.EmployeeCount, &l.Status, &l.SalesID, &l.SalesName, &l.FollowUpHistory, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

// GetAllLeads fetches all leads
func GetAllLeads() ([]Lead, error) {
	query := `SELECT l.id, l.source, l.company_name, l.contact_name, COALESCE(l.phone, ''), COALESCE(l.email, ''), COALESCE(l.employee_count, 0), l.status, l.sales_id, u.username as sales_name, COALESCE(l.follow_up_history, ''), l.created_at, l.updated_at
	          FROM leads l
	          INNER JOIN users u ON l.sales_id = u.id
	          ORDER BY l.created_at DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leads []Lead
	for rows.Next() {
		var l Lead
		err := rows.Scan(&l.ID, &l.Source, &l.CompanyName, &l.ContactName, &l.Phone, &l.Email, &l.EmployeeCount, &l.Status, &l.SalesID, &l.SalesName, &l.FollowUpHistory, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}
		leads = append(leads, l)
	}
	return leads, nil
}

// GetLeadCountByUserAndPeriod counts leads assigned to a sales officer (user) in a period
func GetLeadCountByUserAndPeriod(userID int, startDate, endDate string) (int, error) {
	query := `SELECT COUNT(*) FROM leads WHERE sales_id = ?`
	var args []interface{}
	args = append(args, userID)

	if startDate != "" {
		query += " AND DATE(created_at) >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND DATE(created_at) <= ?"
		args = append(args, endDate)
	}

	var count int
	err := db.DB.QueryRow(query, args...).Scan(&count)
	return count, err
}
