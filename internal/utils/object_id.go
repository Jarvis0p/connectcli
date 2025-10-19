package utils

import (
	"fmt"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
)

// EnsureObjectID ensures that the punch clock object ID is available in config
// If it's not available, it fetches it automatically and saves it
func EnsureObjectID() error {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If object ID is already available, return early
	if cfg.PunchClockObjectID != "" {
		return nil
	}

	// Load credentials
	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	fmt.Println("🔍 Punch clock object ID not found in config. Fetching automatically...")

	// Create content structure client and fetch data
	client := api.NewContentStructureClient()
	response, err := client.FetchContentStructure(creds)
	if err != nil {
		return fmt.Errorf("failed to fetch content structure: %w", err)
	}

	// Extract punch clock object ID
	objectID, err := client.ExtractPunchClockObjectID(response)
	if err != nil {
		return fmt.Errorf("failed to extract punch clock object ID: %w", err)
	}

	// Update config with the object ID
	cfg.PunchClockObjectID = fmt.Sprintf("%d", objectID)

	// Save updated config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Punch clock object ID automatically fetched and saved: %d\n", objectID)

	return nil
}
