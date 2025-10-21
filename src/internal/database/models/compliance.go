package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog records all significant actions for compliance
type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID       *uuid.UUID `gorm:"type:uuid"`
	Action       string     `gorm:"size:100;not null"`
	ResourceType string     `gorm:"size:50"`
	ResourceID   string     `gorm:"size:255"`
	Details      string     `gorm:"type:jsonb"`
	IPAddress    string     `gorm:"size:45"`
	UserAgent    string     `gorm:"size:500"`
	Success      bool       `gorm:"default:true"`
	ErrorMessage string     `gorm:"size:500"`
	CreatedAt    time.Time

	// Relations
	User *User `gorm:"constraint:OnDelete:SET NULL"`
}

// SecurityEvent represents security-related events
type SecurityEvent struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID      *uuid.UUID `gorm:"type:uuid"`
	EventType   string     `gorm:"size:50;not null"`
	Severity    string     `gorm:"size:20;not null"`
	Description string     `gorm:"size:500;not null"`
	Details     string     `gorm:"type:jsonb"`
	IPAddress   string     `gorm:"size:45"`
	UserAgent   string     `gorm:"size:500"`
	Resolved    bool       `gorm:"default:false"`
	ResolvedByID *uuid.UUID `gorm:"type:uuid"`
	ResolvedAt  *time.Time
	CreatedAt   time.Time

	// Relations
	User       *User `gorm:"constraint:OnDelete:SET NULL"`
	ResolvedBy *User `gorm:"foreignKey:ResolvedByID;constraint:OnDelete:SET NULL"`
}

// ComplianceLog tracks compliance-related actions
type ComplianceLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID         *uuid.UUID `gorm:"type:uuid"`
	ComplianceType string     `gorm:"size:20;not null"`
	Action         string     `gorm:"size:100;not null"`
	DataType       string     `gorm:"size:50"`
	Details        string     `gorm:"type:jsonb"`
	LegalBasis     string     `gorm:"size:100"`
	RetentionDays  int
	CreatedAt      time.Time

	// Relations
	User *User `gorm:"constraint:OnDelete:SET NULL"`
}


// GDPRExportRequest tracks user data export requests
type GDPRExportRequest struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	Status         string     `gorm:"size:20;default:'pending'"`
	ExportFilePath string     `gorm:"size:255"`
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	CompletedAt    *time.Time

	// Relations
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// GDPRDeletionRequest tracks user deletion requests
type GDPRDeletionRequest struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	Status         string     `gorm:"size:20;default:'pending'"`
	DeletionReason string     `gorm:"size:500"`
	ProcessedByID  *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time
	ProcessedAt    *time.Time

	// Relations
	User        User  `gorm:"constraint:OnDelete:CASCADE"`
	ProcessedBy *User `gorm:"foreignKey:ProcessedByID;constraint:OnDelete:SET NULL"`
}

// BeforeCreate hooks
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (s *SecurityEvent) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (c *ComplianceLog) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (d *DataRetentionPolicy) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

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