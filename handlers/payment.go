package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"task-manager-go/config"
	"task-manager-go/models"
	"task-manager-go/services"

	"rsc.io/pdf"
)

// PaymentPageData structures the variables sent to payments.html
type PaymentPageData struct {
	Payments  []models.Payment
	Clients   []models.Client
	ClientKey string
}

// ListPayments displays the transactions dashboard
func ListPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := models.GetAllPayments()
	if err != nil {
		log.Printf("Error fetching payments: %v", err)
		RenderTemplate(w, r, "payments.html", "Kelola Pembayaran", "payments", nil, "Gagal memuat riwayat pembayaran.", "")
		return
	}

	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients for payment creation: %v", err)
	}

	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	data := PaymentPageData{
		Payments:  payments,
		Clients:   clients,
		ClientKey: config.AppConfig.MidtransClientKey,
	}

	RenderTemplate(w, r, "payments.html", "Kelola Pembayaran", "payments", data, errorMsg, successMsg)
}

// CreatePayment processes form post and initiates Midtrans Snap API transaction
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/payments", http.StatusSeeOther)
		return
	}

	clientIDStr := r.FormValue("client_id")
	amountStr := r.FormValue("amount")
	packageName := strings.TrimSpace(r.FormValue("package_name"))

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Redirect(w, r, "/payments?error=Client+tidak+valid", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Redirect(w, r, "/payments?error=Jumlah+pembayaran+harus+lebih+dari+nol", http.StatusSeeOther)
		return
	}

	if packageName == "" {
		http.Redirect(w, r, "/payments?error=Nama+paket+harus+diisi", http.StatusSeeOther)
		return
	}

	// Fetch client details for Midtrans billing address / customer details
	clientObj, err := models.GetClientByID(clientID)
	if err != nil || clientObj == nil {
		http.Redirect(w, r, "/payments?error=Klien+tidak+ditemukan", http.StatusSeeOther)
		return
	}

	// Generate unique Order ID
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	orderID := fmt.Sprintf("INV-%d-%d", clientID, timestamp)

	// Call Midtrans Snap API
	token, snapURL, err := services.CreateSnapTransaction(
		orderID,
		int64(amount),
		clientObj.Name,
		clientObj.Email,
		clientObj.Phone,
	)
	if err != nil {
		log.Printf("Midtrans Snap creation failed: %v", err)
		http.Redirect(w, r, "/payments?error=Gagal+terhubung+ke+Midtrans:+"+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}

	// Create local payment record
	err = models.CreatePayment(clientID, orderID, amount, packageName, token, snapURL)
	if err != nil {
		log.Printf("Failed to record payment in database: %v", err)
		http.Redirect(w, r, "/payments?error=Gagal+menyimpan+data+pembayaran", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/payments?success=Pembayaran+berhasil+diinisiasi", http.StatusSeeOther)
}

// MidtransNotification represents the JSON payload posted by Midtrans webhook
type MidtransNotification struct {
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`
}

// MidtransWebhook handles payment status updates posted from Midtrans
func MidtransWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook error reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var notification MidtransNotification
	err = json.Unmarshal(bodyBytes, &notification)
	if err != nil {
		log.Printf("Webhook error parsing JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Received Midtrans Webhook for Order: %s | Status: %s | Amount: %s",
		notification.OrderID, notification.TransactionStatus, notification.GrossAmount)

	// Verify the webhook is authentic
	isValid := services.VerifyWebhookSignature(
		notification.OrderID,
		notification.StatusCode,
		notification.GrossAmount,
		notification.SignatureKey,
	)

	if !isValid {
		log.Printf("WARNING: Invalid signature received for Midtrans Webhook on order: %s", notification.OrderID)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status": "error", "message": "invalid signature"}`))
		return
	}

	// Map Midtrans transaction status to local status representation
	localStatus := "Pending"
	switch notification.TransactionStatus {
	case "capture", "settlement":
		if notification.TransactionStatus == "capture" && notification.FraudStatus == "challenge" {
			localStatus = "Pending"
		} else {
			localStatus = "Success"
		}
	case "pending":
		localStatus = "Pending"
	case "deny", "cancel":
		localStatus = "Failed"
	case "expire":
		localStatus = "Expired"
	}

	// Check if this payment exists locally
	payment, err := models.GetPaymentByOrderID(notification.OrderID)
	if err != nil {
		log.Printf("Error looking up payment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if payment == nil {
		log.Printf("Warning: Webhook received for order_id %s but no payment record found locally.", notification.OrderID)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status": "error", "message": "payment record not found"}`))
		return
	}

	// Update payment status in local database
	err = models.UpdatePaymentStatus(notification.OrderID, localStatus, notification.PaymentType)
	if err != nil {
		log.Printf("Error updating local payment status: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Payment status updated to %s for Order %s", localStatus, notification.OrderID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok", "message": "payment status updated successfully"}`))
}

// ParseInvoicePDF handles uploaded invoice PDF files and returns extracted fields as JSON
func ParseInvoicePDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit size to 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("invoice_pdf")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Ensure upload folder exists
	tempDir := "static/uploads/temp"
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Printf("Failed to create temp upload directory: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Save to temp file
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("temp-%d.pdf", time.Now().UnixNano()))
	tempFile, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		tempFile.Close()
		_ = os.Remove(tempFilePath) // clean up temp file
	}()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		log.Printf("Failed to save file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	tempFile.Close() // close immediately so we can open it with the reader

	// Parse PDF Text
	pdfReader, err := pdf.Open(tempFilePath)
	if err != nil {
		log.Printf("Failed to open PDF: %v", err)
		http.Error(w, "Failed to read PDF structure", http.StatusBadRequest)
		return
	}

	type PdfText struct {
		Text string
		X    float64
		Y    float64
	}

	var textElements []PdfText
	totalPage := pdfReader.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		page := pdfReader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		texts := page.Content().Text
		for _, t := range texts {
			textElements = append(textElements, PdfText{Text: t.S, X: t.X, Y: t.Y})
		}
	}

	// Group texts into rows based on Y coordinate with tolerance threshold
	// 1. Sort all elements by Y coordinate descending (top to bottom)
	sort.Slice(textElements, func(i, j int) bool {
		return textElements[i].Y > textElements[j].Y
	})

	// 2. Group into rows
	var rows [][]PdfText
	for _, el := range textElements {
		foundRow := false
		for idx, r := range rows {
			// Compare with Y of the first element in that row (within 4.0 tolerance)
			if math.Abs(r[0].Y-el.Y) < 4.0 {
				rows[idx] = append(rows[idx], el)
				foundRow = true
				break
			}
		}
		if !foundRow {
			rows = append(rows, []PdfText{el})
		}
	}

	// 3. Sort each row by X coordinate ascending (left to right)
	for i := range rows {
		sort.Slice(rows[i], func(j, k int) bool {
			return rows[i][j].X < rows[i][k].X
		})
	}

	// 4. Construct lines of text
	var lines []string
	for _, r := range rows {
		var rowText strings.Builder
		for idx, el := range r {
			if idx > 0 {
				rowText.WriteString(" ")
			}
			rowText.WriteString(el.Text)
		}
		lines = append(lines, rowText.String())
	}

	// Parse Invoice Number
	var orderID string
	noRegex := regexp.MustCompile(`(?i)no\.?\s*([0-9a-zA-Z\-/]+)`)
	for _, line := range lines {
		matches := noRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			orderID = strings.TrimSpace(matches[1])
			break
		}
	}

	// Parse Client Name (look for "Untuk" / "Kepada" and pick next line)
	var clientName string
	for i, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "untuk") || strings.Contains(lower, "kepada") || strings.Contains(lower, "billed to") {
			for j := i + 1; j < len(lines); j++ {
				trimmed := strings.TrimSpace(lines[j])
				if trimmed != "" {
					// Exclude vendor name (our own company)
					if !strings.Contains(strings.ToLower(trimmed), "visi bangun sejahtera") {
						clientName = trimmed
						break
					}
				}
			}
			if clientName != "" {
				break
			}
		}
	}
	
	// Clean up clientName (remove our own company name if it got merged horizontally)
	clientName = strings.ReplaceAll(clientName, "PT. VISI BANGUN SEJAHTERA", "")
	clientName = strings.ReplaceAll(clientName, "PT VISI BANGUN SEJAHTERA", "")
	clientName = strings.ReplaceAll(clientName, "PT. Visi Bangun Sejahtera", "")
	clientName = strings.ReplaceAll(clientName, "PT Visi Bangun Sejahtera", "")
	clientName = strings.TrimPrefix(clientName, ":")
	clientName = strings.TrimSpace(clientName)

	// Parse Package Name / Subject (look for "Perihal" or "Subject")
	var packageName string
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "perihal") || strings.Contains(lower, "subject") {
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				packageName = strings.TrimSpace(parts[1])
			} else {
				words := strings.Fields(line)
				if len(words) > 1 {
					packageName = strings.TrimSpace(strings.Join(words[1:], " "))
				}
			}
			break
		}
	}
	// Clean up packageName (remove vendor if merged horizontally)
	packageName = strings.ReplaceAll(packageName, "PT. VISI BANGUN SEJAHTERA", "")
	packageName = strings.ReplaceAll(packageName, "PT VISI BANGUN SEJAHTERA", "")
	packageName = strings.ReplaceAll(packageName, "PT. Visi Bangun Sejahtera", "")
	packageName = strings.ReplaceAll(packageName, "PT Visi Bangun Sejahtera", "")
	packageName = strings.TrimPrefix(packageName, ":")
	packageName = strings.TrimSpace(packageName)

	if packageName == "" {
		packageName = "HRMS Payroll Ngabsen" // default fallback
	}

	// Parse Total Amount (find the maximum number in any line containing total keywords)
	var amount float64
	amountRegex := regexp.MustCompile(`[0-9\.,]{4,}`)
	var maxAmount float64

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "total") || strings.Contains(lower, "jumlah") || strings.Contains(lower, "grand total") || strings.Contains(lower, "bayar") {
			matches := amountRegex.FindAllString(line, -1)
			for _, m := range matches {
				cleaned := m
				if strings.HasSuffix(cleaned, ",00") || strings.HasSuffix(cleaned, ".00") {
					cleaned = cleaned[:len(cleaned)-3]
				}
				cleaned = strings.ReplaceAll(cleaned, ".", "")
				cleaned = strings.ReplaceAll(cleaned, ",", "")
				val, err := strconv.ParseFloat(cleaned, 64)
				if err == nil && val > 0 {
					if val > maxAmount {
						maxAmount = val
					}
				}
			}
		}
	}
	amount = maxAmount

	// Respond with JSON
	response := map[string]interface{}{
		"success":      true,
		"order_id":     orderID,
		"client_name":  clientName,
		"amount":       amount,
		"package_name": packageName,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// SyncPayments handles the POST request to sync payments from the finance database
func SyncPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/payments", http.StatusSeeOther)
		return
	}

	month := strings.TrimSpace(r.FormValue("month"))
	if month == "" {
		http.Redirect(w, r, "/payments?error=Bulan+sinkronisasi+harus+diisi", http.StatusSeeOther)
		return
	}

	// month format validation: YYYY-MM
	matched, err := regexp.MatchString(`^\d{4}-\d{2}$`, month)
	if err != nil || !matched {
		http.Redirect(w, r, "/payments?error=Format+bulan+harus+YYYY-MM", http.StatusSeeOther)
		return
	}

	syncCount, err := models.SyncPaymentsFromRemote(month)
	if err != nil {
		log.Printf("Sync error: %v", err)
		http.Redirect(w, r, "/payments?error=Gagal+sinkronisasi:+"+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}

	msg := fmt.Sprintf("Berhasil+menyinkronkan+%d+pembayaran+untuk+bulan+%s", syncCount, month)
	http.Redirect(w, r, "/payments?success="+msg, http.StatusSeeOther)
}

