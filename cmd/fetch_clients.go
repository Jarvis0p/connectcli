package cmd

import (
	"fmt"
	"strconv"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/paths"
	"connectcli/internal/storage"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var fetchClientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Fetch client data from Connecteam",
	Long: `Fetch client data from the Connecteam API and save client names along with their IDs.
This command will save each client as an individual file in the clients/ directory (relative to the current working directory).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunFetchClients()
	},
}

// RunFetchClients fetches Connecteam clients and saves them to ~/.connectcli/clients.
func RunFetchClients() error {
	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
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

	fmt.Printf("📊 Punch clock object ID: %d\n", objectID)
	fmt.Println()

	client := api.NewClientsClient()
	fmt.Println("🔄 Fetching client data from Connecteam API...")

	response, err := client.FetchClients(creds, objectID)
	if err != nil {
		return fmt.Errorf("failed to fetch clients: %w", err)
	}

	fmt.Printf("✅ Client data fetched successfully (Request ID: %s)\n", response.RequestID)
	fmt.Printf("Server version: %s\n", response.ServerVersion)
	fmt.Printf("Response code: %d - %s\n", response.Code, response.Message)

	fmt.Println("\n🔍 Extracting client information...")
	clients, err := client.ExtractClients(response)
	if err != nil {
		return fmt.Errorf("failed to extract clients: %w", err)
	}

	if len(clients) == 0 {
		fmt.Println("⚠️  No clients found in the response")
		fmt.Println("📋 Raw response data structure:")
		for key, value := range response.Data.RawData {
			fmt.Printf("  %s: %T\n", key, value)
		}
		return nil
	}

	fmt.Printf("✅ Found %d clients\n", len(clients))

	st := storage.NewClientsStorage()
	if err := st.LoadClients(); err != nil {
		return fmt.Errorf("failed to load existing clients: %w", err)
	}

	fmt.Printf("📊 Current clients in storage: %d\n", st.GetTotalClients())

	added, duplicates := st.AddClients(clients)

	fmt.Printf("📝 Added %d new clients\n", added)
	if duplicates > 0 {
		fmt.Printf("⚠️  Skipped %d duplicate clients\n", duplicates)
	}

	if err := st.SaveClients(); err != nil {
		return fmt.Errorf("failed to save clients: %w", err)
	}

	clientsDir, _ := paths.ClientsDir()
	fmt.Printf("💾 Saved %d total clients to %s\n", st.GetTotalClients(), clientsDir)

	fmt.Println("\n📋 Client List:")
	fmt.Println("ID                                     Name")
	fmt.Println("--                                     ----")
	for _, cl := range clients {
		fmt.Printf("%-38s %s\n", cl.ID, cl.Name)
	}

	return nil
}

func init() {
	fetchCmd.AddCommand(fetchClientsCmd)
}
