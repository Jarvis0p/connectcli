package cmd

import (
	"fmt"

	"connectcli/internal/api"
	"connectcli/internal/credentials"
	"connectcli/internal/storage"

	"github.com/spf13/cobra"
)

var (
	moreFlag bool
)

var fetchJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Fetch Jira tickets from TECH project",
	Long: `Fetch Jira tickets from the TECH project and save them locally.
By default, fetches 500 tickets. Use --more to fetch the next batch of 100 tickets.`,
	RunE: runFetchJira,
}

func runFetchJira(cmd *cobra.Command, args []string) error {
	// Load credentials
	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds.Jira == "" {
		return fmt.Errorf("Jira token not found in credentials file. Please add 'jira=krishna@securify.llc:<token>' to ~/.connectcli/credentials")
	}

	// Create Jira client
	client := api.NewJiraClient(creds.Jira)
	if client == nil {
		return fmt.Errorf("failed to create Jira client. Check your Jira token format")
	}

	// Create storage
	storage := storage.NewJiraStorage()

	// Load existing tickets
	if err := storage.LoadTickets(); err != nil {
		return fmt.Errorf("failed to load existing tickets: %w", err)
	}

	fmt.Printf("📊 Current tickets in storage: %d\n", storage.GetTotalTickets())

	var startAt int
	var maxResults int

	if moreFlag {
		// Fetch next 100 tickets
		startAt = storage.GetNextStartAt()
		maxResults = 100
		fmt.Printf("🔄 Fetching next 100 tickets starting from position %d...\n", startAt)

		// Fetch tickets
		response, err := client.FetchJiraTickets(startAt, maxResults)
		if err != nil {
			return fmt.Errorf("failed to fetch Jira tickets: %w", err)
		}

		fmt.Printf("✅ Fetched %d tickets from Jira API\n", len(response.Issues))
		fmt.Printf("📈 Total tickets in project: %d\n", response.Total)

		// Convert to tickets
		tickets := client.ConvertToTickets(response)

		// Add tickets to storage (prevents duplicates)
		added, duplicates := storage.AddTickets(tickets)

		fmt.Printf("📝 Added %d new tickets\n", added)
		if duplicates > 0 {
			fmt.Printf("⚠️  Skipped %d duplicate tickets\n", duplicates)
		}

		// Save to file
		if err := storage.SaveTickets(); err != nil {
			return fmt.Errorf("failed to save tickets: %w", err)
		}

		fmt.Printf("💾 Saved %d total tickets to jira-tickets/ directory\n", storage.GetTotalTickets())

		// Show some sample tickets
		if len(tickets) > 0 {
			fmt.Println("\n📋 Sample tickets:")
			for i, ticket := range tickets {
				if i >= 5 { // Show only first 5
					break
				}
				fmt.Printf("  %s: %s\n", ticket.Key, ticket.Summary)
			}
			if len(tickets) > 5 {
				fmt.Printf("  ... and %d more\n", len(tickets)-5)
			}
		}

		return nil
	} else {
		// Fetch first 500 tickets (in batches of 100)
		fmt.Printf("🔄 Fetching first 500 tickets (in batches of 100)...\n")

		totalFetched := 0
		totalAdded := 0
		totalDuplicates := 0

		for batch := 0; batch < 5; batch++ { // 5 batches of 100 = 500
			startAt = batch * 100
			maxResults = 100

			fmt.Printf("  📥 Fetching batch %d (tickets %d-%d)...\n", batch+1, startAt+1, startAt+100)

			response, err := client.FetchJiraTickets(startAt, maxResults)
			if err != nil {
				return fmt.Errorf("failed to fetch Jira tickets batch %d: %w", batch+1, err)
			}

			if len(response.Issues) == 0 {
				fmt.Printf("  ⚠️  No more tickets available (reached end)\n")
				break
			}

			totalFetched += len(response.Issues)
			fmt.Printf("  ✅ Fetched %d tickets in this batch\n", len(response.Issues))

			// Convert to tickets
			tickets := client.ConvertToTickets(response)

			// Add tickets to storage (prevents duplicates)
			added, duplicates := storage.AddTickets(tickets)
			totalAdded += added
			totalDuplicates += duplicates

			// Save after each batch
			if err := storage.SaveTickets(); err != nil {
				return fmt.Errorf("failed to save tickets: %w", err)
			}
		}

		fmt.Printf("\n📊 Summary:\n")
		fmt.Printf("✅ Total fetched from API: %d tickets\n", totalFetched)
		fmt.Printf("📝 Total added to storage: %d tickets\n", totalAdded)
		if totalDuplicates > 0 {
			fmt.Printf("⚠️  Total duplicates skipped: %d tickets\n", totalDuplicates)
		}
		fmt.Printf("💾 Total tickets in storage: %d\n", storage.GetTotalTickets())

		// Show some sample tickets from the last batch
		allTickets := storage.GetTickets()
		if len(allTickets) > 0 {
			fmt.Println("\n📋 Sample tickets:")
			for i, ticket := range allTickets {
				if i >= 5 { // Show only first 5
					break
				}
				fmt.Printf("  %s: %s\n", ticket.Key, ticket.Summary)
			}
			if len(allTickets) > 5 {
				fmt.Printf("  ... and %d more\n", len(allTickets)-5)
			}
		}

		return nil
	}
}

func init() {
	fetchJiraCmd.Flags().BoolVarP(&moreFlag, "more", "m", false, "Fetch next 100 tickets (instead of first 500)")
	fetchCmd.AddCommand(fetchJiraCmd)
}
