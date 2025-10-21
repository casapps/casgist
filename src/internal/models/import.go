package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ImportJob represents an import job from external sources
type ImportJob struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	UserID           uuid.UUID      `gorm:"type:char(36);not null;index" json:"user_id"`
	Source           string         `gorm:"type:varchar(50);not null" json:"source"` // github, gitlab, opengist
	Status           string         `gorm:"type:varchar(50);not null;default:'pending'" json:"status"` // pending, running, completed, failed, cancelled
	Type             string         `gorm:"type:varchar(50);not null" json:"type"` // all, selected
	TotalGists       int64          `json:"total_gists"`
	ImportedGists    int64          `json:"imported_gists"`
	FailedGists      int64          `json:"failed_gists"`
	Configuration    string         `gorm:"type:text" json:"-"` // JSON configuration
	Errors           string         `gorm:"type:text" json:"errors,omitempty"` // JSON array of errors
	MappingData      string         `gorm:"type:text" json:"-"` // JSON mapping of old IDs to new IDs
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate hook to set UUID
func (i *ImportJob) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}