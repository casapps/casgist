package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GDPRExportRequest represents a GDPR data export request
type GDPRExportRequest struct {
	ID             uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	UserID         uuid.UUID  `gorm:"type:char(36);not null;index" json:"user_id"`
	Status         string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, processing, completed, failed
	ExportFilePath string     `gorm:"type:varchar(255)" json:"export_file_path,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// GDPRDeletionRequest represents a GDPR data deletion request
type GDPRDeletionRequest struct {
	ID              uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	UserID          uuid.UUID  `gorm:"type:char(36);not null;index" json:"user_id"`
	Status          string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, processing, completed, failed
	DeletionReason  string     `gorm:"type:text" json:"deletion_reason"`
	ProcessedByID   *uuid.UUID `gorm:"type:char(36)" json:"processed_by_id,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	User        User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ProcessedBy *User `gorm:"foreignKey:ProcessedByID" json:"processed_by,omitempty"`
}

// CustomDomain represents a custom domain configuration
type CustomDomain struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	UserID           *uuid.UUID     `gorm:"type:char(36);index" json:"user_id,omitempty"`
	OrganizationID   *uuid.UUID     `gorm:"type:char(36);index" json:"organization_id,omitempty"`
	Domain           string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"domain"`
	IsVerified       bool           `gorm:"default:false" json:"is_verified"`
	VerificationCode string         `gorm:"type:varchar(255)" json:"verification_code,omitempty"`
	SSLEnabled       bool           `gorm:"default:false" json:"ssl_enabled"`
	SSLCertificate   string         `gorm:"type:text" json:"ssl_certificate,omitempty"`
	SSLPrivateKey    string         `gorm:"type:text" json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

// BeforeCreate hooks for compliance models
func (e *GDPRExportRequest) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (d *GDPRDeletionRequest) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (c *CustomDomain) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.VerificationCode == "" {
		c.VerificationCode = uuid.New().String()
	}
	return nil
}