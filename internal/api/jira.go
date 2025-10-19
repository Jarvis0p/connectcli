package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type JiraClient struct {
	client  *http.Client
	baseURL string
	auth    string
}

type JiraSearchResponse struct {
	StartAt    int `json:"startAt"`
	MaxResults int `json:"maxResults"`
	Total      int `json:"total"`
	Issues     []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary string `json:"summary"`
		} `json:"fields"`
	} `json:"issues"`
}

type JiraTicket struct {
	Key     string `json:"key"`
	Summary string `json:"summary"`
}

func NewJiraClient(jiraToken string) *JiraClient {
	// Parse the Jira token format: krishna@securify.llc:<token>
	parts := strings.Split(jiraToken, ":")
	if len(parts) != 2 {
		return nil
	}

	email := parts[0]
	token := parts[1]

	// Create basic auth header
	auth := email + ":" + token
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	return &JiraClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://aanyanshllc.atlassian.net",
		auth:    "Basic " + encodedAuth,
	}
}

// FetchJiraTickets fetches Jira tickets with pagination
func (j *JiraClient) FetchJiraTickets(startAt int, maxResults int) (*JiraSearchResponse, error) {
	url := fmt.Sprintf("%s/rest/api/3/search?jql=project=TECH&fields=key,summary&maxResults=%d&startAt=%d",
		j.baseURL, maxResults, startAt)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", j.auth)
	req.Header.Set("User-Agent", "ConnectCLI/1.0")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response JiraSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}

// ConvertToTickets converts the API response to a slice of tickets
func (j *JiraClient) ConvertToTickets(response *JiraSearchResponse) []JiraTicket {
	var tickets []JiraTicket
	for _, issue := range response.Issues {
		tickets = append(tickets, JiraTicket{
			Key:     issue.Key,
			Summary: issue.Fields.Summary,
		})
	}
	return tickets
}
