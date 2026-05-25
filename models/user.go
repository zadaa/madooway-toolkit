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
