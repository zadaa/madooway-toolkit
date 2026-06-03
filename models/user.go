package models

import (
	"database/sql"
	"errors"
	"time"

	"task-manager-go/db"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Username     string
	Email        string
	NIP          string
	PasswordHash string
	Color        string
	Role         string
	CreatedAt    time.Time
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateUser(username, email, password, color, role, nip string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	if color == "" {
		color = "#4f46e5"
	}
	if role == "" {
		role = "user"
	}

	var nipVal interface{}
	if nip != "" {
		nipVal = nip
	}

	query := "INSERT INTO users (username, email, password_hash, color, role, nip) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = db.DB.Exec(query, username, email, hash, color, role, nipVal)
	return err
}

func GetUserByEmail(email string) (*User, error) {
	query := "SELECT id, username, email, COALESCE(nip, '') as nip, password_hash, color, role, created_at FROM users WHERE email = ?"
	row := db.DB.QueryRow(query, email)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.NIP, &user.PasswordHash, &user.Color, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByIdentifier retrieves a user by either email or NIP
func GetUserByIdentifier(identifier string) (*User, error) {
	query := "SELECT id, username, email, COALESCE(nip, '') as nip, password_hash, color, role, created_at FROM users WHERE email = ? OR nip = ?"
	row := db.DB.QueryRow(query, identifier, identifier)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.NIP, &user.PasswordHash, &user.Color, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// FormattedCreatedAt returns CreatedAt formatted as DD Jan 2006, 15:04
func (u *User) FormattedCreatedAt() string {
	return u.CreatedAt.Format("02 Jan 2006, 15:04")
}

// GetAllUsers retrieves all registered users ordered by username
func GetAllUsers() ([]User, error) {
	query := "SELECT id, username, email, COALESCE(nip, '') as nip, color, role, created_at FROM users ORDER BY username ASC"
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.NIP, &u.Color, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserByID retrieves a single user by ID
func GetUserByID(id int) (*User, error) {
	query := "SELECT id, username, email, COALESCE(nip, '') as nip, password_hash, color, role, created_at FROM users WHERE id = ?"
	row := db.DB.QueryRow(query, id)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.NIP, &u.PasswordHash, &u.Color, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// UpdateUser updates a user's credentials. Hashes password if provided.
func UpdateUser(id int, username, email, password, color, role, nip string) error {
	var err error
	if color == "" {
		color = "#4f46e5"
	}
	if role == "" {
		role = "user"
	}
	var nipVal interface{}
	if nip != "" {
		nipVal = nip
	}
	if password != "" {
		hash, err := HashPassword(password)
		if err != nil {
			return err
		}
		query := "UPDATE users SET username = ?, email = ?, password_hash = ?, color = ?, role = ?, nip = ? WHERE id = ?"
		_, err = db.DB.Exec(query, username, email, hash, color, role, nipVal, id)
	} else {
		query := "UPDATE users SET username = ?, email = ?, color = ?, role = ?, nip = ? WHERE id = ?"
		_, err = db.DB.Exec(query, username, email, color, role, nipVal, id)
	}
	return err
}

// DeleteUser deletes a user by ID
func DeleteUser(id int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := db.DB.Exec(query, id)
	return err
}

// CheckEmailExists checks if email exists for another user (excluding current user)
func CheckEmailExists(email string, excludeID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE email = ? AND id != ?"
	err := db.DB.QueryRow(query, email, excludeID).Scan(&count)
	return count > 0, err
}

// CheckUsernameExists checks if username exists for another user (excluding current user)
func CheckUsernameExists(username string, excludeID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = ? AND id != ?"
	err := db.DB.QueryRow(query, username, excludeID).Scan(&count)
	return count > 0, err
}

// CheckNIPExists checks if NIP exists for another user (excluding current user)
func CheckNIPExists(nip string, excludeID int) (bool, error) {
	if nip == "" {
		return false, nil
	}
	var count int
	query := "SELECT COUNT(*) FROM users WHERE nip = ? AND id != ?"
	err := db.DB.QueryRow(query, nip, excludeID).Scan(&count)
	return count > 0, err
}

type LeaderboardEntry struct {
	Rank          int
	UserID        int
	Username      string
	Email         string
	Color         string
	Role          string
	LeadsCount    int
	TicketsCount  int
	TasksCount    int
	TotalActivity int
}

// GetLeaderboard returns the list of users ranked by activity (leads, completed tickets, completed tasks)
func GetLeaderboard() ([]LeaderboardEntry, error) {
	query := `
		SELECT 
			u.id, 
			u.username, 
			u.email, 
			u.color, 
			u.role,
			COALESCE(leads_tbl.cnt, 0) as leads_count,
			COALESCE(tickets_tbl.cnt, 0) as tickets_count,
			COALESCE(tasks_tbl.cnt, 0) as tasks_count,
			(COALESCE(leads_tbl.cnt, 0) + COALESCE(tickets_tbl.cnt, 0) + COALESCE(tasks_tbl.cnt, 0)) as total_activity
		FROM users u
		LEFT JOIN (
			SELECT sales_id, COUNT(*) as cnt 
			FROM leads 
			GROUP BY sales_id
		) leads_tbl ON u.id = leads_tbl.sales_id
		LEFT JOIN (
			SELECT ta.user_id, COUNT(*) as cnt
			FROM ticket_assignees ta
			JOIN tickets t ON ta.ticket_id = t.id
			WHERE t.status = 'Completed'
			GROUP BY ta.user_id
		) tickets_tbl ON u.id = tickets_tbl.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) as cnt 
			FROM tasks 
			WHERE status = 'Completed'
			GROUP BY user_id
		) tasks_tbl ON u.id = tasks_tbl.user_id
		ORDER BY total_activity DESC, u.username ASC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var e LeaderboardEntry
		err := rows.Scan(&e.UserID, &e.Username, &e.Email, &e.Color, &e.Role, &e.LeadsCount, &e.TicketsCount, &e.TasksCount, &e.TotalActivity)
		if err != nil {
			return nil, err
		}
		e.Rank = rank
		entries = append(entries, e)
		rank++
	}
	return entries, nil
}

