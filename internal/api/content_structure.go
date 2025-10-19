package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"connectcli/internal/credentials"
)

type ContentStructureClient struct {
	client *http.Client
}

type ContentStructureResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Containers []struct {
			LPID   string `json:"lpid"`
			Assets []struct {
				LPID    string `json:"lpid"`
				Courses []struct {
					ID       int `json:"id"`
					Sections []struct {
						ID      int `json:"id"`
						Objects []struct {
							ID           int    `json:"id"`
							Completed    bool   `json:"completed"`
							Viewed       bool   `json:"viewed"`
							Name         string `json:"name"`
							Type         string `json:"type"`
							TimeModified int64  `json:"timeModified"`
							ObjectType   string `json:"objectType"`
						} `json:"objects"`
						Name    string `json:"name"`
						IconURL string `json:"iconUrl"`
					} `json:"sections"`
					Name         string `json:"name"`
					TimeModified int64  `json:"timeModified"`
					IconURL      string `json:"iconUrl"`
					LPID         string `json:"lpid"`
				} `json:"courses"`
				Name          string `json:"name"`
				Index         int    `json:"index"`
				DashboardType string `json:"dashboardType"`
				Icon          struct {
					URL            string   `json:"url"`
					Color          string   `json:"color"`
					GradientColors []string `json:"gradientColors"`
				} `json:"icon"`
				BadgeCount     int   `json:"badgeCount"`
				ActiveInMobile *bool `json:"activeInMobile"`
				IsActivated    bool  `json:"isActivated"`
				IsArchived     bool  `json:"isArchived"`
			} `json:"assets"`
			Name          string `json:"name"`
			Index         int    `json:"index"`
			DashboardType string `json:"dashboardType"`
			IconURL       string `json:"iconUrl"`
		} `json:"containers"`
	} `json:"data"`
	RequestID     string `json:"requestId"`
	ServerVersion string `json:"serverVersion"`
}

func NewContentStructureClient() *ContentStructureClient {
	return &ContentStructureClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchContentStructure fetches the content structure from the Connecteam API
func (c *ContentStructureClient) FetchContentStructure(creds *credentials.Credentials) (*ContentStructureResponse, error) {
	// Create the HTTP request
	req, err := http.NewRequest("GET", "https://app.connecteam.com/api/UserDashboard/ContentStructure/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
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
	var response ContentStructureResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}

// ExtractPunchClockObjectID extracts the punch clock object ID from the content structure
func (c *ContentStructureClient) ExtractPunchClockObjectID(response *ContentStructureResponse) (int, error) {
	for _, container := range response.Data.Containers {
		for _, asset := range container.Assets {
			if asset.LPID == "punchclock" {
				// Found punch clock asset, get the first object ID
				if len(asset.Courses) > 0 && len(asset.Courses[0].Sections) > 0 && len(asset.Courses[0].Sections[0].Objects) > 0 {
					return asset.Courses[0].Sections[0].Objects[0].ID, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("punch clock object not found in content structure")
}
