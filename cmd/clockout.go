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

	"github.com/spf13/cobra"
)

var clockoutNoteFlag string

var clockoutCmd = &cobra.Command{
	Use:   "clockout",
	Short: "Clock out from the current Connecteam shift",
	Long: `Clock out from your current shift, confirm the punch, notify Slack, and clear Slack status.

Example:
  connectcli clockout -n 'huddle with team'`,
	RunE: runClockOut,
}

func runClockOut(cmd *cobra.Command, args []string) error {
	if clockoutNoteFlag == "" {
		return fmt.Errorf("note is required. Use -n '<your note>'")
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

	if err := punchmonitor.Stop(); err != nil {
		return fmt.Errorf("failed to stop punch monitor: %w", err)
	}

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

	fmt.Printf("Clocking out from: %s\n", clientName)
	fmt.Printf("Note: %s\n\n", clockoutNoteFlag)

	punchOutClient := api.NewPunchOutClient()
	fmt.Println("Sending clock out request...")

	outResp, err := punchOutClient.PunchOut(creds, objectID, clockoutNoteFlag)
	if err != nil {
		return fmt.Errorf("failed to clock out: %w", err)
	}

	fmt.Printf("Clock out successful! (Request ID: %s)\n", outResp.RequestID)

	fmt.Println("Confirming punch...")

	confirmResp, err := punchOutClient.Confirm(creds, objectID)
	if err != nil {
		return fmt.Errorf("failed to confirm punch: %w", err)
	}

	fmt.Printf("Punch confirmed! (Request ID: %s)\n", confirmResp.RequestID)

	punchInTime := time.Unix(punchInTimestamp, 0)
	elapsed := time.Since(punchInTime)

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

func init() {
	clockoutCmd.Flags().StringVarP(&clockoutNoteFlag, "note", "n", "", "Note for this shift (required)")
	_ = clockoutCmd.MarkFlagRequired("note")

	rootCmd.AddCommand(clockoutCmd)
}
