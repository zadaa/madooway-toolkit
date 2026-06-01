package models

import (
	"time"

	"task-manager-go/db"
)

type TicketMessage struct {
	ID        int       `json:"id"`
	TicketID  int       `json:"ticket_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	FilePath  string    `json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
}

// PrettyCreatedAt returns CreatedAt formatted as "02 Jan 2006, 15:04"
func (tm *TicketMessage) PrettyCreatedAt() string {
	return tm.CreatedAt.Format("02 Jan 2006, 15:04")
}

// CreateTicketMessage inserts a message into ticket threads
func CreateTicketMessage(ticketID int, userID int, message string, filePath string) error {
	query := `INSERT INTO ticket_messages (ticket_id, user_id, message, file_path) 
	          VALUES (?, ?, ?, ?)`
	_, err := db.DB.Exec(query, ticketID, userID, message, filePath)
	return err
}

// GetTicketMessagesByTicketID retrieves all messages for a ticket
func GetTicketMessagesByTicketID(ticketID int) ([]TicketMessage, error) {
	query := `SELECT tm.id, tm.ticket_id, tm.user_id, COALESCE(u.username, '') as username, COALESCE(tm.message, ''), COALESCE(tm.file_path, ''), tm.created_at 
	          FROM ticket_messages tm 
	          JOIN users u ON tm.user_id = u.id 
	          WHERE tm.ticket_id = ? 
	          ORDER BY tm.created_at ASC`
	rows, err := db.DB.Query(query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []TicketMessage
	for rows.Next() {
		var tm TicketMessage
		err := rows.Scan(&tm.ID, &tm.TicketID, &tm.UserID, &tm.Username, &tm.Message, &tm.FilePath, &tm.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, tm)
	}
	return messages, nil
}
