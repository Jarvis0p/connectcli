package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var addshiftCmd = &cobra.Command{
	Use:   "addshift",
	Short: "Add a shift request to Connecteam",
	Long: `Add a shift request to Connecteam with client, date, duration, and optional ticket information.

The shift will be automatically scheduled after existing shifts for the day, starting at 9:00 AM if no shifts exist.

Examples:
  connectcli addshift --client "clients/Keyo.json" --date "01/07" --duration "02:30" --note "Working on project"
  connectcli addshift --client "clients/Silentpush.json" --date "15/07" --duration "08:00" --ticket "jira-tickets/TECH-2200 - BiMo Appetize Pentest.json" --note "Performed pentest"
  connectcli addshift --client "clients/Keyo.json" --date "today" --duration "04:00" --note "Today's work"
  connectcli addshift --client "clients/Keyo.json" --date "yesterday" --duration "06:00" --note "Yesterday's work"
`,
	RunE: runAddShift,
}

var (
	clientFlag   string
	dateFlag     string
	durationFlag string
	ticketFlag   string
	noteFlag     string
)

func runAddShift(cmd *cobra.Command, args []string) error {
	// Validate required flags
	if clientFlag == "" {
		return fmt.Errorf("client flag is required")
	}
	if dateFlag == "" {
		return fmt.Errorf("date flag is required")
	}
	if durationFlag == "" {
		return fmt.Errorf("duration flag is required")
	}
	if noteFlag == "" {
		return fmt.Errorf("note flag is required")
	}

	// Parse duration
	duration, err := parseDuration(durationFlag)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	// Load credentials
	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// Ensure object ID is available
	if err := utils.EnsureObjectID(); err != nil {
		return fmt.Errorf("failed to ensure object ID: %w", err)
	}

	// Load config to get punch clock object ID
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	objectID, err := strconv.Atoi(cfg.PunchClockObjectID)
	if err != nil {
		return fmt.Errorf("invalid punch clock object ID: %w", err)
	}

	// Parse client file
	clientID, err := parseClientFile(clientFlag)
	if err != nil {
		return fmt.Errorf("failed to parse client file: %w", err)
	}

	// Parse date - handle "today" and "yesterday" keywords
	var fullDate time.Time
	dateLower := strings.ToLower(dateFlag)
	
	if dateLower == "today" {
		// Use today's date
		loc, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			return fmt.Errorf("failed to load timezone: %w", err)
		}
		fullDate = time.Now().In(loc)
	} else if dateLower == "yesterday" {
		// Use yesterday's date
		loc, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			return fmt.Errorf("failed to load timezone: %w", err)
		}
		fullDate = time.Now().In(loc).AddDate(0, 0, -1)
	} else {
		// Parse date using utils package
		startDate, _, err := utils.ParseDateRange(dateFlag)
		if err != nil {
			return fmt.Errorf("failed to parse date: %w", err)
		}
		
		// Convert to time.Time
		fullDate, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return fmt.Errorf("failed to parse date: %w", err)
		}
	}

	// Calculate start time based on existing shifts
	startTime, err := calculateStartTime(creds, objectID, fullDate)
	if err != nil {
		return fmt.Errorf("failed to calculate start time: %w", err)
	}

	// Calculate end time
	endTime := startTime.Add(duration)



	// Check if shift extends past midnight
	dayEnd := time.Date(fullDate.Year(), fullDate.Month(), fullDate.Day(), 23, 59, 59, 0, startTime.Location())
	if endTime.After(dayEnd) {
		return fmt.Errorf("shift would extend past midnight (end time: %s). Please reduce duration or choose a different date", endTime.Format("15:04"))
	}

	// Convert times to timestamps
	punchInTime := startTime.Unix()
	punchOutTime := endTime.Unix()

	// Build note with optional ticket information
	finalNote := noteFlag
	if ticketFlag != "" {
		ticketInfo, err := parseTicketFile(ticketFlag)
		if err != nil {
			return fmt.Errorf("failed to parse ticket file: %w", err)
		}
		finalNote = fmt.Sprintf("%s - %s - %s", ticketInfo.Key, ticketInfo.Summary, noteFlag)
	}

	// Create shift request
	request := &api.ShiftRequest{
		TagHierarchy:     []string{clientID},
		PunchInTime:      punchInTime,
		PunchOutTime:     punchOutTime,
		Note:             finalNote,
		ShiftAttachments: []string{},
		Timezone:         "Asia/Kolkata",
	}

	// Display request details
	fmt.Println("📋 Shift Request Details:")
	fmt.Printf("Client ID: %s\n", clientID)
	fmt.Printf("Date: %s\n", fullDate.Format("02/01/2006"))
	fmt.Printf("Start Time: %s\n", startTime.Format("15:04"))
	fmt.Printf("End Time: %s\n", endTime.Format("15:04"))
	fmt.Printf("Duration: %s\n", durationFlag)
	fmt.Printf("Note: %s\n", finalNote)
	fmt.Printf("Object ID: %d\n", objectID)
	fmt.Println()

	// Create shift request client and send request
	client := api.NewShiftRequestClient()
	fmt.Println("🔄 Sending shift request to Connecteam API...")

	response, err := client.AddShiftRequest(creds, objectID, request)
	if err != nil {
		return fmt.Errorf("failed to add shift request: %w", err)
	}

	fmt.Printf("✅ Shift request sent successfully (Request ID: %s)\n", response.RequestID)
	fmt.Printf("Server version: %s\n", response.ServerVersion)
	fmt.Printf("Response code: %d - %s\n", response.Code, response.Message)

	return nil
}

