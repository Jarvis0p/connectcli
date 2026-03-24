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

type PunchInClient struct {
	client *http.Client
}

type PunchInLocation struct {
	Latitude         float64  `json:"latitude"`
	Longitude        float64  `json:"longitude"`
	Altitude         *float64 `json:"altitude"`
	Timestamp        int64    `json:"timestamp"`
	Accuracy         int      `json:"accuracy"`
	AltitudeAccuracy *int     `json:"altitudeAccuracy"`
	Speed            *float64 `json:"speed"`
	Address          string   `json:"address"`
}

type PunchDetails struct {
	Source  string `json:"source"`
	ShiftID string `json:"shiftId"`
}

type PunchInRequest struct {
	ObjectID     int             `json:"objectId"`
	TagHierarchy []string        `json:"tagHierarchy"`
	Timezone     string          `json:"timezone"`
	Location     PunchInLocation `json:"location"`
	PunchDetails PunchDetails    `json:"punchDetails"`
	Spirit       string          `json:"_spirit"`
}

type PunchInResponse struct {
	Code          int                    `json:"code"`
	Message       string                 `json:"message"`
	Data          map[string]interface{} `json:"data"`
	RequestID     string                 `json:"requestId"`
	ServerVersion string                 `json:"serverVersion"`
}

func NewPunchInClient() *PunchInClient {
	return &PunchInClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *PunchInClient) PunchIn(creds *credentials.Credentials, objectID int, clientID string) (*PunchInResponse, error) {
	request := PunchInRequest{
		ObjectID:     objectID,
		TagHierarchy: []string{clientID},
		Timezone:     "Asia/Kolkata",
		Location: PunchInLocation{
			Latitude:  28.6188345885139,
			Longitude: 77.37885834499728,
			Timestamp: time.Now().Unix(),
			Accuracy:  35,
			Address:   "Sector 63 Road 130، 201309 Noida، India",
		},
		PunchDetails: PunchDetails{
			Source:  "timeclock",
			ShiftID: "",
		},
		Spirit: creds.CSRF,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/UserDashboard/PunchClock/Punch/In/", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", creds.Session))
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	resp, err := p.client.Do(req)
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

	var response PunchInResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}
