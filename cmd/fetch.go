package cmd

import (
	"encoding/json"
	"fmt"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from Connecteam",
	Long:  `Fetch various types of data from the Connecteam application.`,
}

var (
	verboseFlag bool
)

var fetchTimesheetCmd = &cobra.Command{
	Use:   "timesheet [date-range]",
	Short: "Fetch timesheet data for a specific date or date range",
	Long: `Fetch timesheet data from Connecteam.
	
Date format: dd/mm or dd/mm/yy
Examples:
  connectcli fetch timesheet 01/07             # Single date (current year)
  connectcli fetch timesheet 01/07/25          # Single date with year
  connectcli fetch timesheet 29/06-01/07       # Date range (current year)
  connectcli fetch timesheet 29/06/25-01/07/25 # Date range with year
  connectcli fetch timesheet -v 01/07          # With verbose output (full notes)`,
	Args: cobra.ExactArgs(1),
	RunE: runFetchTimesheet,
}

func runFetchTimesheet(cmd *cobra.Command, args []string) error {
	dateRange := args[0]

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

	// Parse date range
	startDate, endDate, err := utils.ParseDateRange(dateRange)
	if err != nil {
		return fmt.Errorf("failed to parse date range: %w", err)
	}

	// Display parsed dates
	startDisplay, _ := utils.FormatDateForDisplay(startDate)
	endDisplay, _ := utils.FormatDateForDisplay(endDate)

	if startDate == endDate {
		fmt.Printf("Fetching timesheet for: %s\n", startDisplay)
	} else {
		fmt.Printf("Fetching timesheet from: %s to: %s\n", startDisplay, endDisplay)
	}
	fmt.Println()

	// Create timesheet client and fetch data
	client := api.NewTimesheetClient()

	response, err := client.FetchTimesheet(creds, cfg.PunchClockObjectID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch timesheet: %w", err)
	}

	// Parse and display timesheet data in table format
	if len(response.Data.RawData) > 0 {
		fmt.Println("\n📊 Timesheet Summary:")

		// Parse the timesheet data
		shifts, err := utils.ParseTimesheetData(response.Data.RawData)
		if err != nil {
			fmt.Printf("Error parsing timesheet data: %v\n", err)
			// Fallback to raw JSON display
			prettyJSON, err := json.MarshalIndent(response.Data.RawData, "", "  ")
			if err != nil {
				fmt.Printf("Raw data: %+v\n", response.Data.RawData)
			} else {
				fmt.Println(string(prettyJSON))
			}
		} else {
			// Display in table format
			table := utils.FormatTimesheetTable(shifts, verboseFlag)
			fmt.Println(table)

			// Show summary statistics
			if len(shifts) > 0 {
				var totalHours float64
				for _, shift := range shifts {
					totalHours += shift.TotalHours
				}
				fmt.Printf("\n📈 Summary: %d shifts, Total: %.2f hours\n", len(shifts), totalHours)
			}
		}
	}

	return nil
}

func init() {
	fetchTimesheetCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show full employee notes (not truncated)")
	fetchCmd.AddCommand(fetchTimesheetCmd)
}
