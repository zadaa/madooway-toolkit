package models

import (
	"database/sql"
	"errors"
	"time"

	"task-manager-go/db"
)

type Client struct {
	ID           int
	Name         string
	ShortName    string
	Email        string
	Phone        string
	PicName      string
	PricePackage string
	Logo         string
	Province     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateClient inserts a client into the database
func CreateClient(name, shortName, email, phone, picName, pricePackage, logo, province string) error {
	query := `INSERT INTO clients (name, short_name, email, phone, pic_name, price_package, logo, province) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query, name, shortName, email, phone, picName, pricePackage, logo, province)
	return err
}

// UpdateClient updates details of an existing client
func UpdateClient(clientID int, name, shortName, email, phone, picName, pricePackage, logo, province string) error {
	query := `UPDATE clients SET name = ?, short_name = ?, email = ?, phone = ?, pic_name = ?, price_package = ?, logo = ?, province = ? WHERE id = ?`
	_, err := db.DB.Exec(query, name, shortName, email, phone, picName, pricePackage, logo, province, clientID)
	return err
}

// DeleteClient deletes a client from the database
func DeleteClient(clientID int) error {
	query := `DELETE FROM clients WHERE id = ?`
	_, err := db.DB.Exec(query, clientID)
	return err
}

// GetClientByID fetches a single client by id
func GetClientByID(clientID int) (*Client, error) {
	query := `SELECT id, name, COALESCE(short_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(pic_name, ''), price_package, COALESCE(logo, ''), COALESCE(province, 'DKI Jakarta'), created_at, updated_at FROM clients WHERE id = ?`
	row := db.DB.QueryRow(query, clientID)

	var c Client
	err := row.Scan(&c.ID, &c.Name, &c.ShortName, &c.Email, &c.Phone, &c.PicName, &c.PricePackage, &c.Logo, &c.Province, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// GetAllClients lists all clients in the system
func GetAllClients() ([]Client, error) {
	query := `SELECT id, name, COALESCE(short_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(pic_name, ''), price_package, COALESCE(logo, ''), COALESCE(province, 'DKI Jakarta'), created_at, updated_at FROM clients ORDER BY name ASC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var c Client
		err := rows.Scan(&c.ID, &c.Name, &c.ShortName, &c.Email, &c.Phone, &c.PicName, &c.PricePackage, &c.Logo, &c.Province, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, nil
}

type ProvinceStat struct {
	Province string `json:"province"`
	Count    int    `json:"count"`
}

// GetClientStatsByProvince counts clients grouped by province
func GetClientStatsByProvince() ([]ProvinceStat, error) {
	query := `SELECT COALESCE(province, 'DKI Jakarta') as prov, COUNT(*) 
	          FROM clients 
	          GROUP BY prov`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ProvinceStat
	for rows.Next() {
		var ps ProvinceStat
		if err := rows.Scan(&ps.Province, &ps.Count); err != nil {
			return nil, err
		}
		stats = append(stats, ps)
	}
	return stats, nil
}
