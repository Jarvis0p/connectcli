package cmd

import (
	"fmt"
	"strings"

	"connectcli/internal/api"
	"connectcli/internal/credentials"
	"connectcli/internal/paths"
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
		if strings.TrimSpace(creds.Jira) == "" {
			return fmt.Errorf("jira credentials not set (email:api_token in ~/.connectcli/credentials)")
		}

		client := api.NewJiraClient(creds.Jira)
		st := storage.NewJiraStorage()
		_ = st.LoadTickets()

		fmt.Printf("📊 Current tickets in storage: %d\n", st.GetTotalTickets())

		if moreFlag {
			fmt.Printf("🔄 Fetching next 100 tickets starting from position %d...\n", st.GetTotalTickets())
			resp, err := client.FetchJiraTickets(st.GetNextPageToken(), 100)
			if err != nil {
				return fmt.Errorf("failed to fetch Jira tickets: %w", err)
			}
			tickets := client.ConvertToTickets(resp)
			added, dups := st.AddTickets(tickets)
			st.SetNextPageToken(resp.NextPageToken)
			if err := st.SaveTickets(); err != nil {
				return fmt.Errorf("failed to save tickets: %w", err)
			}
			fmt.Printf("✅ Fetched %d tickets from Jira API\n", len(tickets))
			fmt.Printf("📝 Added %d new tickets\n", added)
			if dups > 0 {
				fmt.Printf("⚠️  Skipped %d duplicate tickets\n", dups)
			}
			jiraDir, _ := paths.JiraTicketsDir()
			fmt.Printf("💾 Saved %d total tickets to %s\n", st.GetTotalTickets(), jiraDir)
			return nil
		}

		fmt.Printf("🔄 Fetching first 500 tickets (in batches of 100)...\n")
		return runFetchJiraBatches(client, st, 500, true)
	},
}

// RunFetchJiraUpTo fetches up to maxTickets issues (batches of 100) into ~/.connectcli/jira-tickets.
func RunFetchJiraUpTo(maxTickets int) error {
	creds, err := credentials.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}
	if strings.TrimSpace(creds.Jira) == "" {
		return fmt.Errorf("jira credentials not set")
	}
	client := api.NewJiraClient(creds.Jira)
	st := storage.NewJiraStorage()
	_ = st.LoadTickets()
	fmt.Printf("🔄 Fetching up to %d Jira tickets (batches of 100)...\n", maxTickets)
	return runFetchJiraBatches(client, st, maxTickets, true)
}

func runFetchJiraBatches(client *api.JiraClient, st *storage.JiraStorage, maxTickets int, printBatches bool) error {
	if maxTickets <= 0 {
		return fmt.Errorf("maxTickets must be positive")
	}
	batchCount := (maxTickets + 99) / 100
	next := ""
	totalAdded := 0
	for batch := 0; batch < batchCount; batch++ {
		if printBatches {
			fmt.Printf("  📥 Fetching batch %d...\n", batch+1)
		}
		resp, err := client.FetchJiraTickets(next, 100)
		if err != nil {
			return fmt.Errorf("failed to fetch Jira tickets batch %d: %w", batch+1, err)
		}
		tickets := client.ConvertToTickets(resp)
		if len(tickets) == 0 {
			if printBatches {
				fmt.Println("  ⚠️  No more tickets available (reached end)")
			}
			break
		}
		added, _ := st.AddTickets(tickets)
		totalAdded += added
		next = resp.NextPageToken
		if err := st.SaveTickets(); err != nil {
			return fmt.Errorf("failed to save tickets: %w", err)
		}
		if next == "" {
			if printBatches {
				fmt.Println("  ⚠️  No more tickets available (reached end)")
			}
			break
		}
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("📝 Total added to storage this run: %d tickets\n", totalAdded)
	jiraDir, _ := paths.JiraTicketsDir()
	fmt.Printf("💾 Total tickets in storage: %d (%s)\n", st.GetTotalTickets(), jiraDir)
	return nil
}

func init() {
	fetchJiraCmd.Flags().BoolVarP(&moreFlag, "more", "m", false, "Fetch next 100 tickets (instead of first 500)")
	fetchCmd.AddCommand(fetchJiraCmd)
}
