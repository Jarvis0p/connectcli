package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const slackAPIBase = "https://slack.com/api"

type slackProfileSetRequest struct {
	Profile slackProfileFields `json:"profile"`
}

type slackProfileFields struct {
	StatusText       string `json:"status_text"`
	StatusEmoji      string `json:"status_emoji"`
	StatusExpiration int64  `json:"status_expiration,omitempty"`
}

type slackAPIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// SetSlackClockedInStatus sets Slack custom status to the client name (user token xoxp- or bot xoxb- with users.profile:write as applicable).
func SetSlackClockedInStatus(userToken, clientName string) error {
	if strings.TrimSpace(userToken) == "" {
		return nil
	}

	text := strings.TrimSpace(clientName)
	if text == "" {
		text = "Clocked in"
	}
	text = truncateStatusText(text, 100)

	body, err := json.Marshal(slackProfileSetRequest{
		Profile: slackProfileFields{
			StatusText:       text,
			StatusEmoji:      ":clock3:",
			StatusExpiration: 0,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal slack profile: %w", err)
	}

	return postSlackProfileSet(userToken, body)
}

// ClearSlackUserStatus clears the authenticated user's custom status.
func ClearSlackUserStatus(userToken string) error {
	if strings.TrimSpace(userToken) == "" {
		return nil
	}

	body, err := json.Marshal(slackProfileSetRequest{
		Profile: slackProfileFields{
			StatusText:  "",
			StatusEmoji: "",
		},
	})
	if err != nil {
		return fmt.Errorf("marshal slack profile: %w", err)
	}

	return postSlackProfileSet(userToken, body)
}

func postSlackProfileSet(userToken string, jsonBody []byte) error {
	req, err := http.NewRequest("POST", slackAPIBase+"/users.profile.set", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("slack API request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read slack response: %w", err)
	}

	var api slackAPIResponse
	if err := json.Unmarshal(raw, &api); err != nil {
		return fmt.Errorf("parse slack response: %w (body: %s)", err, string(raw))
	}
	if !api.OK {
		return fmt.Errorf("slack API error: %s", api.Error)
	}
	return nil
}

func truncateStatusText(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
