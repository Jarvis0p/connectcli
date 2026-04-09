package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var (
	editShiftStart  string
	editShiftEnd    string
	editShiftNote   string
	editShiftClient string
	editShiftDate   string
)

var editshiftCmd = &cobra.Command{
	Use:   "editshift [punchId]",
	Short: "Edit an existing shift (Mobile API)",
	Long: `Update punch in/out times and note for a shift by punch id.

Times use HH:MM in Asia/Kolkata on the shift day (--date, default today).

The client/tag UUID is required for the API (-c / --client), same as addshift (UUID or clients/*.json).

Example:
  connectcli editshift 69ccd2f81185ec591cb99bb1 -c d0f16214-1112-0bfb-3db7-910e6cf99258 -s 09:30 -e 17:45 -n 'TECH-3185'
  connectcli editshift 69ccd2f81185ec591cb99bb1 -c clients/Keyo.json -d 15/03/26 -s 10:00 -e 18:00 -n 'note'`,
	Args: cobra.ExactArgs(1),
	RunE: runEditShift,
}

func runEditShift(cmd *cobra.Command, args []string) error {
	punchID := strings.TrimSpace(args[0])
	if punchID == "" {
		return fmt.Errorf("punch id is required")
	}
	if editShiftClient == "" {
		return fmt.Errorf("client is required (-c uuid or clients/*.json)")
	}
	if editShiftStart == "" || editShiftEnd == "" {
		return fmt.Errorf("-s and -e (HH:MM) are required")
	}
	if editShiftNote == "" {
		return fmt.Errorf("-n note is required")
	}

	tagID, err := resolveEditShiftTagID(editShiftClient)
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return err
	}

	dateStr := strings.TrimSpace(editShiftDate)
	if dateStr == "" {
		dateStr = "today"
	}

	var day time.Time
	dl := strings.ToLower(dateStr)
	if dl == "today" {
		day = time.Now().In(loc)
	} else if dl == "yesterday" {
		day = time.Now().In(loc).AddDate(0, 0, -1)
	} else {
		start, _, err := utils.ParseDateRange(dateStr)
		if err != nil {
			return fmt.Errorf("date: %w", err)
		}
		t, err := time.ParseInLocation("2006-01-02", start, loc)
		if err != nil {
			return fmt.Errorf("date: %w", err)
		}
		day = t
	}

	startClock, err := parseClockHHMM(editShiftStart)
	if err != nil {
		return fmt.Errorf("punch in time: %w", err)
	}
	endClock, err := parseClockHHMM(editShiftEnd)
	if err != nil {
		return fmt.Errorf("punch out time: %w", err)
	}

	punchIn := time.Date(day.Year(), day.Month(), day.Day(), startClock.h, startClock.m, 0, 0, loc)
	punchOut := time.Date(day.Year(), day.Month(), day.Day(), endClock.h, endClock.m, 0, 0, loc)

	if !punchOut.After(punchIn) {
		punchOut = punchOut.AddDate(0, 0, 1)
	}
	if !punchOut.After(punchIn) {
		return fmt.Errorf("punch out must be after punch in")
	}

	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("credentials: %w", err)
	}

	if err := utils.EnsureObjectID(); err != nil {
		return fmt.Errorf("object id: %w", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	objectID, err := strconv.Atoi(cfg.PunchClockObjectID)
	if err != nil {
		return fmt.Errorf("invalid punch clock object id: %w", err)
	}

	fmt.Printf("PUT Mobile ShiftRequest  punchId=%s  tagId=%s\n", punchID, tagID)
	fmt.Printf("  %s → %s  (%s)\n", punchIn.Format("2006-01-02 15:04"), punchOut.Format("15:04"), loc.String())
	fmt.Printf("  note: %s\n", editShiftNote)

	resp, err := api.PutEditShift(creds, objectID, punchID, tagID, punchIn.Unix(), punchOut.Unix(), editShiftNote)
	if err != nil {
		return err
	}

	fmt.Printf("OK  code=%d  %s\n", resp.Code, resp.Message)
	return nil
}

type clockParts struct {
	h, m int
}

func resolveEditShiftTagID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if looksLikeEditShiftUUID(input) {
		return input, nil
	}
	return parseClientFile(input)
}

func looksLikeEditShiftUUID(s string) bool {
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

func parseClockHHMM(s string) (clockParts, error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return clockParts{}, fmt.Errorf("expected HH:MM")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return clockParts{}, fmt.Errorf("hours: %w", err)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return clockParts{}, fmt.Errorf("minutes: %w", err)
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return clockParts{}, fmt.Errorf("invalid time")
	}
	return clockParts{h: h, m: m}, nil
}

func init() {
	editshiftCmd.Flags().StringVarP(&editShiftStart, "start", "s", "", "Punch in time HH:MM (Asia/Kolkata)")
	editshiftCmd.Flags().StringVarP(&editShiftEnd, "end", "e", "", "Punch out time HH:MM")
	editshiftCmd.Flags().StringVarP(&editShiftNote, "note", "n", "", "Shift note")
	editshiftCmd.Flags().StringVarP(&editShiftClient, "client", "c", "", "Client tag UUID or clients/*.json")
	editshiftCmd.Flags().StringVarP(&editShiftDate, "date", "d", "", "Shift date: dd/mm, dd/mm/yy, today, yesterday (default: today)")

	_ = editshiftCmd.MarkFlagRequired("start")
	_ = editshiftCmd.MarkFlagRequired("end")
	_ = editshiftCmd.MarkFlagRequired("note")
	_ = editshiftCmd.MarkFlagRequired("client")

	rootCmd.AddCommand(editshiftCmd)
}
