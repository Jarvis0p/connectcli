package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DataDir returns ~/.connectcli
func DataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".connectcli"), nil
}

// ClientsDir returns ./clients (relative to the process working directory).
func ClientsDir() (string, error) {
	return "clients", nil
}

// JiraTicketsDir returns ./jira-tickets (relative to the process working directory).
func JiraTicketsDir() (string, error) {
	return "jira-tickets", nil
}

// EnsureDataDir creates ~/.connectcli if missing.
func EnsureDataDir() error {
	d, err := DataDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(d, 0755)
}

// ResolveClientJSONPath resolves a user-provided client reference to a JSON file under ./clients/.
func ResolveClientJSONPath(userInput string) (string, error) {
	userInput = strings.TrimSpace(userInput)
	if userInput == "" {
		return "", fmt.Errorf("empty client path")
	}
	p := userInput
	p = strings.TrimPrefix(p, "clients/")
	p = strings.TrimPrefix(p, "./clients/")
	p = strings.TrimPrefix(p, `.\\clients\\`)
	name := filepath.Base(p)
	return filepath.Join("clients", name), nil
}
