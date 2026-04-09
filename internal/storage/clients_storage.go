package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"connectcli/internal/api"
)

type ClientsStorage struct {
	dirPath string
	clients map[string]api.Client // id -> client for quick lookup
}

func NewClientsStorage() *ClientsStorage {
	return &ClientsStorage{
		dirPath: "clients",
		clients: make(map[string]api.Client),
	}
}

// ensureDirectory ensures the clients directory exists
func (c *ClientsStorage) ensureDirectory() error {
	if err := os.MkdirAll(c.dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create clients directory: %w", err)
	}
	return nil
}

// sanitizeFileName creates a safe filename from client name
func (c *ClientsStorage) sanitizeFileName(name string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	sanitized := name
	for _, char := range invalidChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	// Limit length
	if len(sanitized) > 100 {
		sanitized = sanitized[:100]
	}
	return sanitized
}

// LoadClients loads existing clients from individual files
func (c *ClientsStorage) LoadClients() error {
	if err := c.ensureDirectory(); err != nil {
		return err
	}

	// Check if directory exists
	if _, err := os.Stat(c.dirPath); os.IsNotExist(err) {
		// Directory doesn't exist, start with empty map
		return nil
	}

	// Read all files in the directory
	files, err := os.ReadDir(c.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read clients directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(c.dirPath, file.Name())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		var client api.Client
		if err := json.Unmarshal(fileContent, &client); err != nil {
			// Skip files that can't be parsed
			continue
		}

		// Only add if client has valid ID and name
		if client.ID != "" && client.Name != "" {
			c.clients[client.ID] = client
		}
	}

	return nil
}

// SaveClients saves clients to individual files
func (c *ClientsStorage) SaveClients() error {
	if err := c.ensureDirectory(); err != nil {
		return err
	}

	// Save each client to its own file
	for _, client := range c.clients {
		if err := c.saveClient(client); err != nil {
			return fmt.Errorf("failed to save client %s: %w", client.ID, err)
		}
	}

	return nil
}

// saveClient saves a single client to a file
func (c *ClientsStorage) saveClient(client api.Client) error {
	// Create filename from client name
	fileName := c.sanitizeFileName(client.Name) + ".json"
	filePath := filepath.Join(c.dirPath, fileName)

	// Marshal client to JSON
	jsonData, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write client file: %w", err)
	}

	return nil
}

// AddClients adds new clients, preventing duplicates
func (c *ClientsStorage) AddClients(newClients []api.Client) (int, int) {
	added := 0
	duplicates := 0

	for _, client := range newClients {
		if _, exists := c.clients[client.ID]; exists {
			duplicates++
		} else {
			c.clients[client.ID] = client
			added++
		}
	}

	return added, duplicates
}

// GetTotalClients returns the total number of stored clients
func (c *ClientsStorage) GetTotalClients() int {
	return len(c.clients)
}

// GetClients returns all clients as a slice
func (c *ClientsStorage) GetClients() []api.Client {
	var clients []api.Client
	for _, client := range c.clients {
		clients = append(clients, client)
	}
	return clients
}

// GetClientByID returns a specific client by ID
func (c *ClientsStorage) GetClientByID(id string) (api.Client, bool) {
	client, exists := c.clients[id]
	return client, exists
}

// GetClientByName returns a client by name (case-insensitive)
func (c *ClientsStorage) GetClientByName(name string) (api.Client, bool) {
	for _, client := range c.clients {
		if client.Name == name {
			return client, true
		}
	}
	return api.Client{}, false
}
