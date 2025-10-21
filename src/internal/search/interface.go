package search

import (
	"github.com/google/uuid"
)

// Result represents a search result
type Result struct {
	GistID      uuid.UUID              `json:"gist_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Excerpt     string                 `json:"excerpt"`
	Score       float64                `json:"score"`
	User        map[string]interface{} `json:"user"`
	Visibility  string                 `json:"visibility"`
	Language    string                 `json:"language"`
	FileCount   int                    `json:"file_count"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// SearchResults represents search results with metadata
type SearchResults struct {
	Query        string                 `json:"query"`
	TotalResults int                    `json:"total_results"`
	SearchTime   string                 `json:"search_time"`
	Results      []Result               `json:"results"`
	Facets       map[string]interface{} `json:"facets,omitempty"`
}

// Engine defines the search engine interface
type Engine interface {
	// Search performs a search query
	Search(query string, page, limit int) (*SearchResults, error)
	
	// Index adds or updates a document in the search index
	Index(gistID uuid.UUID, data map[string]interface{}) error
	
	// Delete removes a document from the search index
	Delete(gistID uuid.UUID) error
	
	// Reindex rebuilds the entire search index
	Reindex() error
}