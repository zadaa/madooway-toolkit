package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"task-manager-go/config"
)

// SnapRequest structures the payload for Midtrans Snap API
type SnapRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CreditCard         CreditCard         `json:"credit_card"`
	CustomerDetails    CustomerDetails    `json:"customer_details"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type CreditCard struct {
	Secure bool `json:"secure"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// SnapResponse captures the response from Midtrans Snap API
type SnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

// CreateSnapTransaction calls Midtrans API to get payment token and redirect url
func CreateSnapTransaction(orderID string, grossAmount int64, clientName, clientEmail, clientPhone string) (string, string, error) {
	serverKey := config.AppConfig.MidtransServerKey
	if serverKey == "" {
		return "", "", errors.New("midtrans server key is not configured")
	}

	apiURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"
	if config.AppConfig.MidtransIsProduction {
		apiURL = "https://app.midtrans.com/snap/v1/transactions"
	}

	// Sanitize email and phone to prevent Midtrans 400 validation error
	cleanEmail := strings.TrimSpace(clientEmail)
	if cleanEmail == "" || !strings.Contains(cleanEmail, "@") {
		cleanEmail = "" // set to empty so it is omitted from JSON
	}

	cleanPhone := strings.TrimSpace(clientPhone)
	// Midtrans only accepts alphanumeric, space, plus, hyphen, parentheses for phone
	// If it's empty, we omit it
	if cleanPhone == "" {
		cleanPhone = ""
	}

	reqBody := SnapRequest{
		TransactionDetails: TransactionDetails{
			OrderID:     orderID,
			GrossAmount: grossAmount,
		},
		CreditCard: CreditCard{
			Secure: true,
		},
		CustomerDetails: CustomerDetails{
			FirstName: clientName,
			Email:     cleanEmail,
			Phone:     cleanPhone,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", "", fmt.Errorf("failed to create http request: %w", err)
	}

	// Midtrans Authorization is Basic with Base64 of (ServerKey + ":")
	authValue := base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+authValue)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("midtrans api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var snapResp SnapResponse
	err = json.Unmarshal(respBytes, &snapResp)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return snapResp.Token, snapResp.RedirectURL, nil
}

// VerifyWebhookSignature checks if the signature_key from Midtrans notification is authentic
// The formula: SHA512(order_id + status_code + gross_amount + server_key)
func VerifyWebhookSignature(orderID, statusCode, grossAmount, incomingSignature string) bool {
	serverKey := config.AppConfig.MidtransServerKey
	if serverKey == "" {
		return false
	}

	// Construct payload signature
	payload := orderID + statusCode + grossAmount + serverKey
	
	hasher := sha512.New()
	hasher.Write([]byte(payload))
	calculatedSignature := hex.EncodeToString(hasher.Sum(nil))

	return calculatedSignature == incomingSignature
}

// GetTransactionStatus retrieves the status of a transaction from Midtrans API
func GetTransactionStatus(orderID string) (string, string, error) {
	serverKey := config.AppConfig.MidtransServerKey
	if serverKey == "" {
		return "", "", errors.New("midtrans server key is not configured")
	}

	apiURL := fmt.Sprintf("https://api.sandbox.midtrans.com/v2/%s/status", orderID)
	if config.AppConfig.MidtransIsProduction {
		apiURL = fmt.Sprintf("https://api.midtrans.com/v2/%s/status", orderID)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create http request: %w", err)
	}

	authValue := base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+authValue)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("midtrans api status returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	type StatusResponse struct {
		TransactionStatus string `json:"transaction_status"`
		PaymentType       string `json:"payment_type"`
		FraudStatus       string `json:"fraud_status"`
	}

	var statusResp StatusResponse
	err = json.Unmarshal(respBytes, &statusResp)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	return statusResp.TransactionStatus, statusResp.PaymentType, nil
}

