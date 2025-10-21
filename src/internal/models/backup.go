package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Backup represents a system backup
type Backup struct {
	ID            uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Type          string         `gorm:"type:varchar(50);not null" json:"type"` // manual, scheduled
	Status        string         `gorm:"type:varchar(50);not null" json:"status"` // running, completed, failed
	Size          int64          `json:"size"` // bytes
	Duration      int64          `json:"duration"` // seconds
	FilePath      string         `gorm:"type:varchar(500)" json:"-"`
	IncludesData  bool           `gorm:"default:true" json:"includes_data"`
	IncludesRepos bool           `gorm:"default:true" json:"includes_repos"`
	Error         string         `gorm:"type:text" json:"error,omitempty"`
	Metadata      string         `gorm:"type:text" json:"metadata,omitempty"` // JSON metadata
	CreatedBy     *uuid.UUID     `gorm:"type:char(36)" json:"created_by,omitempty"`
	StartedAt     time.Time      `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// BeforeCreate hook to set UUID
func (b *Backup) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}