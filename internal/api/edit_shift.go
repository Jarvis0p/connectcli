package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"connectcli/internal/credentials"
)

const (
	mobileShiftRequestVersion = "9.1.7.132"
	mobileShiftRequestURL     = "https://app.connecteam.com/api/Mobile/PunchClock/ShiftRequest/"
)

// EditShiftMobileRequest is the JSON body for PUT Mobile/PunchClock/ShiftRequest.
type EditShiftMobileRequest struct {
	ObjectID         int      `json:"objectId"`
	PunchID          string   `json:"punchId"`
	TagID            string   `json:"tagId"`
	IsEmptyState     bool     `json:"isEmptyState"`
	Timezone         string   `json:"timezone"`
	PunchInTime      int64    `json:"punchInTime"`
	PunchOutTime     int64    `json:"punchOutTime"`
	ShiftAttachments []string `json:"shiftAttachments"`
	Note             string   `json:"note"`
	ApprovalNote     string   `json:"approvalNote"`
	TagIDHierarchy   []string `json:"tagIdHierarchy"`
}

// EditShiftMobileResponse is a minimal parse of the API response.
type EditShiftMobileResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newEditShiftHTTPClient() *http.Client {
	return &http.Client{Timeout: 45 * time.Second}
}

// PutEditShift sends the edit shift request (Mobile API).
func PutEditShift(creds *credentials.Credentials, objectID int, punchID, tagID string, punchInUnix, punchOutUnix int64, note string) (*EditShiftMobileResponse, error) {
	body := EditShiftMobileRequest{
		ObjectID:         objectID,
		PunchID:          punchID,
		TagID:            tagID,
		IsEmptyState:     false,
		Timezone:         "Asia/Kolkata",
		PunchInTime:      punchInUnix,
		PunchOutTime:     punchOutUnix,
		ShiftAttachments: []string{},
		Note:             note,
		ApprovalNote:     "",
		TagIDHierarchy:   []string{tagID},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	u := fmt.Sprintf("%s?version=%s", mobileShiftRequestURL, mobileShiftRequestVersion)
	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", creds.Session))
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	hc := newEditShiftHTTPClient()
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var out EditShiftMobileResponse
	if len(strings.TrimSpace(string(raw))) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(raw))
		}
	} else {
		out.Code = 200
		out.Message = "ok"
	}
	return &out, nil
}
