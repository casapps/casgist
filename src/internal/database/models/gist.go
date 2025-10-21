package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Visibility type for gists
type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
	VisibilityUnlisted Visibility = "unlisted"
)

// Gist represents a code snippet
type Gist struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	Name           string     `gorm:"size:63"`
	Title          string     `gorm:"size:255;not null"`
	Description    string     `gorm:"size:1000"`
	Visibility     Visibility `gorm:"size:20;default:'private'"`
	UserID         *uuid.UUID `gorm:"type:uuid"`
	OrganizationID *uuid.UUID `gorm:"type:uuid"`
	ForkedFromID   *uuid.UUID `gorm:"type:uuid"`
	GitRepoPath    string     `gorm:"size:255;not null"`
	StarCount      int        `gorm:"default:0"`
	ForkCount      int        `gorm:"default:0"`
	ViewCount      int        `gorm:"default:0"`
	Language       string     `gorm:"size:50"`
	TagsString     string     `gorm:"size:500" json:"-"`
	ImportID       string     `gorm:"size:255;index"` // External ID for imported gists
	ImportURL      string     `gorm:"size:500"`       // Original URL for imported gists
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relations
	User         *User         `gorm:"constraint:OnDelete:CASCADE"`
	Organization *Organization `gorm:"constraint:OnDelete:CASCADE"`
	ForkedFrom   *Gist         `gorm:"foreignKey:ForkedFromID;constraint:OnDelete:SET NULL"`
	Files        []GistFile    `gorm:"constraint:OnDelete:CASCADE"`
	Stars        []GistStar    `gorm:"constraint:OnDelete:CASCADE"`
	Comments     []GistComment `gorm:"constraint:OnDelete:CASCADE"`
	Views        []GistView    `gorm:"constraint:OnDelete:CASCADE"`
	Watches      []GistWatch   `gorm:"constraint:OnDelete:CASCADE"`
	Forks        []Gist        `gorm:"foreignKey:ForkedFromID"`
	Tags         []Tag         `gorm:"many2many:gist_tags;" json:"tags"`
}

// GistFile represents a file within a gist
type GistFile struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	GistID    uuid.UUID `gorm:"type:uuid;not null"`
	Filename  string    `gorm:"size:255;not null"`
	Content   string    `gorm:"type:text"`
	Size      int64     `gorm:"default:0"`
	Language  string    `gorm:"size:50"`
	Lines     int       `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Gist Gist `gorm:"constraint:OnDelete:CASCADE"`
}

// GistStar represents a star on a gist
type GistStar struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	GistID    uuid.UUID `gorm:"type:uuid;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time

	// Relations
	Gist Gist `gorm:"constraint:OnDelete:CASCADE"`
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// GistComment represents a comment on a gist
type GistComment struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	GistID    uuid.UUID `gorm:"type:uuid;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Gist Gist `gorm:"constraint:OnDelete:CASCADE"`
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// GistView represents a view of a gist
type GistView struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key"`
	GistID    uuid.UUID  `gorm:"type:uuid;not null"`
	UserID    *uuid.UUID `gorm:"type:uuid"`
	IPAddress string     `gorm:"size:45"`
	UserAgent string     `gorm:"size:500"`
	Referrer  string     `gorm:"size:500"`
	CreatedAt time.Time

	// Relations
	Gist Gist  `gorm:"constraint:OnDelete:CASCADE"`
	User *User `gorm:"constraint:OnDelete:SET NULL"`
}

// GistWatch represents a watch subscription on a gist
type GistWatch struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	GistID    uuid.UUID `gorm:"type:uuid;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time

	// Relations
	Gist Gist `gorm:"constraint:OnDelete:CASCADE"`
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate hooks
func (g *Gist) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	// Ensure gist belongs to either user or organization
	if (g.UserID == nil && g.OrganizationID == nil) ||
		(g.UserID != nil && g.OrganizationID != nil) {
		return gorm.ErrInvalidData
	}
	return nil
}

func (f *GistFile) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	// Calculate file size and lines
	f.Size = int64(len(f.Content))
	f.Lines = countLines(f.Content)
	return nil
}

func (s *GistStar) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	// Check for duplicate
	result := tx.Where("gist_id = ? AND user_id = ?", s.GistID, s.UserID).First(&GistStar{})
	if result.Error == nil {
		return gorm.ErrDuplicatedKey
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	return nil
}

func (c *GistComment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (v *GistView) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

func (w *GistWatch) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	// Check for duplicate
	result := tx.Where("gist_id = ? AND user_id = ?", w.GistID, w.UserID).First(&GistWatch{})
	if result.Error == nil {
		return gorm.ErrDuplicatedKey
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	return nil
}

// AfterCreate/AfterDelete hooks to update counters
func (s *GistStar) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&Gist{}).Where("id = ?", s.GistID).
		UpdateColumn("star_count", gorm.Expr("star_count + ?", 1)).Error
}

func (s *GistStar) AfterDelete(tx *gorm.DB) error {
	return tx.Model(&Gist{}).Where("id = ?", s.GistID).
		UpdateColumn("star_count", gorm.Expr("star_count - ?", 1)).Error
}

// countLines counts the number of lines in a string
func countLines(s string) int {
	if s == "" {
		return 0
	}
	lines := 1
	for _, c := range s {
		if c == '\n' {
			lines++
		}
	}
	return lines
}
