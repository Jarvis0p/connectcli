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
	dirPath       string
	tickets       map[string]api.JiraTicket // key -> ticket for quick lookup
	nextPageToken string
}

func NewJiraStorage() *JiraStorage {
	return &JiraStorage{
		dirPath: "jira-tickets",
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

func (j *JiraStorage) sanitizeFileName(key string, summary string) string {
	name := fmt.Sprintf("%s - %s.json", key, summary)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	return name
}

// LoadTickets loads tickets from disk into memory
func (j *JiraStorage) LoadTickets() error {
	if err := j.ensureDirectory(); err != nil {
		return err
	}

	entries, err := os.ReadDir(j.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read jira-tickets directory: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(j.dirPath, e.Name()))
		if err != nil {
			continue
		}
		var t api.JiraTicket
		if err := json.Unmarshal(content, &t); err == nil && t.Key != "" {
			j.tickets[t.Key] = t
		}
	}
	return nil
}

// SaveTickets saves in-memory tickets to disk
func (j *JiraStorage) SaveTickets() error {
	if err := j.ensureDirectory(); err != nil {
		return err
	}
	for _, t := range j.tickets {
		if err := j.saveTicket(t); err != nil {
			return err
		}
	}
	return nil
}

func (j *JiraStorage) saveTicket(ticket api.JiraTicket) error {
	fileName := j.sanitizeFileName(ticket.Key, ticket.Summary)
	path := filepath.Join(j.dirPath, fileName)
	data, _ := json.MarshalIndent(ticket, "", "  ")
	return os.WriteFile(path, data, 0644)
}

// AddTickets adds tickets and returns (added, duplicates)
func (j *JiraStorage) AddTickets(newTickets []api.JiraTicket) (int, int) {
	added := 0
	dups := 0
	for _, t := range newTickets {
		if _, exists := j.tickets[t.Key]; exists {
			dups++
			continue
		}
		j.tickets[t.Key] = t
		added++
	}
	return added, dups
}

func (j *JiraStorage) GetTotalTickets() int {
	return len(j.tickets)
}

func (j *JiraStorage) GetTickets() []api.JiraTicket {
	var out []api.JiraTicket
	for _, t := range j.tickets {
		out = append(out, t)
	}
	return out
}

// Pagination helpers
func (j *JiraStorage) GetNextPageToken() string      { return j.nextPageToken }
func (j *JiraStorage) SetNextPageToken(token string) { j.nextPageToken = token }