// GetPaymentToken handles generating a Snap Token on-demand for synced payments
func GetPaymentToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	payment, err := models.GetPaymentByOrderID(orderID)
	if err != nil {
		log.Printf("Error getting payment by orderID: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if payment == nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	// If snap token already exists, just return it
	if payment.SnapToken != "" && payment.SnapURL != "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"snap_token": payment.SnapToken,
			"snap_url":   payment.SnapURL,
		})
		return
	}

	// Fetch client details for Midtrans
	clientObj, err := models.GetClientByID(payment.ClientID)
	if err != nil || clientObj == nil {
		http.Error(w, "Client not found", http.StatusBadRequest)
		return
	}

	// Call Midtrans Snap API
	token, snapURL, err := services.CreateSnapTransaction(
		payment.OrderID,
		int64(payment.Amount),
		clientObj.Name,
		clientObj.Email,
		clientObj.Phone,
	)
	if err != nil {
		log.Printf("Midtrans Snap creation failed: %v", err)
		http.Error(w, "Gagal terhubung ke Midtrans: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the payment record with token and url
	err = models.UpdatePaymentSnapToken(payment.OrderID, token, snapURL)
	if err != nil {
		log.Printf("Failed to update payment snap token in database: %v", err)
		http.Error(w, "Database update error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"snap_token": token,
		"snap_url":   snapURL,
	})
}

// RefreshPaymentStatus checks the latest status from Midtrans and updates the local database
func RefreshPaymentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/payments", http.StatusSeeOther)
		return
	}

	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	// Fetch transaction status from Midtrans
	midtransStatus, paymentType, err := services.GetTransactionStatus(orderID)
	if err != nil {
		log.Printf("Failed to fetch status from Midtrans for order %s: %v", orderID, err)
		
		// If Midtrans returns 404 (transaction not found on their side yet)
		if strings.Contains(err.Error(), "404") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success":      true,
				"status":       "Pending",
				"payment_type": "",
				"message":      "Transaksi belum diinisiasi di Midtrans.",
			})
			return
		}
		
		http.Error(w, "Gagal sinkron status dari Midtrans: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Map Midtrans transaction status to local status representation
	localStatus := "Pending"
	switch midtransStatus {
	case "capture", "settlement":
		localStatus = "Success"
	case "pending":
		localStatus = "Pending"
	case "deny", "cancel":
		localStatus = "Failed"
	case "expire":
		localStatus = "Expired"
	}

	// Update payment status in local database
	err = models.UpdatePaymentStatus(orderID, localStatus, paymentType)
	if err != nil {
		log.Printf("Error updating local payment status: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"status":       localStatus,
		"payment_type": paymentType,
	})
}

