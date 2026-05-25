package models

import (
	"database/sql"
	"strings"
	"time"

	"task-manager-go/db"
)

type Task struct {
	ID          int
	UserID      int
	Title       string
	Description string
	Category    string
	Source      string
	Status      string
	DueDate     sql.NullTime
	ClientID    sql.NullInt64
	ClientName  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DateStat struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// FormattedDueDate returns DueDate in YYYY-MM-DD format
func (t *Task) FormattedDueDate() string {
	if t.DueDate.Valid {
		return t.DueDate.Time.Format("2006-01-02")
	}
	return ""
}

// FormattedCreatedAt returns CreatedAt in YYYY-MM-DD format
func (t *Task) FormattedCreatedAt() string {
	return t.CreatedAt.Format("2006-01-02")
}

// PrettyCreatedAt returns CreatedAt in DD Jan 2006, 15:04 format
func (t *Task) PrettyCreatedAt() string {
	return t.CreatedAt.Format("02 Jan 2006, 15:04")
}

// IsOverdue checks if the task is past its due date and not completed
func (t *Task) IsOverdue() bool {
	if !t.DueDate.Valid || t.Status == "Completed" {
		return false
	}
	// Compare dates at midnight to ignore time of day
	today := time.Now().Truncate(24 * time.Hour)
	due := t.DueDate.Time.Truncate(24 * time.Hour)
	return due.Before(today)
}

// CreateTask adds a new task to the database
func CreateTask(userID int, title, description, category, source, status, dueDateStr string, clientID sql.NullInt64) error {
	var dueDate sql.NullTime
	if dueDateStr != "" {
		parsedTime, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			dueDate = sql.NullTime{Time: parsedTime, Valid: true}
		}
	}

	query := `INSERT INTO tasks (user_id, title, description, category, source, status, due_date, client_id)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query, userID, title, description, category, source, status, dueDate, clientID)
	return err
}

// UpdateTask updates an existing task for a user
func UpdateTask(taskID int, userID int, title, description, category, source, status, dueDateStr string, clientID sql.NullInt64) error {
	var dueDate sql.NullTime
	if dueDateStr != "" {
		parsedTime, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			dueDate = sql.NullTime{Time: parsedTime, Valid: true}
		}
	}

	query := `UPDATE tasks 
	          SET title = ?, description = ?, category = ?, source = ?, status = ?, due_date = ?, client_id = ?
	          WHERE id = ? AND user_id = ?`
	_, err := db.DB.Exec(query, title, description, category, source, status, dueDate, clientID, taskID, userID)
	return err
}

// DeleteTask deletes a task belonging to the user
func DeleteTask(taskID int, userID int) error {
	query := "DELETE FROM tasks WHERE id = ? AND user_id = ?"
	_, err := db.DB.Exec(query, taskID, userID)
	return err
}

// GetTaskByID retrieves a single task
func GetTaskByID(taskID int, userID int) (*Task, error) {
	query := `SELECT id, user_id, title, COALESCE(description, ''), category, source, status, due_date, client_id, created_at, updated_at 
	          FROM tasks WHERE id = ? AND user_id = ?`
	row := db.DB.QueryRow(query, taskID, userID)

	var t Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Category, &t.Source, &t.Status, &t.DueDate, &t.ClientID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTasksByUserID retrieves list of tasks with filters and search query
func GetTasksByUserID(userID int, search, category, status string) ([]Task, error) {
	var queryParts []string
	var args []interface{}

	queryParts = append(queryParts, `SELECT t.id, t.user_id, t.title, COALESCE(t.description, ''), t.category, t.source, t.status, t.due_date, t.client_id, COALESCE(c.name, '') as client_name, t.created_at, t.updated_at 
	                                 FROM tasks t 
	                                 LEFT JOIN clients c ON t.client_id = c.id 
	                                 WHERE t.user_id = ?`)
	args = append(args, userID)

	if search != "" {
		queryParts = append(queryParts, "(t.title LIKE ? OR t.description LIKE ?)")
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if category != "" {
		queryParts = append(queryParts, "t.category = ?")
		args = append(args, category)
	}
	if status != "" {
		queryParts = append(queryParts, "t.status = ?")
		args = append(args, status)
	}

	// Join all with AND
	fullQuery := queryParts[0]
	if len(queryParts) > 1 {
		fullQuery += " AND " + strings.Join(queryParts[1:], " AND ")
	}
	// Sort by due date
	fullQuery += " ORDER BY CASE WHEN t.due_date IS NULL THEN 1 ELSE 0 END, t.due_date ASC, t.created_at DESC"

	rows, err := db.DB.Query(fullQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Category, &t.Source, &t.Status, &t.DueDate, &t.ClientID, &t.ClientName, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// GetTaskStatsByCategory counts tasks per category for a user
func GetTaskStatsByCategory(userID int) (map[string]int, error) {
	query := "SELECT category, COUNT(*) FROM tasks WHERE user_id = ? GROUP BY category"
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		stats[category] = count
	}
	return stats, nil
}

// GetTaskStatsByStatus counts tasks per status for a user
func GetTaskStatsByStatus(userID int) (map[string]int, error) {
	query := "SELECT status, COUNT(*) FROM tasks WHERE user_id = ? GROUP BY status"
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	// Initialize default keys to ensure frontend always has them
	stats["Pending"] = 0
	stats["In Progress"] = 0
	stats["Completed"] = 0

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, nil
}

// GetTaskStatsByDate counts tasks created per date over the last 14 days
func GetTaskStatsByDate(userID int) ([]DateStat, error) {
	// Query task counts grouped by date for the last 14 days
	query := `SELECT DATE(created_at) as task_date, COUNT(*) 
	          FROM tasks 
	          WHERE user_id = ? AND created_at >= DATE_SUB(CURDATE(), INTERVAL 13 DAY)
	          GROUP BY task_date
	          ORDER BY task_date ASC`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbStats := make(map[string]int)
	for rows.Next() {
		var dateStr string
		var count int
		if err := rows.Scan(&dateStr, &count); err != nil {
			return nil, err
		}
		// mysql returns date as YYYY-MM-DD
		// strip time if present
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		dbStats[dateStr] = count
	}

	// Generate all last 14 days list to fill zeros
	var stats []DateStat
	now := time.Now()
	for i := 13; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		count := 0
		if val, ok := dbStats[d]; ok {
			count = val
		}
		// Format to prettier date for charts e.g., "May 24"
		t, _ := time.Parse("2006-01-02", d)
		prettyDate := t.Format("Jan 02")
		stats = append(stats, DateStat{
			Date:  prettyDate,
			Count: count,
		})
	}

	return stats, nil
}

// GetKPIStats gathers total counts and completion rate for KPI cards
type KPIStats struct {
	Total          int
	Pending        int
	InProgress     int
	Completed      int
	CompletionRate float64
}

func GetKPIStats(userID int) (*KPIStats, error) {
	query := `SELECT 
				COUNT(*),
				SUM(CASE WHEN status = 'Pending' THEN 1 ELSE 0 END),
				SUM(CASE WHEN status = 'In Progress' THEN 1 ELSE 0 END),
				SUM(CASE WHEN status = 'Completed' THEN 1 ELSE 0 END)
	          FROM tasks 
	          WHERE user_id = ?`
	row := db.DB.QueryRow(query, userID)

	var total, pending, inProgress, completed sql.NullInt64
	err := row.Scan(&total, &pending, &inProgress, &completed)
	if err != nil {
		return nil, err
	}

	kpi := &KPIStats{
		Total:      int(total.Int64),
		Pending:    int(pending.Int64),
		InProgress: int(inProgress.Int64),
		Completed:  int(completed.Int64),
	}

	if kpi.Total > 0 {
		kpi.CompletionRate = (float64(kpi.Completed) / float64(kpi.Total)) * 100
	} else {
		kpi.CompletionRate = 0.0
	}

	return kpi, nil
}
