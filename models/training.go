package models

import (
	"errors"
	"time"

	"task-manager-go/db"
)

type TrainingSchedule struct {
	ID           int
	UserID       int
	ClientID     int
	ClientName   string
	Title        string
	Description  string
	TrainingDate time.Time
	Location     string
	Trainer      string
	TrainingType string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FormattedTrainingDate formats training date to a pretty string
func (t *TrainingSchedule) FormattedTrainingDate() string {
	return t.TrainingDate.Format("02 Jan 2006, 15:04")
}

// FormattedDateTimeLocal formats date matching <input type="datetime-local">
func (t *TrainingSchedule) FormattedDateTimeLocal() string {
	return t.TrainingDate.Format("2006-01-02T15:04")
}

// CreateTrainingSchedule inserts a new schedule into the database
func CreateTrainingSchedule(userID, clientID int, title, description, dateStr, location, trainer, trainingType, status string) error {
	var parsedDate time.Time
	var err error
	
	// Try parsing with seconds first, then without seconds
	parsedDate, err = time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		parsedDate, err = time.Parse("2006-01-02T15:04", dateStr)
		if err != nil {
			return errors.New("format tanggal/jam tidak valid, gunakan YYYY-MM-DDTHH:MM")
		}
	}

	query := `INSERT INTO training_schedules (user_id, client_id, title, description, training_date, location, trainer, training_type, status)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = db.DB.Exec(query, userID, clientID, title, description, parsedDate, location, trainer, trainingType, status)
	return err
}

// UpdateTrainingSchedule updates details of an existing schedule
func UpdateTrainingSchedule(id, userID, clientID int, title, description, dateStr, location, trainer, trainingType, status string) error {
	var parsedDate time.Time
	var err error
	
	// Try parsing with seconds first, then without seconds
	parsedDate, err = time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		parsedDate, err = time.Parse("2006-01-02T15:04", dateStr)
		if err != nil {
			return errors.New("format tanggal/jam tidak valid, gunakan YYYY-MM-DDTHH:MM")
		}
	}

	query := `UPDATE training_schedules 
	          SET client_id = ?, title = ?, description = ?, training_date = ?, location = ?, trainer = ?, training_type = ?, status = ?
	          WHERE id = ?`
	_, err = db.DB.Exec(query, clientID, title, description, parsedDate, location, trainer, trainingType, status, id)
	return err
}

// DeleteTrainingSchedule deletes a training schedule
func DeleteTrainingSchedule(id, userID int) error {
	query := `DELETE FROM training_schedules WHERE id = ?`
	_, err := db.DB.Exec(query, id)
	return err
}

// GetTrainingSchedulesByUserID lists all training schedules for a user
func GetTrainingSchedulesByUserID(userID int) ([]TrainingSchedule, error) {
	query := `SELECT ts.id, ts.user_id, ts.client_id, c.name as client_name, ts.title, COALESCE(ts.description, ''), ts.training_date, COALESCE(ts.location, ''), COALESCE(ts.trainer, ''), ts.training_type, ts.status, ts.created_at, ts.updated_at 
	          FROM training_schedules ts
	          INNER JOIN clients c ON ts.client_id = c.id
	          ORDER BY ts.training_date ASC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []TrainingSchedule
	for rows.Next() {
		var ts TrainingSchedule
		err := rows.Scan(&ts.ID, &ts.UserID, &ts.ClientID, &ts.ClientName, &ts.Title, &ts.Description, &ts.TrainingDate, &ts.Location, &ts.Trainer, &ts.TrainingType, &ts.Status, &ts.CreatedAt, &ts.UpdatedAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, ts)
	}
	return schedules, nil
}

// GetTrainingScheduleByID fetches a single training schedule by ID and user ID
func GetTrainingScheduleByID(id, userID int) (*TrainingSchedule, error) {
	query := `SELECT ts.id, ts.user_id, ts.client_id, c.name as client_name, ts.title, COALESCE(ts.description, ''), ts.training_date, COALESCE(ts.location, ''), COALESCE(ts.trainer, ''), ts.training_type, ts.status, ts.created_at, ts.updated_at 
	          FROM training_schedules ts
	          INNER JOIN clients c ON ts.client_id = c.id
	          WHERE ts.id = ?`
	row := db.DB.QueryRow(query, id)

	var ts TrainingSchedule
	err := row.Scan(&ts.ID, &ts.UserID, &ts.ClientID, &ts.ClientName, &ts.Title, &ts.Description, &ts.TrainingDate, &ts.Location, &ts.Trainer, &ts.TrainingType, &ts.Status, &ts.CreatedAt, &ts.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ts, nil
}
