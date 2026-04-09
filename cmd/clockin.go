package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/notifications"
	"connectcli/internal/paths"
	"connectcli/internal/punchmonitor"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var (
	clockinClientFlag string
	clockinPeriodFlag string
)

var clockinCmd = &cobra.Command{
	Use:   "clockin",
	Short: "Clock in to a Connecteam client and start Slack reminders",
	Long: `Clock in to a client and send Slack notifications on an interval while you stay clocked in.

Client (-c) can be a UUID or a path under clients/ (e.g. Securify Internal.json).

Optional -p sets the Slack reminder interval as hh:mm (default 00:10 = 10 minutes).

Examples:
  connectcli clockin -c d0f16214-1112-0bfb-3db7-910e6cf99258
  connectcli clockin -c "Securify Internal.json"
  connectcli clockin -c clients/Keyo.json -p 00:15`,
	RunE: runClockIn,
}

func runClockIn(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(clockinClientFlag) == "" {
		return fmt.Errorf("client is required (-c)")
	}

	interval := 10 * time.Minute
	if strings.TrimSpace(clockinPeriodFlag) != "" {
		d, err := utils.ParseHHMMPeriod(clockinPeriodFlag)
		if err != nil {
			return err
		}
		interval = d
	}

	clientID, err := resolveClockinClientID(clockinClientFlag)
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

	punchClient := api.NewPunchInClient()
	fmt.Println("Sending punch in request...")

	punchResp, err := punchClient.PunchIn(creds, objectID, clientID)
	if err != nil {
		return fmt.Errorf("failed to punch in: %w", err)
	}

	fmt.Printf("Punch in successful! (Request ID: %s)\n", punchResp.RequestID)

	statusClient := api.NewClockStatusClient()
	statusResp, err := statusClient.GetStatus(creds, objectID)
	if err != nil {
		return fmt.Errorf("failed to get clock status after punch in: %w", err)
	}

	clientName := statusResp.ClientName()
	if clientName == "" {
		clientName = clientID
	}

	loc, locErr := time.LoadLocation("Asia/Kolkata")
	if locErr != nil {
		loc = time.Local
	}
	punchInTS := statusResp.PunchInTimestamp()
	elapsed := time.Since(time.Unix(punchInTS, 0))
	totalToday, thErr := utils.TotalHoursTodayIncludingOpenShift(creds, objectID, loc, statusResp.OpenPunchID(), elapsed)
	if thErr != nil {
		totalToday = elapsed.Hours()
	}

	slack := notifications.NewSlackClient(creds.SlackWebhook)
	msg := fmt.Sprintf("clocked in %s for %s\ntotal hours today: %.2f h", clientName, notifications.FormatDuration(elapsed), totalToday)
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

	if err := punchmonitor.Spawn(interval); err != nil {
		return fmt.Errorf("failed to start background punch monitor: %w", err)
	}

	logPath, _ := punchmonitor.LogPath()
	periodLabel := utils.FormatDurationAsHHMM(interval)
	fmt.Printf("\nBackground punch monitor started (Slack every %s). Log: %s\n", periodLabel, logPath)
	fmt.Println("It stops when you run clockout or clock out in Connecteam.")

	return nil
}

func resolveClockinClientID(input string) (string, error) {
	if looksLikeClockinUUID(input) {
		return input, nil
	}

	path, err := paths.ResolveClientJSONPath(input)
	if err != nil {
		return "", err
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

func looksLikeClockinUUID(s string) bool {
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

func init() {
	clockinCmd.Flags().StringVarP(&clockinClientFlag, "client", "c", "", "Client UUID or JSON file name/path under clients/")
	clockinCmd.Flags().StringVarP(&clockinPeriodFlag, "period", "p", "", "Slack reminder interval as hh:mm (default 00:10)")
	_ = clockinCmd.MarkFlagRequired("client")

	rootCmd.AddCommand(clockinCmd)
}