// calculateStartTime determines the start time for a new shift based on existing shifts
func calculateStartTime(creds *credentials.Credentials, objectID int, date time.Time) (time.Time, error) {
	// Load Asia/Kolkata timezone
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load timezone: %w", err)
	}

	// Default start time is 9:00 AM
	defaultStart := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, loc)

	// Fetch existing shifts for the date
	shifts, err := fetchExistingShifts(creds, objectID, date)
	if err != nil {
		// If we can't fetch shifts, use default start time
		fmt.Printf("⚠️  Warning: Could not fetch existing shifts, using default start time (9:00 AM): %v\n", err)
		return defaultStart, nil
	}

	if len(shifts) == 0 {
		// No existing shifts, use default start time
		return defaultStart, nil
	}

	// Find the latest end time
	var latestEnd time.Time
	for _, shift := range shifts {
		shiftEnd := time.Unix(shift.EndTime, 0).In(loc)
		if shiftEnd.After(latestEnd) {
			latestEnd = shiftEnd
		}
	}

	// Start time is 1 minute after the latest shift ends
	startTime := latestEnd.Add(time.Minute)
	
	// Ensure we don't start before 9:00 AM
	if startTime.Before(defaultStart) {
		startTime = defaultStart
	}

	// Check if starting at this time would extend past midnight
	// We'll assume a reasonable maximum shift duration of 8 hours for this check
	maxShiftDuration := 8 * time.Hour
	estimatedEndTime := startTime.Add(maxShiftDuration)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, loc)
	
	if estimatedEndTime.After(dayEnd) {
		// If the estimated end time would be past midnight, return an error
		return time.Time{}, fmt.Errorf("cannot schedule new shift: existing shifts extend too late in the day (latest ends at %s). The day is already fully booked", latestEnd.Format("15:04"))
	}

	return startTime, nil
}

// fetchExistingShifts fetches existing shifts for a given date
func fetchExistingShifts(creds *credentials.Credentials, objectID int, date time.Time) ([]Shift, error) {
	// Format date for API
	dateStr := date.Format("2006-01-02")
	
	// Create timesheet client
	client := api.NewTimesheetClient()
	
	// Fetch timesheet data for the specific date
	response, err := client.FetchTimesheet(creds, fmt.Sprintf("%d", objectID), dateStr, dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch timesheet: %w", err)
	}

	// Parse shifts from response
	// Note: This is a simplified implementation. You may need to adjust based on actual API response structure
	shifts, err := parseShiftsFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shifts from response: %w", err)
	}

	return shifts, nil
}

