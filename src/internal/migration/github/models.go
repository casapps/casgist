package github

import (
	"time"
)

// GitHubGist represents a gist from GitHub API
type GitHubGist struct {
	ID          string                 `json:"id"`
	URL         string                 `json:"url"`
	HTMLURL     string                 `json:"html_url"`
	Description string                 `json:"description"`
	Public      bool                   `json:"public"`
	Owner       *GitHubUser            `json:"owner"`
	Files       map[string]GitHubFile  `json:"files"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Comments    int                    `json:"comments"`
	CommentsURL string                 `json:"comments_url"`
}

// GitHubUser represents a user from GitHub API
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	Type      string `json:"type"`
}

// GitHubFile represents a file in a gist
type GitHubFile struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Language string `json:"language"`
	RawURL   string `json:"raw_url"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
}

// GitHubComment represents a gist comment
type GitHubComment struct {
	ID        int        `json:"id"`
	URL       string     `json:"url"`
	Body      string     `json:"body"`
	User      GitHubUser `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}


// URLTransformRule defines how to transform GitHub URLs to CasGists URLs
type URLTransformRule struct {
	Pattern     string // Regex pattern to match
	Replacement string // Replacement pattern
	Description string // Human-readable description
}

// DefaultURLTransformRules returns the default URL transformation rules
func DefaultURLTransformRules(baseURL string) []URLTransformRule {
	return []URLTransformRule{
		{
			Pattern:     `https://gist\.github\.com/([^/]+)/([a-f0-9]+)`,
			Replacement: baseURL + "/u/$1/$2",
			Description: "Transform GitHub Gist URLs to CasGists URLs",
		},
		{
			Pattern:     `https://gist\.githubusercontent\.com/([^/]+)/([a-f0-9]+)/raw/[^/]+/(.+)`,
			Replacement: baseURL + "/u/$1/$2/raw/$3",
			Description: "Transform GitHub Gist raw file URLs",
		},
		{
			Pattern:     `\[gist:([a-f0-9]+)\]`,
			Replacement: "[gist:$1]",
			Description: "Preserve gist embed syntax",
		},
	}
}