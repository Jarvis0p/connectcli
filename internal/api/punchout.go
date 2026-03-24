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

type PunchOutClient struct {
	client *http.Client
}

type ShiftAttachment struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Signature interface{} `json:"signature,omitempty"`
	Images    []string    `json:"images,omitempty"`
	FreeText  interface{} `json:"freeText,omitempty"`
}

type PunchOutRequest struct {
	ObjectID         int               `json:"objectId"`
	ShiftAttachments []ShiftAttachment `json:"shiftAttachments"`
	Note             string            `json:"note"`
	Timezone         string            `json:"timezone"`
	Location         PunchInLocation   `json:"location"`
	PunchDetails     PunchDetails      `json:"punchDetails"`
	Spirit           string            `json:"_spirit"`
}

type PunchOutResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TotalHours    interface{} `json:"totalHours"`
		TagHierarchy  []string    `json:"tagHierarchy"`
		PunchIn       *PunchTime  `json:"punchIn"`
		PunchOut      *PunchTime  `json:"punchOut"`
	} `json:"data"`
	RequestID     string `json:"requestId"`
	ServerVersion string `json:"serverVersion"`
}

type PunchTime struct {
	TimestampWithTimezone struct {
		Timestamp int64  `json:"timestamp"`
		Timezone  string `json:"timezone"`
	} `json:"timestampWithTimezone"`
}

type PunchConfirmRequest struct {
	ObjectID int    `json:"objectId"`
	Type     string `json:"type"`
	Spirit   string `json:"_spirit"`
}

type PunchConfirmResponse struct {
	Code          int    `json:"code"`
	Message       string `json:"message"`
	RequestID     string `json:"requestId"`
	ServerVersion string `json:"serverVersion"`
}

func NewPunchOutClient() *PunchOutClient {
	return &PunchOutClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *PunchOutClient) PunchOut(creds *credentials.Credentials, objectID int, note string) (*PunchOutResponse, error) {
	request := PunchOutRequest{
		ObjectID: objectID,
		ShiftAttachments: []ShiftAttachment{
			{ID: "30fb7944-2165-4ac7-9db8-5dcd9fe77861", Type: "signature", Signature: map[string]interface{}{}},
			{ID: "3f3919ff-eba9-42cc-b6db-d0dc3a17ef3e", Type: "image", Images: []string{}},
			{ID: "fa4c91d9-9d5e-455c-3742-6b4533c6ff95", Type: "freeText"},
		},
		Note:     note,
		Timezone: "Asia/Kolkata",
		Location: PunchInLocation{
			Latitude:  28.618908813584596,
			Longitude: 77.37878032996443,
			Timestamp: time.Now().UnixMilli(),
			Accuracy:  35,
			Address:   "Springboard A-130 91، 201309 Noida، India",
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

	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/UserDashboard/PunchClock/Punch/Out/", bytes.NewBuffer(jsonBody))
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

	var response PunchOutResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}

func (p *PunchOutClient) Confirm(creds *credentials.Credentials, objectID int) (*PunchConfirmResponse, error) {
	request := PunchConfirmRequest{
		ObjectID: objectID,
		Type:     "punch",
		Spirit:   creds.CSRF,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://app.connecteam.com/api/UserDashboard/PunchClock/Punch/Confirm/", bytes.NewBuffer(jsonBody))
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

	var response PunchConfirmResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}
