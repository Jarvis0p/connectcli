package cmd

import (
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("\nTo fix this issue:")
		fmt.Println("1. Create a ~/.connectcli/credentials file")
		fmt.Println("2. Add your session cookie and CSRF token:")
		fmt.Println("   session=your_session_cookie_here")
		fmt.Println("   csrf=your_csrf_token_here")
		os.Exit(1)
	}

	fmt.Println("Loading credentials...")
	fmt.Printf("Session cookie: %s...\n", previewSecret(creds.Session))
	fmt.Printf("CSRF token: %s...\n", previewSecret(creds.CSRF))
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

func previewSecret(s string) string {
	if len(s) <= 24 {
		return s
	}
	return s[:20] + "..."
}