// Shift represents a shift entry
type Shift struct {
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

// parseShiftsFromResponse extracts shift data from the timesheet response
func parseShiftsFromResponse(response *api.TimesheetResponse) ([]Shift, error) {
	var shifts []Shift

	// Parse the timesheet data using the existing utility
	timesheetShifts, err := utils.ParseTimesheetData(response.Data.RawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timesheet data: %w", err)
	}

	// Convert to our Shift format
	for _, ts := range timesheetShifts {
		shifts = append(shifts, Shift{
			StartTime: ts.StartTime.Unix(),
			EndTime:   ts.EndTime.Unix(),
		})
	}

	return shifts, nil
}

// parseDuration parses duration in HH:MM format
func parseDuration(durationStr string) (time.Duration, error) {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format. Expected HH:MM (e.g., '02:30' for 2 hours 30 minutes)")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours in duration: %s", parts[0])
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes in duration: %s", parts[1])
	}

	if hours < 0 || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("invalid duration values. Hours must be >= 0, minutes must be 0-59")
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute, nil
}

// parseClientFile reads the client ID from a client JSON file
func parseClientFile(clientPath string) (string, error) {
	// Clean the path and ensure it points to the clients directory
	clientPath = strings.TrimPrefix(clientPath, "clients/")
	clientPath = strings.TrimPrefix(clientPath, ".\\clients\\")
	clientPath = strings.TrimPrefix(clientPath, "./clients/")
	clientPath = filepath.Join("clients", clientPath)

	// Read the client file
	fileContent, err := os.ReadFile(clientPath)
	if err != nil {
		return "", fmt.Errorf("failed to read client file %s: %w", clientPath, err)
	}

	var client api.Client
	if err := json.Unmarshal(fileContent, &client); err != nil {
		return "", fmt.Errorf("failed to parse client file %s: %w", clientPath, err)
	}

	if client.ID == "" {
		return "", fmt.Errorf("client ID not found in file %s", clientPath)
	}

	return client.ID, nil
}

// parseTicketFile reads the ticket information from a ticket JSON file
func parseTicketFile(ticketPath string) (*api.JiraTicket, error) {
	// Clean the path and ensure it points to the jira-tickets directory
	ticketPath = strings.TrimPrefix(ticketPath, "jira-tickets/")
	ticketPath = strings.TrimPrefix(ticketPath, ".\\jira-tickets\\")
	ticketPath = strings.TrimPrefix(ticketPath, "./jira-tickets/")
	ticketPath = filepath.Join("jira-tickets", ticketPath)

	// Read the ticket file
	fileContent, err := os.ReadFile(ticketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ticket file %s: %w", ticketPath, err)
	}

	var ticket api.JiraTicket
	if err := json.Unmarshal(fileContent, &ticket); err != nil {
		return nil, fmt.Errorf("failed to parse ticket file %s: %w", ticketPath, err)
	}

	if ticket.Key == "" || ticket.Summary == "" {
		return nil, fmt.Errorf("ticket key or summary not found in file %s", ticketPath)
	}

	return &ticket, nil
}



func init() {
	addshiftCmd.Flags().StringVarP(&clientFlag, "client", "c", "", "Client file path (e.g., 'clients/Keyo.json')")
	addshiftCmd.Flags().StringVarP(&dateFlag, "date", "d", "", "Date in dd/mm format (e.g., '01/07'), 'today' for current date, or 'yesterday' for previous date")
	addshiftCmd.Flags().StringVarP(&durationFlag, "duration", "r", "", "Duration of the shift (e.g., '02:30' for 2 hours 30 minutes)")
	addshiftCmd.Flags().StringVarP(&ticketFlag, "ticket", "t", "", "Ticket file path (optional, e.g., 'jira-tickets/TECH-2200 - BiMo Appetize Pentest.json')")
	addshiftCmd.Flags().StringVarP(&noteFlag, "note", "n", "", "Note/task description")

	// Mark required flags
	addshiftCmd.MarkFlagRequired("client")
	addshiftCmd.MarkFlagRequired("date")
	addshiftCmd.MarkFlagRequired("duration")
	addshiftCmd.MarkFlagRequired("note")
}
