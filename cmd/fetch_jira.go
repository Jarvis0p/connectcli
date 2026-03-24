package cmd

import (
	"fmt"

	"connectcli/internal/api"
	"connectcli/internal/credentials"
	"connectcli/internal/storage"

	"github.com/spf13/cobra"
)

var moreFlag bool

var fetchJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Fetch Jira tickets from TECH project",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := credentials.LoadCredentials()
		if err != nil {
			return fmt.Errorf("failed to load credentials: %w", err)
		}

		client := api.NewJiraClient(creds.Jira)
		storage := storage.NewJiraStorage()
		_ = storage.LoadTickets()

		fmt.Printf("📊 Current tickets in storage: %d\n", storage.GetTotalTickets())

		if moreFlag {
			fmt.Printf("🔄 Fetching next 100 tickets starting from position %d...\n", storage.GetTotalTickets())
			resp, err := client.FetchJiraTickets(storage.GetNextPageToken(), 100)
			if err != nil {
				return fmt.Errorf("failed to fetch Jira tickets: %w", err)
			}
			tickets := client.ConvertToTickets(resp)
			added, dups := storage.AddTickets(tickets)
			storage.SetNextPageToken(resp.NextPageToken)
			if err := storage.SaveTickets(); err != nil {
				return fmt.Errorf("failed to save tickets: %w", err)
			}
			fmt.Printf("✅ Fetched %d tickets from Jira API\n", len(tickets))
			fmt.Printf("📝 Added %d new tickets\n", added)
			if dups > 0 { fmt.Printf("⚠️  Skipped %d duplicate tickets\n", dups) }
			fmt.Printf("💾 Saved %d total tickets to jira-tickets/ directory\n", storage.GetTotalTickets())
			return nil
		}

		fmt.Printf("🔄 Fetching first 500 tickets (in batches of 100)...\n")
		next := ""
		totalAdded := 0
		for batch := 0; batch < 5; batch++ {
			fmt.Printf("  📥 Fetching batch %d...\n", batch+1)
			resp, err := client.FetchJiraTickets(next, 100)
			if err != nil { return fmt.Errorf("failed to fetch Jira tickets batch %d: %w", batch+1, err) }
			tickets := client.ConvertToTickets(resp)
			if len(tickets) == 0 { fmt.Println("  ⚠️  No more tickets available (reached end)"); break }
			added, _ := storage.AddTickets(tickets)
			totalAdded += added
			next = resp.NextPageToken
			if next == "" { fmt.Println("  ⚠️  No more tickets available (reached end)"); break }
			if err := storage.SaveTickets(); err != nil { return fmt.Errorf("failed to save tickets: %w", err) }
		}

		fmt.Printf("\n📊 Summary:\n")
		fmt.Printf("📝 Total added to storage: %d tickets\n", totalAdded)
		fmt.Printf("💾 Total tickets in storage: %d\n", storage.GetTotalTickets())
		return nil
	},
}

func init() {
	fetchJiraCmd.Flags().BoolVarP(&moreFlag, "more", "m", false, "Fetch next 100 tickets (instead of first 500)")
	fetchCmd.AddCommand(fetchJiraCmd)
}





