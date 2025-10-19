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

type ShiftRequestClient struct {
	client *http.Client
}

type ShiftRequest struct {
	TagHierarchy     []string `json:"tagHierarchy"`
	PunchInTime      int64    `json:"punchInTime"`
	PunchOutTime     int64    `json:"punchOutTime"`
	Note             string   `json:"note"`
	ShiftAttachments []string `json:"shiftAttachments"`
	ObjectID         int      `json:"objectId"`
	Timezone         string   `json:"timezone"`
	Spirit           string   `json:"_spirit"`
}

type ShiftRequestResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RawData map[string]interface{} `json:"-"`
	} `json:"data"`
	RequestID     string `json:"requestId"`
	ServerVersion string `json:"serverVersion"`
}

func NewShiftRequestClient() *ShiftRequestClient {
	return &ShiftRequestClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AddShiftRequest sends a shift request to the Connecteam API
func (s *ShiftRequestClient) AddShiftRequest(creds *credentials.Credentials, objectID int, request *ShiftRequest) (*ShiftRequestResponse, error) {
	// Set the spirit (CSRF token)
	request.Spirit = creds.CSRF
	request.ObjectID = objectID

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/UserDashboard/PunchClock/ShiftRequest/", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", creds.Session))
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	// Make the request
	resp, err := s.client.Do(req)
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
	var response ShiftRequestResponse
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
