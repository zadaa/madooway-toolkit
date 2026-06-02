package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"task-manager-go/config"
)

type ClickUpTaskRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ClickUpTaskResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// CreateTaskInClickUp sends a POST request to ClickUp API to create a new task.
// If successful, it returns the URL of the created task.
func CreateTaskInClickUp(title, description string) (string, error) {
	token := strings.TrimSpace(config.AppConfig.ClickUpToken)
	listID := strings.TrimSpace(config.AppConfig.ClickUpListID)

	// Sanitize listID in case user copy-pasted a URL containing query parameters (e.g., ?pr=...)
	if idx := strings.Index(listID, "?"); idx != -1 {
		listID = listID[:idx]
	}
	listID = strings.TrimSpace(listID)

	if token == "" || listID == "" {
		log.Printf("Warning: ClickUp integration is disabled (token empty? %t, listID empty? %t)", token == "", listID == "")
		return "", nil
	}

	// Redacted token for safe logging
	safeToken := ""
	if len(token) > 8 {
		safeToken = token[:4] + "..." + token[len(token)-4:]
	} else {
		safeToken = "Invalid/Short Token"
	}
	log.Printf("Attempting ClickUp task creation on list %s with token %s", listID, safeToken)

	apiURL := fmt.Sprintf("https://api.clickup.com/api/v2/list/%s/task", listID)

	payload := ClickUpTaskRequest{
		Name:        title,
		Description: description,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling ClickUp payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request for ClickUp: %w", err)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	// Set timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request to ClickUp API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("clickup API returned error code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var responseData ClickUpTaskResponse
	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling ClickUp API response: %w", err)
	}

	log.Printf("Successfully created ClickUp task with ID: %s", responseData.ID)
	return responseData.URL, nil
}
