package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"task-manager-go/db"
)

type Payment struct {
	ID             int          `json:"id"`
	ClientID       int          `json:"client_id"`
	ClientName     string       `json:"client_name"`
	ClientPhone    string       `json:"client_phone"`
	OrderID        string       `json:"order_id"`
	Amount         float64      `json:"amount"`
	PackageName    string       `json:"package_name"`
	Status         string       `json:"status"`
	PaymentType    string       `json:"payment_type"`
	SnapToken      string       `json:"snap_token"`
	SnapURL        string       `json:"snap_url"`
	Tahap          string       `json:"tahap"`
	DPP            float64      `json:"dpp"`
	PPN            float64      `json:"ppn"`
	PPh            float64      `json:"pph"`
	TanggalInvoice sql.NullTime `json:"tanggal_invoice"`
	InvoiceMonth   string       `json:"invoice_month"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// CreatePayment inserts a new payment record into the database
func CreatePayment(clientID int, orderID string, amount float64, packageName, snapToken, snapURL string) error {
	query := `INSERT INTO payments (client_id, order_id, amount, package_name, status, snap_token, snap_url) VALUES (?, ?, ?, ?, 'Pending', ?, ?)`
	_, err := db.DB.Exec(query, clientID, orderID, amount, packageName, snapToken, snapURL)
	return err
}

// UpdatePaymentStatus updates the status and payment method of a payment by its order ID
func UpdatePaymentStatus(orderID string, status, paymentType string) error {
	query := `UPDATE payments SET status = ?, payment_type = ? WHERE order_id = ?`
	_, err := db.DB.Exec(query, status, paymentType, orderID)
	return err
}

// UpdatePaymentSnapToken updates the snap token and redirect URL of a payment
func UpdatePaymentSnapToken(orderID, snapToken, snapURL string) error {
	query := `UPDATE payments SET snap_token = ?, snap_url = ? WHERE order_id = ?`
	_, err := db.DB.Exec(query, snapToken, snapURL, orderID)
	return err
}

// GetAllPayments fetches all payments, joining with clients to get their details
func GetAllPayments() ([]Payment, error) {
	query := `
		SELECT p.id, p.client_id, c.name, COALESCE(c.phone, ''), p.order_id, p.amount, p.package_name, p.status, 
		       COALESCE(p.payment_type, ''), COALESCE(p.snap_token, ''), COALESCE(p.snap_url, ''), 
		       COALESCE(p.tahap, ''), COALESCE(p.dpp, 0.0), COALESCE(p.ppn, 0.0), COALESCE(p.pph, 0.0), p.tanggal_invoice,
		       p.created_at, p.updated_at 
		FROM payments p
		JOIN clients c ON p.client_id = c.id
		ORDER BY p.created_at DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		err := rows.Scan(
			&p.ID, &p.ClientID, &p.ClientName, &p.ClientPhone, &p.OrderID, &p.Amount, &p.PackageName, &p.Status,
			&p.PaymentType, &p.SnapToken, &p.SnapURL, &p.Tahap, &p.DPP, &p.PPN, &p.PPh, &p.TanggalInvoice,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		// Populate InvoiceMonth
		if p.TanggalInvoice.Valid {
			p.InvoiceMonth = p.TanggalInvoice.Time.Format("2006-01")
		} else {
			p.InvoiceMonth = p.CreatedAt.Format("2006-01")
		}
		payments = append(payments, p)
	}
	return payments, nil
}

