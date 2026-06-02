package models

import (
	"database/sql"
	"strings"
	"time"

	"task-manager-go/db"
)

type Ticket struct {
	ID              int
	ClientID        int
	ClientName      string
	Title           string
	Description     string
	UserID          int
	CreatorUsername string
	FilePath        string
	IssueDate       time.Time
	Category        string
	TicketLink      string
	Status          string
	FinishedDate    sql.NullTime
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Assignees       []User
}

// FormattedIssueDate returns IssueDate in YYYY-MM-DD format
func (t *Ticket) FormattedIssueDate() string {
	return t.IssueDate.Format("2006-01-02")
}

// PrettyIssueDate returns IssueDate in DD Jan 2006 format
func (t *Ticket) PrettyIssueDate() string {
	return t.IssueDate.Format("02 Jan 2006")
}

// FormattedFinishedDate returns FinishedDate in YYYY-MM-DD format
func (t *Ticket) FormattedFinishedDate() string {
	if t.FinishedDate.Valid {
		return t.FinishedDate.Time.Format("2006-01-02")
	}
	return ""
}

// PrettyFinishedDate returns FinishedDate in DD Jan 2006 format
func (t *Ticket) PrettyFinishedDate() string {
	if t.FinishedDate.Valid {
		return t.FinishedDate.Time.Format("02 Jan 2006")
	}
	return "-"
}

