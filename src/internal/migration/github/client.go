package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	githubAPIBase = "https://api.github.com"
	perPage       = 100 // Maximum allowed by GitHub API
)

// Client represents a GitHub API client
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:   token,
		baseURL: githubAPIBase,
	}
}

// doRequest performs an authenticated API request
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}
	
	return resp, nil
}

// GetAuthenticatedUser gets the authenticated user
func (c *Client) GetAuthenticatedUser(ctx context.Context) (*GitHubUser, error) {
	resp, err := c.doRequest(ctx, "GET", "/user", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

// ListGistsOptions contains options for listing gists
type ListGistsOptions struct {
	PerPage int
	Page    int
}

// ListGists lists gists for a user
func (c *Client) ListGists(ctx context.Context, username string, opts *ListGistsOptions) ([]*GitHubGist, error) {
	if opts == nil {
		opts = &ListGistsOptions{PerPage: perPage, Page: 1}
	}
	if opts.PerPage == 0 {
		opts.PerPage = perPage
	}
	if opts.Page == 0 {
		opts.Page = 1
	}
	
	path := fmt.Sprintf("/users/%s/gists?per_page=%d&page=%d", username, opts.PerPage, opts.Page)
	if username == "" {
		// List authenticated user's gists
		path = fmt.Sprintf("/gists?per_page=%d&page=%d", opts.PerPage, opts.Page)
	}
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var gists []*GitHubGist
	if err := json.NewDecoder(resp.Body).Decode(&gists); err != nil {
		return nil, err
	}
	
	return gists, nil
}

// GetUser gets a user by username
func (c *Client) GetUser(ctx context.Context, username string) (*GitHubUser, error) {
	path := fmt.Sprintf("/users/%s", username)
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

// ListStarredGists lists starred gists for the authenticated user
func (c *Client) ListStarredGists(ctx context.Context, page int) ([]*GitHubGist, error) {
	path := fmt.Sprintf("/gists/starred?per_page=%d&page=%d", perPage, page)
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var gists []*GitHubGist
	if err := json.NewDecoder(resp.Body).Decode(&gists); err != nil {
		return nil, err
	}
	
	return gists, nil
}

// GetGist gets a specific gist with full details
func (c *Client) GetGist(ctx context.Context, gistID string) (*GitHubGist, error) {
	path := fmt.Sprintf("/gists/%s", gistID)
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var gist GitHubGist
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, err
	}
	
	return &gist, nil
}

// GetGistComments gets comments for a gist
func (c *Client) GetGistComments(ctx context.Context, gistID string) ([]*GitHubComment, error) {
	path := fmt.Sprintf("/gists/%s/comments", gistID)
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var comments []*GitHubComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}
	
	return comments, nil
}

// GetRateLimit gets the current rate limit status
func (c *Client) GetRateLimit(ctx context.Context) (remaining int, resetTime time.Time, err error) {
	resp, err := c.doRequest(ctx, "GET", "/rate_limit", nil)
	if err != nil {
		return 0, time.Time{}, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Rate struct {
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"rate"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, time.Time{}, err
	}
	
	return result.Rate.Remaining, time.Unix(result.Rate.Reset, 0), nil
}