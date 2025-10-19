package cmd

import (
	"fmt"
	"strconv"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/storage"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var fetchClientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Fetch client data from Connecteam",
	Long: `Fetch client data from the Connecteam API and save client names along with their IDs.
This command will save each client as an individual file in the clients/ directory.`,
	RunE: runFetchClients,
}

func runFetchClients(cmd *cobra.Command, args []string) error {
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

	// Convert object ID to int
	objectID, err := strconv.Atoi(cfg.PunchClockObjectID)
	if err != nil {
		return fmt.Errorf("invalid punch clock object ID: %w", err)
	}

	fmt.Printf("📊 Punch clock object ID: %d\n", objectID)
	fmt.Println()

	// Create clients client and fetch data
	client := api.NewClientsClient()
	fmt.Println("🔄 Fetching client data from Connecteam API...")

	response, err := client.FetchClients(creds, objectID)
	if err != nil {
		return fmt.Errorf("failed to fetch clients: %w", err)
	}

	fmt.Printf("✅ Client data fetched successfully (Request ID: %s)\n", response.RequestID)
	fmt.Printf("Server version: %s\n", response.ServerVersion)
	fmt.Printf("Response code: %d - %s\n", response.Code, response.Message)

	// Extract clients from response
	fmt.Println("\n🔍 Extracting client information...")
	clients, err := client.ExtractClients(response)
	if err != nil {
		return fmt.Errorf("failed to extract clients: %w", err)
	}

	if len(clients) == 0 {
		fmt.Println("⚠️  No clients found in the response")
		fmt.Println("📋 Raw response data structure:")
		// Print the raw data structure for debugging
		for key, value := range response.Data.RawData {
			fmt.Printf("  %s: %T\n", key, value)
		}
		return nil
	}

	fmt.Printf("✅ Found %d clients\n", len(clients))

	// Create storage and load existing clients
	storage := storage.NewClientsStorage()
	if err := storage.LoadClients(); err != nil {
		return fmt.Errorf("failed to load existing clients: %w", err)
	}

	fmt.Printf("📊 Current clients in storage: %d\n", storage.GetTotalClients())

	// Add new clients to storage (prevents duplicates)
	added, duplicates := storage.AddClients(clients)

	fmt.Printf("📝 Added %d new clients\n", added)
	if duplicates > 0 {
		fmt.Printf("⚠️  Skipped %d duplicate clients\n", duplicates)
	}

	// Save to file
	if err := storage.SaveClients(); err != nil {
		return fmt.Errorf("failed to save clients: %w", err)
	}

	fmt.Printf("💾 Saved %d total clients to clients/ directory\n", storage.GetTotalClients())

	// Show all clients
	fmt.Println("\n📋 Client List:")
	fmt.Println("ID                                     Name")
	fmt.Println("--                                     ----")
	for _, client := range clients {
		fmt.Printf("%-38s %s\n", client.ID, client.Name)
	}

	return nil
}

func init() {
	fetchCmd.AddCommand(fetchClientsCmd)
}
