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

type ClientsClient struct {
	client *http.Client
}

type ClientsRequest struct {
	ObjectID        int    `json:"objectId"`
	DefaultTimezone string `json:"defaultTimezone"`
	Spirit          string `json:"_spirit"`
}

type ClientsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RawData map[string]interface{} `json:"-"`
	} `json:"data"`
	RequestID     string `json:"requestId"`
	ServerVersion string `json:"serverVersion"`
}

type Client struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewClientsClient() *ClientsClient {
	return &ClientsClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchClients fetches client data from the Connecteam API
func (c *ClientsClient) FetchClients(creds *credentials.Credentials, objectID int) (*ClientsResponse, error) {
	// Prepare the request body
	reqBody := ClientsRequest{
		ObjectID:        objectID,
		DefaultTimezone: "Asia/Kolkata",
		Spirit:          creds.CSRF,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/UserDashboard/PunchClock/Data/", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", creds.Session))
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	// Make the request
	resp, err := c.client.Do(req)
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
	var response ClientsResponse
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

// ExtractClients extracts client information from the response
func (c *ClientsClient) ExtractClients(response *ClientsResponse) ([]Client, error) {
	var clients []Client

	// Navigate through the nested structure to find clients
	// The exact structure will depend on the API response
	// We'll look for common patterns in the raw data
	rawData := response.Data.RawData

	// Try to find clients in different possible locations
	if clientsData, ok := rawData["clients"].([]interface{}); ok {
		for _, clientInterface := range clientsData {
			if clientMap, ok := clientInterface.(map[string]interface{}); ok {
				id, _ := clientMap["id"].(string)
				name, _ := clientMap["name"].(string)
				if id != "" && name != "" {
					clients = append(clients, Client{ID: id, Name: name})
				}
			}
		}
	}

	// If not found in "clients", try other possible locations
	if len(clients) == 0 {
		// Look for any array that might contain client data
		for _, value := range rawData {
			if array, ok := value.([]interface{}); ok {
				for _, item := range array {
					if itemMap, ok := item.(map[string]interface{}); ok {
						// Check if this looks like a client object
						if id, hasID := itemMap["id"].(string); hasID {
							if name, hasName := itemMap["name"].(string); hasName {
								// Additional check: make sure it's not empty and looks like a UUID
								if id != "" && name != "" && len(id) > 10 {
									clients = append(clients, Client{ID: id, Name: name})
								}
							}
						}
					}
				}
			}
		}
	}

	return clients, nil
}
