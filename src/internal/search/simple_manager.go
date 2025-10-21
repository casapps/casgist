package search

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SimpleSearchOptions provides basic search options
type SimpleSearchOptions struct {
	Query  string
	Page   int
	Limit  int
	UserID *uuid.UUID
}

// SimpleManager manages search operations using the simple Engine interface
type SimpleManager struct {
	engine Engine
}

// NewSimpleManager creates a new simple search manager
func NewSimpleManager(engine Engine) *SimpleManager {
	return &SimpleManager{
		engine: engine,
	}
}

// Search performs a search with the given query
func (m *SimpleManager) Search(query string, page, limit int) (*SearchResults, error) {
	return m.engine.Search(query, page, limit)
}

// SearchWithOptions performs a search with options
func (m *SimpleManager) SearchWithOptions(opts SimpleSearchOptions) (*SearchResults, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}
	
	results, err := m.engine.Search(opts.Query, opts.Page, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	
	// Filter by user if specified
	if opts.UserID != nil && results != nil {
		filtered := []Result{}
		for _, r := range results.Results {
			// Check if result belongs to the specified user
			if userID, ok := r.User["id"].(string); ok {
				if userID == opts.UserID.String() {
					filtered = append(filtered, r)
				}
			}
		}
		results.Results = filtered
		results.TotalResults = len(filtered)
	}
	
	return results, nil
}

// Index adds or updates a gist in the search index
func (m *SimpleManager) Index(gistID uuid.UUID, data map[string]interface{}) error {
	return m.engine.Index(gistID, data)
}

// Delete removes a gist from the search index
func (m *SimpleManager) Delete(gistID uuid.UUID) error {
	return m.engine.Delete(gistID)
}

// Reindex rebuilds the search index
func (m *SimpleManager) Reindex() error {
	return m.engine.Reindex()
}

// SuggestCompletion provides basic search suggestions
func (m *SimpleManager) SuggestCompletion(prefix string, limit int) ([]string, error) {
	// Simple implementation - search and extract titles
	results, err := m.engine.Search(prefix, 1, limit)
	if err != nil {
		return nil, err
	}
	
	suggestions := make([]string, 0, len(results.Results))
	for _, result := range results.Results {
		if result.Title != "" {
			suggestions = append(suggestions, result.Title)
		}
	}
	
	return suggestions, nil
}

// GetStats returns basic search statistics
func (m *SimpleManager) GetStats() (*SearchStats, error) {
	// Return basic stats
	stats := &SearchStats{
		Engine:      "SQLite",
		LastIndexed: nil,
	}
	
	// Try to get count of indexed items
	results, err := m.engine.Search("", 1, 1)
	if err == nil && results != nil {
		stats.IndexedItems = results.TotalResults
	}
	
	return stats, nil
}

// SearchStats represents search engine statistics
type SearchStats struct {
	Engine       string     `json:"engine"`
	IndexedItems int        `json:"indexed_items"`
	IndexSize    int64      `json:"index_size"`
	LastIndexed  *time.Time `json:"last_indexed,omitempty"`
}