// CreateTicket inserts a new ticket into the database
func CreateTicket(clientID int, title, description string, userID int, filePath, issueDateStr, category, ticketLink, status, finishedDateStr string) error {
	var issueDate time.Time
	var err error
	if issueDateStr != "" {
		issueDate, err = time.Parse("2006-01-02", issueDateStr)
		if err != nil {
			issueDate = time.Now()
		}
	} else {
		issueDate = time.Now()
	}

	var finishedDate sql.NullTime
	if finishedDateStr != "" {
		parsedTime, err := time.Parse("2006-01-02", finishedDateStr)
		if err == nil {
			finishedDate = sql.NullTime{Time: parsedTime, Valid: true}
		}
	}

	query := `INSERT INTO tickets (client_id, title, description, user_id, file_path, issue_date, category, ticket_link, status, finished_date)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = db.DB.Exec(query, clientID, title, description, userID, filePath, issueDate, category, ticketLink, status, finishedDate)
	return err
}

// UpdateTicket updates an existing ticket
func UpdateTicket(ticketID int, clientID int, title, description string, filePath, issueDateStr, category, ticketLink, status, finishedDateStr string) error {
	var issueDate time.Time
	var err error
	if issueDateStr != "" {
		issueDate, err = time.Parse("2006-01-02", issueDateStr)
		if err != nil {
			issueDate = time.Now()
		}
	} else {
		issueDate = time.Now()
	}

	var finishedDate sql.NullTime
	if finishedDateStr != "" && status == "Completed" {
		parsedTime, err := time.Parse("2006-01-02", finishedDateStr)
		if err == nil {
			finishedDate = sql.NullTime{Time: parsedTime, Valid: true}
		}
	} else if status == "Completed" {
		// If status is completed but finished date not provided, set to today
		finishedDate = sql.NullTime{Time: time.Now(), Valid: true}
	}

	query := `UPDATE tickets 
	          SET client_id = ?, title = ?, description = ?, file_path = ?, issue_date = ?, category = ?, ticket_link = ?, status = ?, finished_date = ?
	          WHERE id = ?`
	_, err = db.DB.Exec(query, clientID, title, description, filePath, issueDate, category, ticketLink, status, finishedDate, ticketID)
	return err
}

// DeleteTicket deletes a ticket by ID
func DeleteTicket(ticketID int) error {
	query := "DELETE FROM tickets WHERE id = ?"
	_, err := db.DB.Exec(query, ticketID)
	return err
}

// GetTicketByID retrieves a single ticket
func GetTicketByID(ticketID int) (*Ticket, error) {
	query := `SELECT t.id, t.client_id, COALESCE(c.name, '') as client_name, t.title, COALESCE(t.description, ''), t.user_id, COALESCE(u.username, '') as creator_username, COALESCE(t.file_path, ''), t.issue_date, t.category, COALESCE(t.ticket_link, ''), t.status, t.finished_date, t.created_at, t.updated_at 
	          FROM tickets t 
	          JOIN clients c ON t.client_id = c.id
	          JOIN users u ON t.user_id = u.id
	          WHERE t.id = ?`
	row := db.DB.QueryRow(query, ticketID)

	var t Ticket
	err := row.Scan(&t.ID, &t.ClientID, &t.ClientName, &t.Title, &t.Description, &t.UserID, &t.CreatorUsername, &t.FilePath, &t.IssueDate, &t.Category, &t.TicketLink, &t.Status, &t.FinishedDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetAllTickets retrieves all tickets with optional filters
func GetAllTickets(search, clientName, category, status, startDate, endDate string) ([]Ticket, error) {
	var queryParts []string
	var args []interface{}

	baseQuery := `SELECT t.id, t.client_id, COALESCE(c.name, '') as client_name, t.title, COALESCE(t.description, ''), t.user_id, COALESCE(u.username, '') as creator_username, COALESCE(t.file_path, ''), t.issue_date, t.category, COALESCE(t.ticket_link, ''), t.status, t.finished_date, t.created_at, t.updated_at 
	              FROM tickets t 
	              JOIN clients c ON t.client_id = c.id
	              JOIN users u ON t.user_id = u.id`

	if search != "" {
		queryParts = append(queryParts, "(t.title LIKE ? OR t.description LIKE ?)")
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if clientName != "" {
		queryParts = append(queryParts, "c.name = ?")
		args = append(args, clientName)
	}
	if category != "" {
		queryParts = append(queryParts, "t.category = ?")
		args = append(args, category)
	}
	if status != "" {
		queryParts = append(queryParts, "t.status = ?")
		args = append(args, status)
	}
	if startDate != "" {
		queryParts = append(queryParts, "t.issue_date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		queryParts = append(queryParts, "t.issue_date <= ?")
		args = append(args, endDate)
	}

	fullQuery := baseQuery
	if len(queryParts) > 0 {
		fullQuery += " WHERE " + strings.Join(queryParts, " AND ")
	}

	// Sort by issue date descending
	fullQuery += " ORDER BY t.issue_date DESC, t.created_at DESC"

	rows, err := db.DB.Query(fullQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []Ticket
	for rows.Next() {
		var t Ticket
		err := rows.Scan(&t.ID, &t.ClientID, &t.ClientName, &t.Title, &t.Description, &t.UserID, &t.CreatorUsername, &t.FilePath, &t.IssueDate, &t.Category, &t.TicketLink, &t.Status, &t.FinishedDate, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	return tickets, nil
}

// GetTicketAssignees retrieves the list of users assigned to a specific ticket
func GetTicketAssignees(ticketID int) ([]User, error) {
	query := `SELECT u.id, u.username, u.email, u.color, u.created_at 
	          FROM users u 
	          INNER JOIN ticket_assignees ta ON u.id = ta.user_id 
	          WHERE ta.ticket_id = ? 
	          ORDER BY u.username ASC`
	rows, err := db.DB.Query(query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignees []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Color, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		assignees = append(assignees, u)
	}
	return assignees, nil
}

// AssignTicket assigns multiple users to a ticket, replacing previous assignments
func AssignTicket(ticketID int, userIDs []int) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing assignees
	_, err = tx.Exec("DELETE FROM ticket_assignees WHERE ticket_id = ?", ticketID)
	if err != nil {
		return err
	}

	// Insert new assignees
	if len(userIDs) > 0 {
		stmt, err := tx.Prepare("INSERT INTO ticket_assignees (ticket_id, user_id) VALUES (?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, userID := range userIDs {
			_, err = stmt.Exec(ticketID, userID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
