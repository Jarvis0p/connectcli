package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type JiraClient struct {
	client  *http.Client
	baseURL string
	auth    string
}

type JiraSearchResponse struct {
	StartAt       int    `json:"startAt"`
	MaxResults    int    `json:"maxResults"`
	Total         int    `json:"total"`
	NextPageToken string `json:"nextPageToken"`
	Issues        []struct {
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
	// jiraToken format: email:apiToken
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(jiraToken))
	return &JiraClient{
		client: &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://aanyanshllc.atlassian.net",
		auth:    "Basic " + encodedAuth,
	}
}

// FetchJiraTickets fetches Jira tickets with pagination using nextPageToken
func (j *JiraClient) FetchJiraTickets(nextPageToken string, maxResults int) (*JiraSearchResponse, error) {
	endpoint, _ := url.Parse(j.baseURL + "/rest/api/3/search/jql")
	q := endpoint.Query()
	jql := "project = TECH ORDER BY created DESC"
	q.Set("jql", jql)
	q.Set("fields", "key,summary")
	q.Set("maxResults", fmt.Sprintf("%d", maxResults))
	if strings.TrimSpace(nextPageToken) != "" {
		q.Set("nextPageToken", nextPageToken)
	}
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", endpoint.String(), nil)
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
		tickets = append(tickets, JiraTicket{Key: issue.Key, Summary: issue.Fields.Summary})
	}
	return tickets
}





