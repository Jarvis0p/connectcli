package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/notifications"
	"connectcli/internal/punchmonitor"
	"connectcli/internal/utils"
)

var punchInFlag string

func runPunchIn() error {
	clientID, err := resolveClientID(punchInFlag)
	if err != nil {
		return fmt.Errorf("failed to resolve client: %w", err)
	}

	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds.SlackWebhook == "" {
		return fmt.Errorf("slack_webhook not found in credentials file. Add slack_webhook=<url> to ~/.connectcli/credentials")
	}

	if err := utils.EnsureObjectID(); err != nil {
		return fmt.Errorf("failed to ensure object ID: %w", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	objectID, err := strconv.Atoi(cfg.PunchClockObjectID)
	if err != nil {
		return fmt.Errorf("invalid punch clock object ID: %w", err)
	}

	// Punch in
	punchClient := api.NewPunchInClient()
	fmt.Println("Sending punch in request...")

	punchResp, err := punchClient.PunchIn(creds, objectID, clientID)
	if err != nil {
		return fmt.Errorf("failed to punch in: %w", err)
	}

	fmt.Printf("Punch in successful! (Request ID: %s)\n", punchResp.RequestID)

	// Fetch clock status to get the client name from the API
	statusClient := api.NewClockStatusClient()
	statusResp, err := statusClient.GetStatus(creds, objectID)
	if err != nil {
		return fmt.Errorf("failed to get clock status after punch in: %w", err)
	}

	clientName := statusResp.ClientName()
	if clientName == "" {
		clientName = clientID
	}

	// Send initial Slack notification
	slack := notifications.NewSlackClient(creds.SlackWebhook)
	msg := fmt.Sprintf("clocked-in %s", clientName)
	if err := slack.Send(msg); err != nil {
		fmt.Printf("Warning: failed to send Slack notification: %v\n", err)
	} else {
		fmt.Printf("Slack: %s\n", msg)
	}

	if creds.SlackUserToken != "" {
		if err := notifications.SetSlackClockedInStatus(creds.SlackUserToken, clientName); err != nil {
			fmt.Printf("Warning: failed to update Slack status: %v\n", err)
		} else {
			fmt.Printf("Slack status set: %s\n", clientName)
		}
	}

	// Background monitor: every 10 min Slack + stop if no longer clocked in (e.g. UI clock-out)
	if err := punchmonitor.Spawn(); err != nil {
		return fmt.Errorf("failed to start background punch monitor: %w", err)
	}

	logPath, _ := punchmonitor.LogPath()
	fmt.Printf("\nBackground punch monitor started (10 min interval). Log: %s\n", logPath)
	fmt.Println("It stops automatically when you punch out with -o or clock out in Connecteam.")

	return nil
}

func resolveClientID(input string) (string, error) {
	if looksLikeUUID(input) {
		return input, nil
	}

	path := input
	if !strings.HasPrefix(path, "clients/") && !strings.HasPrefix(path, "./clients/") {
		path = filepath.Join("clients", path)
	}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read client file %s: %w", path, err)
	}

	var client api.Client
	if err := json.Unmarshal(fileContent, &client); err != nil {
		return "", fmt.Errorf("failed to parse client file %s: %w", path, err)
	}

	if client.ID == "" {
		return "", fmt.Errorf("client ID not found in file %s", path)
	}

	return client.ID, nil
}

func looksLikeUUID(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}
