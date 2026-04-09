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

type ClockStatusClient struct {
	client *http.Client
}

type ClockStatusRequest struct {
	ObjectID        int    `json:"objectId"`
	DefaultTimezone string `json:"defaultTimezone"`
	Spirit          string `json:"_spirit"`
}

type ClockStatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		OpenPunch *OpenPunch `json:"openPunch"`
	} `json:"data"`
}

type OpenPunch struct {
	PunchIn struct {
		TimestampWithTimezone struct {
			Timestamp int64  `json:"timestamp"`
			Timezone  string `json:"timezone"`
		} `json:"timestampWithTimezone"`
	} `json:"punchIn"`
	PunchTag struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"punchTag"`
	PunchID      string   `json:"punchId"`
	TagHierarchy []string `json:"tagHierarchy"`
}

func NewClockStatusClient() *ClockStatusClient {
	return &ClockStatusClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ClockStatusClient) GetStatus(creds *credentials.Credentials, objectID int) (*ClockStatusResponse, error) {
	reqBody := ClockStatusRequest{
		ObjectID:        objectID,
		DefaultTimezone: "Asia/Kolkata",
		Spirit:          creds.CSRF,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/UserDashboard/PunchClock/Data/", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", creds.Session))
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response ClockStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}

// IsClockedIn returns true if there is an active open punch.
func (r *ClockStatusResponse) IsClockedIn() bool {
	return r.Data.OpenPunch != nil
}

// PunchInTimestamp returns the punch-in unix timestamp.
func (r *ClockStatusResponse) PunchInTimestamp() int64 {
	if r.Data.OpenPunch == nil {
		return 0
	}
	return r.Data.OpenPunch.PunchIn.TimestampWithTimezone.Timestamp
}

// ClientName returns the tag/client name from the open punch.
func (r *ClockStatusResponse) ClientName() string {
	if r.Data.OpenPunch == nil {
		return ""
	}
	return r.Data.OpenPunch.PunchTag.Name
}

// OpenPunchID returns the open punch id, if any.
func (r *ClockStatusResponse) OpenPunchID() string {
	if r.Data.OpenPunch == nil {
		return ""
	}
	return r.Data.OpenPunch.PunchID
}
