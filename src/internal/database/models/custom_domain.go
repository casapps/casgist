package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomDomain represents a custom domain for a user or organization
type CustomDomain struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key"`
	Domain            string     `gorm:"uniqueIndex;size:255;not null"`
	UserID            *uuid.UUID `gorm:"type:uuid"`
	OrganizationID    *uuid.UUID `gorm:"type:uuid"`
	Verified          bool       `gorm:"default:false"`
	VerificationToken string     `gorm:"size:64;not null"`
	VerifiedAt        *time.Time
	SSLEnabled        bool       `gorm:"default:false"`
	SSLCertPath       string     `gorm:"size:255"`
	SSLKeyPath        string     `gorm:"size:255"`
	SSLExpiresAt      *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time

	// Relations
	User         *User         `gorm:"constraint:OnDelete:CASCADE"`
	Organization *Organization `gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook
func (d *CustomDomain) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	// Validate ownership
	if (d.UserID == nil && d.OrganizationID == nil) || 
	   (d.UserID != nil && d.OrganizationID != nil) {
		return gorm.ErrInvalidData
	}
	return nil
}