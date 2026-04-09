package cmd

import (
	"fmt"
	"strings"

	"connectcli/internal/paths"
	"connectcli/internal/search"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search local cached data",
	Long:  `Search files stored by connectcli (e.g. Jira tickets under jira-tickets/).`,
}

var searchJiraCmd = &cobra.Command{
	Use:   "jira [query]...",
	Short: "Search local Jira ticket files (case-insensitive)",
	Long: `Search the jira-tickets/ directory for tickets whose key, summary, or filename contains the query.

Examples:
  connectcli search jira pentest
  connectcli search jira TECH-2200
  connectcli search jira "BiMo Appetize"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		hits, err := search.JiraLocal(query)
		if err != nil {
			return err
		}
		jiraDir, _ := paths.JiraTicketsDir()
		if len(hits) == 0 {
			fmt.Printf("No tickets matching %q in %s\n", query, jiraDir)
			return nil
		}
		fmt.Printf("Matches in %s (%d):\n\n", jiraDir, len(hits))
		fmt.Printf("%-14s %s\n", "Key", "Summary")
		fmt.Printf("%-14s %s\n", "---", "-------")
		for _, h := range hits {
			fmt.Printf("%-14s %s\n", h.Key, h.Summary)
		}
		fmt.Println()
		for _, h := range hits {
			fmt.Println(h.File)
		}
		return nil
	},
}

func init() {
	searchCmd.AddCommand(searchJiraCmd)
	rootCmd.AddCommand(searchCmd)
}
