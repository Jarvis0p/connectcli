package cmd

import (
	"fmt"
	"strconv"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var deleteShiftCmd = &cobra.Command{
	Use:   "shift <punch-id>",
	Short: "Delete a shift by punch ID",
	Long: `Delete a shift from Connecteam using the punch ID.

This command will cancel/delete the specified shift from your timesheet.

Examples:
  connectcli delete shift 68f54b4af693908de452bac7
  connectcli delete shift 68f53ec88f7d24ae49540767`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteShift,
}

func runDeleteShift(cmd *cobra.Command, args []string) error {
	punchID := args[0]

	// Validate punch ID format (basic validation)
	if len(punchID) < 10 {
		return fmt.Errorf("invalid punch ID format: %s", punchID)
	}

	// Load credentials
	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// Check if authorization header is available
	if creds.Authorization == "" {
		return fmt.Errorf("authorization header not found in credentials file. Please add 'authorization=your_auth_token' to ~/.connectcli/credentials")
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

	// Display request details
	fmt.Println("🗑️  Delete Shift Request Details:")
	fmt.Printf("Punch ID: %s\n", punchID)
	fmt.Printf("Object ID: %d\n", objectID)
	fmt.Println()

	// Create shift deletion client and send request
	client := api.NewShiftDeletionClient()
	fmt.Println("🔄 Sending shift deletion request to Connecteam API...")

	response, err := client.DeleteShift(creds, objectID, punchID)
	if err != nil {
		return fmt.Errorf("failed to delete shift: %w", err)
	}

	fmt.Printf("✅ Shift deleted successfully (Request ID: %s)\n", response.RequestID)
	fmt.Printf("Server version: %s\n", response.ServerVersion)
	fmt.Printf("Response code: %d - %s\n", response.Code, response.Message)

	return nil
}

func init() {
	deleteCmd.AddCommand(deleteShiftCmd)
}

