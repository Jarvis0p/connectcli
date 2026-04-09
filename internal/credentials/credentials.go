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

// LoadCredentialsOptional reads ~/.connectcli/credentials without requiring session/csrf (for merging).
func LoadCredentialsOptional() (*Credentials, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	credentialsPath := filepath.Join(homeDir, ".connectcli", "credentials")
	file, err := os.Open(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
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
		return nil, err
	}
	return creds, nil
}

// MergeCredentials overlays non-empty fields from patch onto base.
func MergeCredentials(base, patch *Credentials) *Credentials {
	out := *base
	if patch.Session != "" {
		out.Session = patch.Session
	}
	if patch.CSRF != "" {
		out.CSRF = patch.CSRF
	}
	if patch.Jira != "" {
		out.Jira = patch.Jira
	}
	if patch.Authorization != "" {
		out.Authorization = patch.Authorization
	}
	if patch.SlackWebhook != "" {
		out.SlackWebhook = patch.SlackWebhook
	}
	if patch.SlackUserToken != "" {
		out.SlackUserToken = patch.SlackUserToken
	}
	return &out
}

// SaveCredentials writes ~/.connectcli/credentials (0600). Merges with existing file so optional keys are preserved when not set in c.
func SaveCredentials(c *Credentials) error {
	existing, err := LoadCredentialsOptional()
	if err != nil {
		return err
	}
	merged := MergeCredentials(existing, c)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(homeDir, ".connectcli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	path := filepath.Join(dir, "credentials")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}
	defer f.Close()

	writeLine := func(k, v string) error {
		if v == "" {
			return nil
		}
		_, err := fmt.Fprintf(f, "%s=%s\n", k, v)
		return err
	}
	if merged.Session != "" {
		if err := writeLine("session", merged.Session); err != nil {
			return err
		}
	}
	if merged.CSRF != "" {
		if err := writeLine("csrf", merged.CSRF); err != nil {
			return err
		}
	}
	if err := writeLine("jira", merged.Jira); err != nil {
		return err
	}
	if err := writeLine("authorization", merged.Authorization); err != nil {
		return err
	}
	if err := writeLine("slack_webhook", merged.SlackWebhook); err != nil {
		return err
	}
	if err := writeLine("slack_user_token", merged.SlackUserToken); err != nil {
		return err
	}
	return nil
}
