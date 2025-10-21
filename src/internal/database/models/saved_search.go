package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SavedSearch represents a user's saved search query
type SavedSearch struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Query       string     `gorm:"type:text;not null" json:"query"`
	Filters     *JSON      `gorm:"type:jsonb" json:"filters,omitempty"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	IsPublic    bool       `gorm:"default:false" json:"is_public"`
	UsageCount  int        `gorm:"default:0" json:"usage_count"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// SearchHistory represents a user's search history
type SearchHistory struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Query      string    `gorm:"type:text;not null" json:"query"`
	ResultCount int      `gorm:"default:0" json:"result_count"`
	SearchType string    `gorm:"type:varchar(50)" json:"search_type"` // gist, user, organization
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// SearchIndexMetadata tracks search index status and performance
type SearchIndexMetadata struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IndexName     string     `gorm:"type:varchar(255);unique;not null" json:"index_name"`
	EntityType    string     `gorm:"type:varchar(50);not null" json:"entity_type"` // gist, user, file
	LastIndexedAt *time.Time `json:"last_indexed_at,omitempty"`
	DocumentCount int64      `gorm:"default:0" json:"document_count"`
	IndexSize     int64      `gorm:"default:0" json:"index_size"` // in bytes
	Status        string     `gorm:"type:varchar(50);default:'active'" json:"status"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// CreateSavedSearch creates a new saved search for a user
func CreateSavedSearch(db *gorm.DB, search *SavedSearch) error {
	return db.Create(search).Error
}

// GetUserSavedSearches retrieves all saved searches for a user
func GetUserSavedSearches(db *gorm.DB, userID uuid.UUID) ([]SavedSearch, error) {
	var searches []SavedSearch
	err := db.Where("user_id = ?", userID).
		Order("usage_count DESC, created_at DESC").
		Find(&searches).Error
	return searches, err
}

// GetPublicSavedSearches retrieves public saved searches
func GetPublicSavedSearches(db *gorm.DB, limit, offset int) ([]SavedSearch, error) {
	var searches []SavedSearch
	err := db.Preload("User").
		Where("is_public = ?", true).
		Order("usage_count DESC").
		Limit(limit).
		Offset(offset).
		Find(&searches).Error
	return searches, err
}

// UseSavedSearch increments the usage count and updates last used time
func UseSavedSearch(db *gorm.DB, searchID uuid.UUID) error {
	now := time.Now()
	return db.Model(&SavedSearch{}).
		Where("id = ?", searchID).
		Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + ?", 1),
			"last_used_at": now,
		}).Error
}

// DeleteSavedSearch deletes a saved search
func DeleteSavedSearch(db *gorm.DB, userID, searchID uuid.UUID) error {
	return db.Where("id = ? AND user_id = ?", searchID, userID).
		Delete(&SavedSearch{}).Error
}

// RecordSearchHistory records a search in user's history
func RecordSearchHistory(db *gorm.DB, history *SearchHistory) error {
	// Limit history to last 100 searches per user
	var count int64
	db.Model(&SearchHistory{}).Where("user_id = ?", history.UserID).Count(&count)

	if count >= 100 {
		// Delete oldest entries
		var oldestID uuid.UUID
		db.Model(&SearchHistory{}).
			Where("user_id = ?", history.UserID).
			Order("created_at ASC").
			Limit(1).
			Pluck("id", &oldestID)

		if oldestID != uuid.Nil {
			db.Delete(&SearchHistory{}, "id = ?", oldestID)
		}
	}

	return db.Create(history).Error
}

// GetSearchHistory retrieves user's search history
func GetSearchHistory(db *gorm.DB, userID uuid.UUID, limit int) ([]SearchHistory, error) {
	var history []SearchHistory
	err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// ClearSearchHistory clears all search history for a user
func ClearSearchHistory(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("user_id = ?", userID).Delete(&SearchHistory{}).Error
}

// GetPopularSearches gets the most popular search queries
func GetPopularSearches(db *gorm.DB, limit int, period time.Duration) ([]map[string]interface{}, error) {
	cutoff := time.Now().Add(-period)

	var results []map[string]interface{}
	err := db.Model(&SearchHistory{}).
		Select("query, COUNT(*) as count").
		Where("created_at > ?", cutoff).
		Group("query").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}

// UpdateSearchIndex updates search index metadata
func UpdateSearchIndex(db *gorm.DB, indexName, entityType string, docCount, size int64) error {
	now := time.Now()

	var metadata SearchIndexMetadata
	err := db.Where("index_name = ?", indexName).First(&metadata).Error

	if err == gorm.ErrRecordNotFound {
		// Create new metadata
		metadata = SearchIndexMetadata{
			IndexName:     indexName,
			EntityType:    entityType,
			LastIndexedAt: &now,
			DocumentCount: docCount,
			IndexSize:     size,
			Status:        "active",
		}
		return db.Create(&metadata).Error
	}

	// Update existing metadata
	return db.Model(&metadata).Updates(map[string]interface{}{
		"last_indexed_at": now,
		"document_count":  docCount,
		"index_size":      size,
		"updated_at":      now,
	}).Error
}

// GetSearchIndexStatus retrieves the status of all search indexes
func GetSearchIndexStatus(db *gorm.DB) ([]SearchIndexMetadata, error) {
	var metadata []SearchIndexMetadata
	err := db.Order("entity_type, index_name").Find(&metadata).Error
	return metadata, err
}

// GetSearchSuggestions provides search suggestions based on history and saved searches
func GetSearchSuggestions(db *gorm.DB, userID uuid.UUID, prefix string, limit int) ([]string, error) {
	var suggestions []string

	// Get from user's search history
	db.Model(&SearchHistory{}).
		Where("user_id = ? AND query LIKE ?", userID, prefix+"%").
		Order("created_at DESC").
		Limit(limit/2).
		Pluck("DISTINCT query", &suggestions)

	// Get from saved searches
	var savedSuggestions []string
	db.Model(&SavedSearch{}).
		Where("(user_id = ? OR is_public = ?) AND query LIKE ?", userID, true, prefix+"%").
		Order("usage_count DESC").
		Limit(limit/2).
		Pluck("DISTINCT query", &savedSuggestions)

	// Combine and deduplicate
	suggestionMap := make(map[string]bool)
	for _, s := range suggestions {
		suggestionMap[s] = true
	}
	for _, s := range savedSuggestions {
		suggestionMap[s] = true
	}

	// Convert back to slice
	suggestions = make([]string, 0, len(suggestionMap))
	for s := range suggestionMap {
		suggestions = append(suggestions, s)
		if len(suggestions) >= limit {
			break
		}
	}

	return suggestions, nil
}