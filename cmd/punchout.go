package cmd

import (
	"fmt"
	"strconv"
	"time"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/notifications"
	"connectcli/internal/punchmonitor"
	"connectcli/internal/utils"
)

var (
	punchOutFlag  bool
	punchNoteFlag string
)

func runPunchOut() error {
	if punchNoteFlag == "" {
		return fmt.Errorf("note is required for punch out. Use -n '<your note>'")
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

	// Stop background punch monitor (10-min Slack poller) if running
	if err := punchmonitor.Stop(); err != nil {
		return fmt.Errorf("failed to stop punch monitor: %w", err)
	}

	// Get clock status first to retrieve client name and punch-in time
	statusClient := api.NewClockStatusClient()
	statusResp, err := statusClient.GetStatus(creds, objectID)
	if err != nil {
		return fmt.Errorf("failed to get clock status: %w", err)
	}

	if !statusResp.IsClockedIn() {
		return fmt.Errorf("you are not currently clocked in")
	}

	clientName := statusResp.ClientName()
	punchInTimestamp := statusResp.PunchInTimestamp()

	fmt.Printf("Punching out from: %s\n", clientName)
	fmt.Printf("Note: %s\n\n", punchNoteFlag)

	// Punch out
	punchOutClient := api.NewPunchOutClient()
	fmt.Println("Sending punch out request...")

	outResp, err := punchOutClient.PunchOut(creds, objectID, punchNoteFlag)
	if err != nil {
		return fmt.Errorf("failed to punch out: %w", err)
	}

	fmt.Printf("Punch out successful! (Request ID: %s)\n", outResp.RequestID)

	// Confirm
	fmt.Println("Confirming punch...")

	confirmResp, err := punchOutClient.Confirm(creds, objectID)
	if err != nil {
		return fmt.Errorf("failed to confirm punch: %w", err)
	}

	fmt.Printf("Punch confirmed! (Request ID: %s)\n", confirmResp.RequestID)

	// Calculate shift duration
	punchInTime := time.Unix(punchInTimestamp, 0)
	elapsed := time.Since(punchInTime)

	// Send Slack notification
	slack := notifications.NewSlackClient(creds.SlackWebhook)
	msg := fmt.Sprintf("Shift ended: %s %s", clientName, notifications.FormatDuration(elapsed))
	if err := slack.Send(msg); err != nil {
		fmt.Printf("Warning: failed to send Slack notification: %v\n", err)
	} else {
		fmt.Printf("Slack: %s\n", msg)
	}

	if creds.SlackUserToken != "" {
		if err := notifications.ClearSlackUserStatus(creds.SlackUserToken); err != nil {
			fmt.Printf("Warning: failed to clear Slack status: %v\n", err)
		} else {
			fmt.Println("Slack status cleared.")
		}
	}

	return nil
}
