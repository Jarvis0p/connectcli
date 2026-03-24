package credentials

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Credentials struct {
	Session        string
	CSRF           string
	Jira           string
	Authorization  string
	SlackWebhook   string
	// SlackStatusToken: set via slack_user_token= or slack_bot_token= in credentials (Bearer for users.profile.set).
	SlackUserToken string
}

// LoadCredentials reads the credentials from .connectcli/credentials file
func LoadCredentials() (*Credentials, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	credentialsPath := filepath.Join(homeDir, ".connectcli", "credentials")

	file, err := os.Open(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open credentials file: %w", err)
	}
	defer file.Close()

	creds := &Credentials{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "session":
			creds.Session = value
		case "csrf":
			creds.CSRF = value
		case "jira":
			creds.Jira = value
		case "authorization":
			creds.Authorization = value
		case "slack_webhook":
			creds.SlackWebhook = value
		case "slack_user_token", "slack_bot_token":
			creds.SlackUserToken = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading credentials file: %w", err)
	}

	if creds.Session == "" {
		return nil, fmt.Errorf("session cookie not found in credentials file")
	}

	if creds.CSRF == "" {
		return nil, fmt.Errorf("CSRF token not found in credentials file")
	}

	return creds, nil
}
