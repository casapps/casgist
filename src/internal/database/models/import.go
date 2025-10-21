package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ImportStatus represents the status of an import job
type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusProcessing ImportStatus = "processing"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
	ImportStatusCancelled  ImportStatus = "cancelled"
)

// ImportJob represents a bulk import operation
type ImportJob struct {
	ID              uuid.UUID    `gorm:"type:uuid;primary_key"`
	Platform        string       `gorm:"size:50;not null"`     // github, gitlab, bitbucket
	Status          string       `gorm:"size:20;not null"`     // pending, running, completed, failed, cancelled  
	SourceURL       string       `gorm:"size:500"`             // Source platform URL
	SourceUsername  string       `gorm:"size:255"`             // Username on source platform
	ItemsTotal      int          `gorm:"default:0"`            // Total items to import
	ItemsImported   int          `gorm:"default:0"`            // Items successfully imported
	ErrorCount      int          `gorm:"default:0"`            // Number of errors
	Settings        string       `gorm:"type:text"`            // JSON settings
	Result          string       `gorm:"type:text"`            // JSON result data
	StartedAt       time.Time    `gorm:"not null"`             // When import started
	CompletedAt     *time.Time                                 // When import completed
	CreatedBy       uuid.UUID    `gorm:"type:uuid;not null"`   // User who started import
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Relations
	Creator User          `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
	Items   []ImportItem  `gorm:"constraint:OnDelete:CASCADE"`
}

// ImportItem represents a single item in an import job
type ImportItem struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key"`
	ImportJobID uuid.UUID    `gorm:"type:uuid;not null"`
	SourceURL   string       `gorm:"size:500"`
	SourceID    string       `gorm:"size:255"`
	GistID      *uuid.UUID   `gorm:"type:uuid"`
	Status      ImportStatus `gorm:"size:20;default:'pending'"`
	ErrorMessage string      `gorm:"type:text"`
	SourceData  string       `gorm:"type:jsonb"`
	CreatedAt   time.Time
	ProcessedAt *time.Time

	// Relations
	ImportJob ImportJob `gorm:"constraint:OnDelete:CASCADE"`
	Gist      *Gist     `gorm:"constraint:OnDelete:SET NULL"`
}

// BeforeCreate hooks
func (j *ImportJob) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

func (i *ImportItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}