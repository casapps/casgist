package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a GitLab API client
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// NewClient creates a new GitLab API client
func NewClient(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:   token,
		baseURL: baseURL,
	}
}

// GitLabUser represents a GitLab user
type GitLabUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

// GitLabSnippet represents a GitLab snippet
type GitLabSnippet struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	WebURL      string    `json:"web_url"`
	RawURL      string    `json:"raw_url"`
}

// ListSnippetsOptions contains options for listing snippets
type ListSnippetsOptions struct {
	PerPage int
	Page    int
}

// doRequest performs an authenticated API request
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v4%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("GitLab API error: %s - %s", resp.Status, string(body))
	}
	
	return resp, nil
}

// GetUser gets a user by username
func (c *Client) GetUser(ctx context.Context, username string) (*GitLabUser, error) {
	path := fmt.Sprintf("/users?username=%s", username)
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var users []GitLabUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	
	return &users[0], nil
}

// ListSnippets lists snippets for a user
func (c *Client) ListSnippets(ctx context.Context, username string, opts *ListSnippetsOptions) ([]*GitLabSnippet, error) {
	if opts == nil {
		opts = &ListSnippetsOptions{PerPage: 100, Page: 1}
	}
	
	// Get user first to get their ID
	user, err := c.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	
	path := fmt.Sprintf("/users/%d/snippets?per_page=%d&page=%d", user.ID, opts.PerPage, opts.Page)
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var snippets []*GitLabSnippet
	if err := json.NewDecoder(resp.Body).Decode(&snippets); err != nil {
		return nil, err
	}
	
	return snippets, nil
}