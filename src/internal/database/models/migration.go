package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Migration represents a data migration operation
type Migration struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	Type           string     `gorm:"size:50;not null"`        // opengist, github, gitlab, etc.
	Status         string     `gorm:"size:20;not null"`        // pending, running, completed, failed, cancelled
	SourceURL      string     `gorm:"size:500"`                // Source database/API URL
	SourceUsername string     `gorm:"size:255"`                // Username for source (if applicable)
	ItemsTotal     int        `gorm:"default:0"`               // Total items to migrate
	ItemsProcessed int        `gorm:"default:0"`               // Items successfully processed
	ItemsSkipped   int        `gorm:"default:0"`               // Items skipped
	ErrorCount     int        `gorm:"default:0"`               // Number of errors encountered
	ErrorDetails   string     `gorm:"type:text"`               // JSON array of error details
	Settings       string     `gorm:"type:text"`               // JSON of migration settings
	Result         string     `gorm:"type:text"`               // JSON result data
	StartedAt      time.Time  `gorm:"not null"`                // When migration started
	CompletedAt    *time.Time                                  // When migration completed
	CreatedBy      uuid.UUID  `gorm:"type:uuid;not null"`      // User who initiated migration
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relations
	Creator User `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook
func (m *Migration) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// Duration returns the duration of the migration
func (m *Migration) Duration() time.Duration {
	if m.CompletedAt == nil {
		return time.Since(m.StartedAt)
	}
	return m.CompletedAt.Sub(m.StartedAt)
}

// IsCompleted returns true if the migration is completed (success or failed)
func (m *Migration) IsCompleted() bool {
	return m.Status == "completed" || m.Status == "failed" || m.Status == "cancelled"
}

// SuccessRate returns the success rate as a percentage
func (m *Migration) SuccessRate() float64 {
	if m.ItemsTotal == 0 {
		return 0
	}
	return float64(m.ItemsProcessed) / float64(m.ItemsTotal) * 100
}