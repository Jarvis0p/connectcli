package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"connectcli/internal/credentials"
)

type TimesheetClient struct {
	client *http.Client
}

type TimesheetRequest struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	ObjectID        string `json:"objectId"`
	DefaultTimezone string `json:"defaultTimezone"`
	Spirit          string `json:"_spirit"`
}

type TimesheetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		// Add specific fields based on the actual response structure
		// For now, we'll use a generic map to handle any response structure
		RawData map[string]interface{} `json:"-"`
	} `json:"data"`
	RequestID     string `json:"requestId"`
	ServerVersion string `json:"serverVersion"`
}

func NewTimesheetClient() *TimesheetClient {
	return &TimesheetClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchTimesheet fetches timesheet data for the specified date range
func (t *TimesheetClient) FetchTimesheet(creds *credentials.Credentials, objectID, startDate, endDate string) (*TimesheetResponse, error) {
	// Prepare the request body
	reqBody := TimesheetRequest{
		StartDate:       startDate,
		EndDate:         endDate,
		ObjectID:        objectID,
		DefaultTimezone: "Asia/Kolkata",
		Spirit:          creds.CSRF,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/UserDashboard/PunchClock/Timesheet/", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", creds.Session))
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	// Make the request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var response TimesheetResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Also parse the raw data for flexibility
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err == nil {
		if data, exists := rawData["data"]; exists {
			if dataMap, ok := data.(map[string]interface{}); ok {
				response.Data.RawData = dataMap
			}
		}
	}

	return &response, nil
}
