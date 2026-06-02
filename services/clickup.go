package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	token := config.AppConfig.ClickUpToken
	listID := config.AppConfig.ClickUpListID

	if token == "" || listID == "" {
		log.Println("Warning: ClickUp integration is disabled (token or list ID is empty in configuration)")
		return "", nil
	}

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
