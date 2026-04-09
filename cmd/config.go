package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"connectcli/internal/api"
	"connectcli/internal/credentials"
	"connectcli/internal/paths"
	sesspkg "connectcli/internal/session"
	"connectcli/internal/utils"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Interactive one-time setup: credentials, clients, and Jira cache",
	Long: `Walks through saving Connecteam session + CSRF, validates them, fetches clients,
then Jira credentials (validated) and up to 1000 tickets, then optional Slack webhook and user token
for channel notifications and profile status when using clockin/clockout.

Credentials and punch-clock config are stored under ~/.connectcli/.
Client and Jira JSON files are stored under ./clients/ and ./jira-tickets/ in the current working directory.`,
	Args: cobra.NoArgs,
	RunE: runConfigWizard,
}

func runConfigWizard(cmd *cobra.Command, args []string) error {
	if err := paths.EnsureDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	readLine := func(prompt string) (string, error) {
		fmt.Print(prompt)
		s, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(s), nil
	}

	fmt.Println("ConnectCLI setup — credentials under ~/.connectcli/; clients & Jira JSON in ./clients/ and ./jira-tickets/ (pwd).")
	fmt.Println()

	sessionCookie, err := readLine("Connecteam session cookie: ")
	if err != nil {
		return err
	}
	csrf, err := readLine("CSRF (_spirit) token: ")
	if err != nil {
		return err
	}
	if sessionCookie == "" || csrf == "" {
		return fmt.Errorf("session and CSRF token are required")
	}

	patch := &credentials.Credentials{Session: sessionCookie, CSRF: csrf}
	validator := sesspkg.NewValidator()
	fmt.Println("\nValidating session with Connecteam...")
	ok, err := validator.ValidateSession(patch)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("session is invalid or expired — check cookie and CSRF token")
	}
	fmt.Println("Session is valid.")

	if err := credentials.SaveCredentials(patch); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}
	fmt.Println("Saved ~/.connectcli/credentials")

	fmt.Println("\nResolving punch clock object ID...")
	if err := utils.EnsureObjectID(); err != nil {
		return fmt.Errorf("could not fetch punch clock object id: %w", err)
	}

	fmt.Println("\nFetching and saving clients...")
	if err := RunFetchClients(); err != nil {
		return err
	}

	jiraCreds, err := readLine("\nAtlassian Jira (email:api_token): ")
	if err != nil {
		return err
	}
	if jiraCreds == "" {
		return fmt.Errorf("Jira credentials are required to finish setup")
	}
	if !strings.Contains(jiraCreds, ":") {
		return fmt.Errorf("expected format email:api_token")
	}

	jiraClient := api.NewJiraClient(jiraCreds)
	fmt.Println("Validating Jira API access...")
	if _, err := jiraClient.FetchJiraTickets("", 1); err != nil {
		return fmt.Errorf("Jira validation failed: %w", err)
	}
	fmt.Println("Jira credentials OK.")

	if err := credentials.SaveCredentials(&credentials.Credentials{Jira: jiraCreds}); err != nil {
		return fmt.Errorf("failed to save Jira credentials: %w", err)
	}

	fmt.Println("\n--- Slack (optional; used by clockin / clockout / background reminders) ---")
	slackWebhook, err := readLine("Slack incoming webhook URL (Enter to skip): ")
	if err != nil {
		return err
	}
	slackUserTok, err := readLine("Slack user token (xoxp-..., for profile status; Enter to skip): ")
	if err != nil {
		return err
	}
	if slackWebhook != "" || slackUserTok != "" {
		slackPatch := &credentials.Credentials{}
		if slackWebhook != "" {
			slackPatch.SlackWebhook = slackWebhook
		}
		if slackUserTok != "" {
			slackPatch.SlackUserToken = slackUserTok
		}
		if err := credentials.SaveCredentials(slackPatch); err != nil {
			return fmt.Errorf("failed to save Slack credentials: %w", err)
		}
		if slackWebhook != "" {
			fmt.Println("Saved Slack webhook.")
		}
		if slackUserTok != "" {
			fmt.Println("Saved Slack user token (status updates).")
		}
	} else {
		fmt.Println("(Skipped Slack — you can add slack_webhook and slack_user_token to ~/.connectcli/credentials later.)")
	}

	fmt.Println("\nDownloading up to 1000 Jira tickets (may take a minute)...")
	if err := RunFetchJiraUpTo(1000); err != nil {
		return err
	}

	fmt.Println("\nSetup complete. You can use clockin, addshift, fetch, and search jira.")
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
}