// GetPaymentByOrderID retrieves a single payment record by its order ID
func GetPaymentByOrderID(orderID string) (*Payment, error) {
	query := `
		SELECT id, client_id, order_id, amount, package_name, status, 
		       COALESCE(payment_type, ''), COALESCE(snap_token, ''), COALESCE(snap_url, ''), 
		       COALESCE(tahap, ''), COALESCE(dpp, 0.0), COALESCE(ppn, 0.0), COALESCE(pph, 0.0), tanggal_invoice,
		       created_at, updated_at 
		FROM payments 
		WHERE order_id = ?`

	row := db.DB.QueryRow(query, orderID)

	var p Payment
	err := row.Scan(
		&p.ID, &p.ClientID, &p.OrderID, &p.Amount, &p.PackageName, &p.Status,
		&p.PaymentType, &p.SnapToken, &p.SnapURL, &p.Tahap, &p.DPP, &p.PPN, &p.PPh, &p.TanggalInvoice,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if p.TanggalInvoice.Valid {
		p.InvoiceMonth = p.TanggalInvoice.Time.Format("2006-01")
	} else {
		p.InvoiceMonth = p.CreatedAt.Format("2006-01")
	}
	return &p, nil
}

// SyncPaymentsFromRemote connects to the local db_madoo_ms_finance and syncs payments for the given month (YYYY-MM)
func SyncPaymentsFromRemote(month string) (int, error) {
	// Connect to local db_madoo_ms_finance database
	dsn := "root:07Mei2015@tcp(127.0.0.1:3306)/db_madoo_ms_finance?parseTime=true"
	sourceDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to db_madoo_ms_finance: %w", err)
	}
	defer sourceDB.Close()

	// Query payments filtered by printing date / transaction date month
	query := `
		SELECT COALESCE(project_name, ''), COALESCE(tahap, ''), COALESCE(invoice, ''), 
		       COALESCE(jumlah, 0.0), COALESCE(dpp, 0.0), COALESCE(ppn, 0.0), COALESCE(pph, 0.0), 
		       tanggal 
		FROM payments 
		WHERE DATE_FORMAT(tanggal, '%Y-%m') = ?`

	rows, err := sourceDB.Query(query, month)
	if err != nil {
		return 0, fmt.Errorf("failed to query source payments table: %w", err)
	}
	defer rows.Close()

	syncCount := 0
	for rows.Next() {
		var projectName, tahap, invoice string
		var jumlah, dpp, ppn, pph float64
		var tanggal sql.NullTime

		err = rows.Scan(&projectName, &tahap, &invoice, &jumlah, &dpp, &ppn, &pph, &tanggal)
		if err != nil {
			return syncCount, fmt.Errorf("failed to scan source row: %w", err)
		}

		projectName = strings.TrimSpace(projectName)
		invoice = strings.TrimSpace(invoice)

		if projectName == "" || invoice == "" || jumlah <= 0 {
			continue // skip empty/invalid payments
		}

		// 1. Find or dynamically create client
		var clientID int
		var pricePackage string
		err = db.DB.QueryRow("SELECT id, price_package FROM clients WHERE name = ?", projectName).Scan(&clientID, &pricePackage)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Create client dynamically using default properties
				err = CreateClient(projectName, "", "billing@madoo.id", "", "", "Custom Plan", "", "DKI Jakarta")
				if err != nil {
					return syncCount, fmt.Errorf("failed to auto-create client '%s': %w", projectName, err)
				}
				// Fetch client ID and price package
				err = db.DB.QueryRow("SELECT id, price_package FROM clients WHERE name = ?", projectName).Scan(&clientID, &pricePackage)
				if err != nil {
					return syncCount, fmt.Errorf("failed to retrieve ID for created client: %w", err)
				}
			} else {
				return syncCount, fmt.Errorf("database client check failed: %w", err)
			}
		}

		// 2. Check if invoice/order_id already exists locally in payments
		var existingID int
		err = db.DB.QueryRow("SELECT id FROM payments WHERE order_id = ?", invoice).Scan(&existingID)
		
		packageName := pricePackage
		if packageName == "" {
			packageName = "Custom Plan"
		}

		var tVal interface{}
		if tanggal.Valid {
			tVal = tanggal.Time
		} else {
			tVal = nil
		}

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Insert new payment (with empty snap token/url - will be generated on-demand)
				insertQuery := `
					INSERT INTO payments (client_id, order_id, amount, package_name, status, tahap, dpp, ppn, pph, tanggal_invoice) 
					VALUES (?, ?, ?, ?, 'Pending', ?, ?, ?, ?, ?)`
				_, err = db.DB.Exec(insertQuery, clientID, invoice, jumlah, packageName, tahap, dpp, ppn, pph, tVal)
				if err != nil {
					return syncCount, fmt.Errorf("failed to insert synced payment '%s': %w", invoice, err)
				}
				syncCount++
			} else {
				return syncCount, fmt.Errorf("database check for payment failed: %w", err)
			}
		} else {
			// Update existing payment
			updateQuery := `
				UPDATE payments 
				SET client_id = ?, amount = ?, package_name = ?, tahap = ?, dpp = ?, ppn = ?, pph = ?, tanggal_invoice = ? 
				WHERE id = ?`
			_, err = db.DB.Exec(updateQuery, clientID, jumlah, packageName, tahap, dpp, ppn, pph, tVal, existingID)
			if err != nil {
				return syncCount, fmt.Errorf("failed to update synced payment '%s': %w", invoice, err)
			}
		}
	}

	return syncCount, nil
}

// DeletePayment deletes a payment record from the database by its order ID
func DeletePayment(orderID string) error {
	query := `DELETE FROM payments WHERE order_id = ?`
	_, err := db.DB.Exec(query, orderID)
	return err
}

type PaymentStats struct {
	SuccessAmount float64 `json:"success_amount"`
	SuccessCount  int     `json:"success_count"`
	PendingAmount float64 `json:"pending_amount"`
	PendingCount  int     `json:"pending_count"`
	FailedAmount  float64 `json:"failed_amount"`
	FailedCount   int     `json:"failed_count"`
}

// GetPaymentStats retrieves sum of amounts and transaction counts grouped by status
func GetPaymentStats() (*PaymentStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'Success' THEN amount ELSE 0 END), 0) as success_amount,
			COUNT(CASE WHEN status = 'Success' THEN 1 END) as success_count,
			COALESCE(SUM(CASE WHEN status = 'Pending' THEN amount ELSE 0 END), 0) as pending_amount,
			COUNT(CASE WHEN status = 'Pending' THEN 1 END) as pending_count,
			COALESCE(SUM(CASE WHEN status IN ('Failed', 'Expired') THEN amount ELSE 0 END), 0) as failed_amount,
			COUNT(CASE WHEN status IN ('Failed', 'Expired') THEN 1 END) as failed_count
		FROM payments`

	var stats PaymentStats
	err := db.DB.QueryRow(query).Scan(
		&stats.SuccessAmount, &stats.SuccessCount,
		&stats.PendingAmount, &stats.PendingCount,
		&stats.FailedAmount, &stats.FailedCount,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

