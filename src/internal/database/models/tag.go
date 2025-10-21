package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag represents a tag for categorizing gists
type Tag struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key"`
	Name      string         `gorm:"size:50;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	
	// Relations
	Gists []Gist `gorm:"many2many:gist_tags;"`
}

// GistTag represents the many-to-many relationship between gists and tags
type GistTag struct {
	GistID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	
	// Relations
	Gist Gist `gorm:"constraint:OnDelete:CASCADE"`
	Tag  Tag  `gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate hooks
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}