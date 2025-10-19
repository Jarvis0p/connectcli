package cmd

import (
	"fmt"

	"connectcli/internal/credentials"
	"connectcli/internal/session"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var validateSessionCmd = &cobra.Command{
	Use:   "validate-session",
	Short: "Validate the saved session credentials",
	Long: `Validate the session cookie and CSRF token saved in .connectcli/credentials file.
This command will make a request to the Connecteam API to check if the session is still valid.`,
	RunE: runValidateSession,
}

func runValidateSession(cmd *cobra.Command, args []string) error {
	// Load credentials
	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	fmt.Println("Loading credentials...")
	fmt.Printf("Session cookie: %s...\n", creds.Session[:20]+"...")
	fmt.Printf("CSRF token: %s...\n", creds.CSRF[:20]+"...")
	fmt.Println()

	// Create validator and validate session
	validator := session.NewValidator()
	fmt.Println("Validating session with Connecteam API...")

	isValid, err := validator.ValidateSession(creds)
	if err != nil {
		return fmt.Errorf("failed to validate session: %w", err)
	}

	if isValid {
		fmt.Println("✅ Session is valid!")

		// Automatically fetch and store object ID if not available
		if err := utils.EnsureObjectID(); err != nil {
			fmt.Printf("⚠️  Warning: Could not fetch object ID: %v\n", err)
		}
	} else {
		fmt.Println("❌ Session is invalid or expired.")
	}

	return nil
}
