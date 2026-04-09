package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"connectcli/internal/api"
	"connectcli/internal/paths"
)

// JiraHit is one matching ticket file.
type JiraHit struct {
	File    string
	Key     string
	Summary string
}

// JiraLocal searches jira-tickets/*.json for tickets whose key, summary, or filename
// contains query (case-insensitive substring).
func JiraLocal(query string) ([]JiraHit, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	needle := strings.ToLower(q)

	dir, err := paths.JiraTicketsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory %q does not exist (run fetch jira first)", dir)
		}
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	var hits []JiraHit
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var t api.JiraTicket
		if err := json.Unmarshal(data, &t); err != nil || t.Key == "" {
			continue
		}
		nameLower := strings.ToLower(e.Name())
		keyLower := strings.ToLower(t.Key)
		sumLower := strings.ToLower(t.Summary)
		if strings.Contains(nameLower, needle) || strings.Contains(keyLower, needle) || strings.Contains(sumLower, needle) {
			hits = append(hits, JiraHit{
				File:    path,
				Key:     t.Key,
				Summary: t.Summary,
			})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Key < hits[j].Key
	})

	return hits, nil
}
