package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Backup represents a system backup
type Backup struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"`
	FilePath    string     `gorm:"size:500;not null"`      // Path to backup file
	Size        int64      `gorm:"not null"`               // Size in bytes
	Description string     `gorm:"size:1000"`              // User description
	Metadata    string     `gorm:"type:text"`              // JSON metadata
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null"`     // User who created backup
	CreatedAt   time.Time  `gorm:"not null"`               // When backup was created
	CompletedAt *time.Time                                 // When backup completed
	Status      string     `gorm:"size:20;default:'pending'"` // pending, running, completed, failed

	// Relations
	Creator User `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook
func (b *Backup) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Duration returns the time taken to create the backup
func (b *Backup) Duration() time.Duration {
	if b.CompletedAt == nil {
		return time.Since(b.CreatedAt)
	}
	return b.CompletedAt.Sub(b.CreatedAt)
}

// IsCompleted returns true if backup is completed
func (b *Backup) IsCompleted() bool {
	return b.Status == "completed"
}

// HumanReadableSize returns size in human readable format
func (b *Backup) HumanReadableSize() string {
	const unit = 1024
	if b.Size < unit {
		return fmt.Sprintf("%d B", b.Size)
	}
	div, exp := int64(unit), 0
	for n := b.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b.Size)/float64(div), "KMGTPE"[exp])
}