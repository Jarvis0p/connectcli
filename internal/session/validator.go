package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"connectcli/internal/credentials"
)

type Validator struct {
	client *http.Client
}

type ValidateRequest struct {
	Spirit string `json:"_spirit"`
}

func NewValidator() *Validator {
	return &Validator{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateSession checks if the provided session is valid by making a request to the Connecteam API
func (v *Validator) ValidateSession(creds *credentials.Credentials) (bool, error) {
	// Prepare the request body
	reqBody := ValidateRequest{
		Spirit: creds.CSRF,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/IsLoggedIn/", bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", creds.Session))
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	// Make the request
	resp, err := v.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging (optional)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	switch resp.StatusCode {
	case 200:
		return true, nil
	case 401:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
}
