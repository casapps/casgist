package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Gist represents a code snippet collection
type Gist struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	UserID           uuid.UUID      `gorm:"type:char(36);index" json:"user_id"`
	OrganizationID   *uuid.UUID     `gorm:"type:char(36);index" json:"organization_id,omitempty"`
	Title            string         `gorm:"type:varchar(255);not null" json:"title"`
	Description      string         `gorm:"type:text" json:"description"`
	IsPublic         bool           `gorm:"default:true" json:"is_public"`
	IsAnonymous      bool           `gorm:"default:false" json:"is_anonymous"`
	ViewCount        int64          `gorm:"default:0" json:"view_count"`
	CloneCount       int64          `gorm:"default:0" json:"clone_count"`
	StarCount        int64          `gorm:"default:0" json:"star_count"`
	ForkCount        int64          `gorm:"default:0" json:"fork_count"`
	Language         string         `gorm:"type:varchar(50)" json:"language"` // Primary language
	ForkedFromID     *uuid.UUID     `gorm:"type:char(36);index" json:"forked_from_id,omitempty"`
	ImportedFrom     string         `gorm:"type:varchar(50)" json:"imported_from,omitempty"` // github, gitlab, opengist
	ImportedURL      string         `gorm:"type:varchar(255)" json:"imported_url,omitempty"`
	CustomDomain     string         `gorm:"type:varchar(255);uniqueIndex:idx_custom_domain,where:custom_domain IS NOT NULL AND custom_domain != ''" json:"custom_domain,omitempty"`
	Metadata         string         `gorm:"type:text" json:"metadata,omitempty"` // JSON metadata
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Files        []GistFile    `gorm:"foreignKey:GistID" json:"files"`
	Stars        []Star        `gorm:"foreignKey:GistID" json:"-"`
	Comments     []Comment     `gorm:"foreignKey:GistID" json:"-"`
	Forks        []Gist        `gorm:"foreignKey:ForkedFromID" json:"-"`
	ForkedFrom   *Gist         `gorm:"foreignKey:ForkedFromID" json:"forked_from,omitempty"`
	Transfers    []Transfer    `gorm:"foreignKey:GistID" json:"-"`
}

// BeforeCreate hook to set UUID
func (g *Gist) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

// GistFile represents a file within a gist
type GistFile struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	GistID     uuid.UUID      `gorm:"type:char(36);not null;index" json:"gist_id"`
	Filename   string         `gorm:"type:varchar(255);not null" json:"filename"`
	Language   string         `gorm:"type:varchar(50)" json:"language"`
	Content    string         `gorm:"type:longtext" json:"content"`
	Size       int64          `gorm:"default:0" json:"size"`
	LineCount  int64          `gorm:"default:0" json:"line_count"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Gist Gist `gorm:"foreignKey:GistID" json:"-"`
}

// BeforeCreate hook to set UUID
func (f *GistFile) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	// Calculate size and line count
	f.Size = int64(len(f.Content))
	f.LineCount = int64(countLines(f.Content))
	return nil
}

// Star represents a star on a gist
type Star struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
	GistID    uuid.UUID `gorm:"type:char(36);not null;index"`
	CreatedAt time.Time

	// Relationships
	User User `gorm:"foreignKey:UserID"`
	Gist Gist `gorm:"foreignKey:GistID"`
}

// BeforeCreate hook to set UUID
func (s *Star) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// Comment represents a comment on a gist
type Comment struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	UserID    uuid.UUID      `gorm:"type:char(36);not null;index" json:"user_id"`
	GistID    uuid.UUID      `gorm:"type:char(36);not null;index" json:"gist_id"`
	Body      string         `gorm:"type:text;not null" json:"body"`
	IsEdited  bool           `gorm:"default:false" json:"is_edited"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Gist Gist `gorm:"foreignKey:GistID" json:"-"`
}

// BeforeCreate hook to set UUID
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// countLines counts the number of lines in a string
func countLines(s string) int {
	if s == "" {
		return 0
	}
	lines := 1
	for _, ch := range s {
		if ch == '\n' {
			lines++
		}
	}
	return lines
}