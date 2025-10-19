package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"connectcli/internal/api"
)

type JiraStorage struct {
	dirPath string
	tickets map[string]api.JiraTicket // key -> ticket for quick lookup
}

func NewJiraStorage() *JiraStorage {
	// Store in jira-tickets directory
	dirPath := "jira-tickets"
	return &JiraStorage{
		dirPath: dirPath,
		tickets: make(map[string]api.JiraTicket),
	}
}

// ensureDirectory ensures the jira-tickets directory exists
func (j *JiraStorage) ensureDirectory() error {
	if err := os.MkdirAll(j.dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create jira-tickets directory: %w", err)
	}
	return nil
}

// sanitizeFileName creates a safe filename from ticket key and summary
func (j *JiraStorage) sanitizeFileName(key string, summary string) string {
	// Create filename format: "ticket summary - ticket key"
	fileName := fmt.Sprintf("%s - %s", summary, key)

	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r"}
	sanitized := fileName
	for _, char := range invalidChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	// Remove all other control characters and excessive whitespace
	sanitized = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1 // remove control chars
		}
		return r
	}, sanitized)
	// Collapse multiple spaces
	sanitized = strings.Join(strings.Fields(sanitized), " ")
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	// Limit length
	if len(sanitized) > 150 {
		sanitized = sanitized[:150]
	}
	return sanitized
}

// LoadTickets loads existing tickets from individual files
func (j *JiraStorage) LoadTickets() error {
	if err := j.ensureDirectory(); err != nil {
		return err
	}

	// Check if directory exists
	if _, err := os.Stat(j.dirPath); os.IsNotExist(err) {
		// Directory doesn't exist, start with empty map
		return nil
	}

	// Read all files in the directory
	files, err := os.ReadDir(j.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read jira-tickets directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(j.dirPath, file.Name())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		var ticket api.JiraTicket
		if err := json.Unmarshal(fileContent, &ticket); err != nil {
			// Skip files that can't be parsed
			continue
		}

		// Only add if ticket has valid key and summary
		if ticket.Key != "" && ticket.Summary != "" {
			j.tickets[ticket.Key] = ticket
		}
	}

	return nil
}

// SaveTickets saves tickets to individual files
func (j *JiraStorage) SaveTickets() error {
	if err := j.ensureDirectory(); err != nil {
		return err
	}

	// Save each ticket to its own file
	for _, ticket := range j.tickets {
		if err := j.saveTicket(ticket); err != nil {
			return fmt.Errorf("failed to save ticket %s: %w", ticket.Key, err)
		}
	}

	return nil
}

// saveTicket saves a single ticket to a file
func (j *JiraStorage) saveTicket(ticket api.JiraTicket) error {
	// Create filename from ticket key and summary
	fileName := j.sanitizeFileName(ticket.Key, ticket.Summary) + ".json"
	filePath := filepath.Join(j.dirPath, fileName)

	// Marshal ticket to JSON
	jsonData, err := json.MarshalIndent(ticket, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write ticket file: %w", err)
	}

	return nil
}

// AddTickets adds new tickets, preventing duplicates
func (j *JiraStorage) AddTickets(newTickets []api.JiraTicket) (int, int) {
	added := 0
	duplicates := 0

	for _, ticket := range newTickets {
		if _, exists := j.tickets[ticket.Key]; exists {
			duplicates++
		} else {
			j.tickets[ticket.Key] = ticket
			added++
		}
	}

	return added, duplicates
}

// GetTotalTickets returns the total number of stored tickets
func (j *JiraStorage) GetTotalTickets() int {
	return len(j.tickets)
}

// GetTickets returns all tickets as a slice
func (j *JiraStorage) GetTickets() []api.JiraTicket {
	var tickets []api.JiraTicket
	for _, ticket := range j.tickets {
		tickets = append(tickets, ticket)
	}
	return tickets
}

// GetNextStartAt returns the next startAt value for pagination
func (j *JiraStorage) GetNextStartAt() int {
	return len(j.tickets)
}
