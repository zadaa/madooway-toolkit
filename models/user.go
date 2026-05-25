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
	PasswordHash string
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

func CreateUser(username, email, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	query := "INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)"
	_, err = db.DB.Exec(query, username, email, hash)
	return err
}

func GetUserByEmail(email string) (*User, error) {
	query := "SELECT id, username, email, password_hash, created_at FROM users WHERE email = ?"
	row := db.DB.QueryRow(query, email)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
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
	query := "SELECT id, username, email, created_at FROM users ORDER BY username ASC"
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserByID retrieves a single user by ID
func GetUserByID(id int) (*User, error) {
	query := "SELECT id, username, email, password_hash, created_at FROM users WHERE id = ?"
	row := db.DB.QueryRow(query, id)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// UpdateUser updates a user's credentials. Hashes password if provided.
func UpdateUser(id int, username, email, password string) error {
	var err error
	if password != "" {
		hash, err := HashPassword(password)
		if err != nil {
			return err
		}
		query := "UPDATE users SET username = ?, email = ?, password_hash = ? WHERE id = ?"
		_, err = db.DB.Exec(query, username, email, hash, id)
	} else {
		query := "UPDATE users SET username = ?, email = ? WHERE id = ?"
		_, err = db.DB.Exec(query, username, email, id)
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