// DeletePayment handles the request to delete a payment record
func DeletePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/payments", http.StatusSeeOther)
		return
	}

	orderID := strings.TrimSpace(r.FormValue("order_id"))
	if orderID == "" {
		http.Redirect(w, r, "/payments?error=Order+ID+wajib+diisi", http.StatusSeeOther)
		return
	}

	err := models.DeletePayment(orderID)
	if err != nil {
		log.Printf("Error deleting payment: %v", err)
		http.Redirect(w, r, "/payments?error=Gagal+menghapus+tagihan", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/payments?success=Tagihan+berhasil+dihapus", http.StatusSeeOther)
}

// PaymentsDashboardData holds data for payments dashboard
type PaymentsDashboardData struct {
	Stats        *models.PaymentStats
	Payments     []models.Payment
	Clients      []models.Client
	PaymentsJSON template.JS
	ClientsJSON  template.JS
}

// ShowPaymentsDashboard displays a dashboard summary for payments
func ShowPaymentsDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := models.GetPaymentStats()
	if err != nil {
		log.Printf("Error fetching payment stats: %v", err)
		stats = &models.PaymentStats{}
	}

	payments, err := models.GetAllPayments()
	if err != nil {
		log.Printf("Error fetching payments history: %v", err)
		payments = []models.Payment{}
	}

	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients list: %v", err)
		clients = []models.Client{}
	}

	paymentsJSON, err := json.Marshal(payments)
	if err != nil {
		paymentsJSON = []byte("[]")
	}

	clientsJSON, err := json.Marshal(clients)
	if err != nil {
		clientsJSON = []byte("[]")
	}

	data := PaymentsDashboardData{
		Stats:        stats,
		Payments:     payments,
		Clients:      clients,
		PaymentsJSON: template.JS(paymentsJSON),
		ClientsJSON:  template.JS(clientsJSON),
	}

	RenderTemplate(w, r, "payments_dashboard.html", "Dashboard Pembayaran", "payments_dashboard", data, "", "")
}




