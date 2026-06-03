package models

import (
	"database/sql"
	"errors"
	"fmt"
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

// SyncClientsFromRemote connects to a remote MySQL database and syncs companies to our local clients table
func SyncClientsFromRemote(remoteHost, remoteUser, remotePassword string) (int, error) {
	// Connect to remote server without specifying database
	dsnWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:3306)/?parseTime=true", remoteUser, remotePassword, remoteHost)
	remoteDB, err := sql.Open("mysql", dsnWithoutDB)
	if err != nil {
		return 0, err
	}
	defer remoteDB.Close()

	// Show databases to find the one containing "companies" table
	rows, err := remoteDB.Query("SHOW DATABASES")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var targetDB string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err == nil {
			if dbName == "information_schema" || dbName == "mysql" || dbName == "performance_schema" || dbName == "sys" {
				continue
			}
			var tableExists int
			checkQuery := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = 'companies'", dbName)
			err2 := remoteDB.QueryRow(checkQuery).Scan(&tableExists)
			if err2 == nil && tableExists > 0 {
				targetDB = dbName
				break
			}
		}
	}

	if targetDB == "" {
		return 0, errors.New("remote table 'companies' not found in any database")
	}

	// Connect to the specific remote database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", remoteUser, remotePassword, remoteHost, targetDB)
	dbRemote, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, err
	}
	defer dbRemote.Close()

	// Query remote columns dynamically to handle schema differences
	companiesRows, err := dbRemote.Query("SELECT * FROM companies")
	if err != nil {
		return 0, err
	}
	defer companiesRows.Close()

	cols, err := companiesRows.Columns()
	if err != nil {
		return 0, err
	}

	getRemoteValue := func(row map[string]interface{}, possibleKeys []string, defaultVal string) string {
		for _, k := range possibleKeys {
			if val, ok := row[k]; ok && val != nil {
				if strVal, ok := val.(string); ok {
					return strVal
				}
				if bytesVal, ok := val.([]byte); ok {
					return string(bytesVal)
				}
			}
		}
		return defaultVal
	}

	syncCount := 0
	for companiesRows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := companiesRows.Scan(columnPointers...); err != nil {
			return syncCount, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columns[i]
			m[colName] = val
		}

		name := getRemoteValue(m, []string{"name", "company_name", "title"}, "")
		if name == "" {
			continue // Skip unnamed companies
		}

		shortName := getRemoteValue(m, []string{"short_name", "shortname", "code", "alias"}, "")
		email := getRemoteValue(m, []string{"email", "contact_email"}, "")
		phone := getRemoteValue(m, []string{"phone", "phone_number", "telp", "whatsapp"}, "")
		picName := getRemoteValue(m, []string{"pic_name", "pic", "picname", "contact_name"}, "")
		pricePackage := getRemoteValue(m, []string{"price_package", "package", "subscription", "plan"}, "Basic Plan")
		logo := getRemoteValue(m, []string{"logo", "avatar", "image"}, "")
		province := getRemoteValue(m, []string{"province", "provinsi", "region", "city"}, "DKI Jakarta")

		var existingID int
		err = db.DB.QueryRow("SELECT id FROM clients WHERE name = ?", name).Scan(&existingID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return syncCount, err
		}

		if existingID > 0 {
			query := `UPDATE clients SET short_name = ?, email = ?, phone = ?, pic_name = ?, price_package = ?, logo = ?, province = ? WHERE id = ?`
			_, err = db.DB.Exec(query, shortName, email, phone, picName, pricePackage, logo, province, existingID)
		} else {
			query := `INSERT INTO clients (name, short_name, email, phone, pic_name, price_package, logo, province) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
			_, err = db.DB.Exec(query, name, shortName, email, phone, picName, pricePackage, logo, province)
		}

		if err != nil {
			return syncCount, err
		}
		syncCount++
	}

	return syncCount, nil
}
