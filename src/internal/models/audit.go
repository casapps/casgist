package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID             uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	UserID         *uuid.UUID `gorm:"type:char(36);index" json:"user_id,omitempty"`
	OrganizationID *uuid.UUID `gorm:"type:char(36);index" json:"organization_id,omitempty"`
	Action         string     `gorm:"type:varchar(100);not null;index" json:"action"`
	ResourceType   string     `gorm:"type:varchar(50);index" json:"resource_type"`
	ResourceID     string     `gorm:"type:varchar(36);index" json:"resource_id"`
	Changes        string     `gorm:"type:text" json:"changes,omitempty"` // JSON diff
	IPAddress      string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent      string     `gorm:"type:text" json:"user_agent"`
	Success        bool       `gorm:"default:true" json:"success"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`
	Metadata       string     `gorm:"type:text" json:"metadata,omitempty"` // JSON metadata
	CreatedAt      time.Time  `json:"created_at"`

	// Relationships
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

// BeforeCreate hook to set UUID
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}